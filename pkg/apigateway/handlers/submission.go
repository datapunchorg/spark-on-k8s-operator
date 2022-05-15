/*
Copyright spark-on-k8s-operator contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package handlers

import (
	"context"
	"fmt"
	"github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apigateway/apis/v1"
	"github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apis/sparkoperator.k8s.io/v1beta2"
	crdclientset "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/client/clientset/versioned"
	"github.com/golang/glog"
	"github.com/google/uuid"
	"io"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func PostSubmission(c *gin.Context, config *ApiConfig) {
	var request v1.SparkApplicationSubmissionRequest

	if err := c.BindJSON(&request); err != nil {
		msg := fmt.Sprintf("Bad json request: %s", err.Error())
		writeErrorResponse(c, http.StatusBadRequest, msg, nil)
		return
	}

	crdClient, err := createSparkApplicationClient()
	if err != nil {
		writeErrorResponse(c, http.StatusInternalServerError, "", err)
		return
	}

	submissionIdPrefix := "app-"
	submissionId := submissionIdPrefix + strings.ReplaceAll(uuid.New().String(), "-", "")

	app := v1beta2.SparkApplication{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "sparkoperator.k8s.io/v1beta2",
			Kind:       "SparkApplication",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: config.SparkApplicationNamespace,
			Name:      submissionId,
		},
		Spec: request.SparkApplicationSpec,
	}

	serviceAccount := config.SubmissionConfig.ServiceAccount
	if serviceAccount == "" {
		msg := fmt.Sprintf("Cannot submis Spark application due to empty service account in server side config")
		writeErrorResponse(c, http.StatusInternalServerError, msg, nil)
		return
	}

	app.Spec.Driver.ServiceAccount = &serviceAccount
	app.Spec.Executor.ServiceAccount = &serviceAccount

	sparkVersion := request.SparkVersion
	if sparkVersion == "" {
		sparkVersion = config.SubmissionConfig.DefaultSparkVersion
		glog.Infof("Add Spark version %s for submission %s", sparkVersion, submissionId)
		app.Spec.SparkVersion = sparkVersion
	}

	if app.Spec.SparkVersion == "" {
		msg := fmt.Sprintf("Cannot submis Spark application due to empty Spark version")
		writeErrorResponse(c, http.StatusBadRequest, msg, nil)
		return
	}

	sparkType := request.Type
	if sparkType == "" {
		if request.MainClass != nil && *request.MainClass != "" {
			sparkType = v1beta2.JavaApplicationType
		} else {
			sparkType = v1beta2.PythonApplicationType
		}
		glog.Infof("Add type %s for submission %s", sparkType, submissionId)
		app.Spec.Type = sparkType
	}

	if app.Spec.Type == "" {
		msg := fmt.Sprintf("Cannot submis Spark application due to empty Spark application type")
		writeErrorResponse(c, http.StatusBadRequest, msg, nil)
		return
	}

	sparkImage := request.Image
	if sparkImage == nil || *sparkImage == "" {
		foundSparkImage, ok := findSparkImageName(config.SubmissionConfig.SparkImages, sparkVersion, string(sparkType))
		if ok {
			glog.Infof("Add image %s for submission %s", foundSparkImage, submissionId)
			app.Spec.Image = &foundSparkImage
		} else {
			msg := fmt.Sprintf("Cannot find Spark image for Spark %s and %s", sparkVersion, sparkType)
			writeErrorResponse(c, http.StatusBadRequest, msg, nil)
			return
		}
	}

	if app.Spec.Image == nil || *app.Spec.Image == "" {
		msg := fmt.Sprintf("Cannot submis Spark application due to empty Spark image")
		writeErrorResponse(c, http.StatusBadRequest, msg, nil)
		return
	}

	sparkConf := app.Spec.SparkConf
	if sparkConf == nil {
		sparkConf = map[string]string{}
	}

	addMissingKeysInStringMap(sparkConf, config.SubmissionConfig.SparkConf)

	sparkConf["spark.kubernetes.executor.podNamePrefix"] = submissionId
	sparkConf["spark.ui.proxyBase"] = fmt.Sprintf("%s/%s", config.SparkUIBaseProxyPrefix, submissionId)

	app.Spec.SparkConf = sparkConf

	err = createSparkApplication(config.SparkApplicationNamespace, &app, crdClient)
	if err != nil {
		msg := fmt.Sprintf("Failed to create SparkApplication: %s", err.Error())
		writeErrorResponse(c, http.StatusInternalServerError, msg, nil)
		return
	}

	response := v1.SparkApplicationSubmissionResponse{
		SubmissionId: submissionId,
	}

	c.IndentedJSON(http.StatusOK, response)
}

func GetSubmissionStatus(c *gin.Context, config *ApiConfig) {
	id := c.Param("id")

	crdClient, err := createSparkApplicationClient()
	if err != nil {
		writeErrorResponse(c, http.StatusInternalServerError, "", err)
		return
	}

	app, err := crdClient.SparkoperatorV1beta2().SparkApplications(config.SparkApplicationNamespace).Get(context.TODO(), id, metav1.GetOptions{})
	if err != nil {
		msg := fmt.Sprintf("Failed to get SparkApplication %s: %s", id, err.Error())
		writeErrorResponse(c, http.StatusInternalServerError, msg, nil)
		return
	}

	state := v1beta2.UnknownState
	if app.Status.AppState.State != "" {
		state = app.Status.AppState.State
	}

	applicationMessage := app.Status.AppState.ErrorMessage

	appId := app.Status.SparkApplicationID

	response := v1.SubmissionStatusResponse{
		SubmissionId: id,
		State:        string(state),
		ApplicationMessage: applicationMessage,
		RecentAppId:  appId,
	}

	// TODO return spark history UI url for completed application
	if response.State == "RUNNING" {
		// Example urls for status and spark ui:
		// http://a5aa5b8a98ebc4c5bb247b2ed762c2fe-1896699602.us-west-1.elb.amazonaws.com/sparkapi/v1/submissions/app-8df31ec2a9a44a269b78a859ee05ecaf/status
		// http://a5aa5b8a98ebc4c5bb247b2ed762c2fe-1896699602.us-west-1.elb.amazonaws.com/sparkapi/v1/sparkui/app-8df31ec2a9a44a269b78a859ee05ecaf
		requestUrl := c.Request.URL
		path, err := trimStatusUrlFromSubmissions(requestUrl.Path)
		if err != nil {
			glog.Warningf("Got invalid status endpoint url %s in request: %s", requestUrl, err.Error())
		} else {
			requestUrl.Path = fmt.Sprintf("%s/sparkui/%s", path, id)
			response.SparkUI = requestUrl.String()
		}
	}

	c.IndentedJSON(http.StatusOK, response)
}

func GetLog(c *gin.Context, config *ApiConfig) {
	id := c.Param("id")
	glog.Infof("Getting log for Spark application %s", id)

	follow := false

	followQueryParamName := "follow"
	if queryValue, ok := c.GetQuery(followQueryParamName); ok {
		if queryValue != "" {
			var err error
			follow, err = strconv.ParseBool(queryValue)
			if err != nil {
				msg := fmt.Sprintf("Invalid query parameter value for %s: %s. It should be true or false.", followQueryParamName, queryValue)
				writeErrorResponse(c, http.StatusBadRequest, msg, nil)
				return
			}
		}
	}

	executorId := -1

	executorQueryParamName := "executor"
	if queryValue, ok := c.GetQuery(executorQueryParamName); ok {
		if queryValue != "" {
			var err error
			executorId, err = strconv.Atoi(queryValue)
			if err != nil {
				msg := fmt.Sprintf("Invalid query parameter value for %s: %s. It should be a number.", executorQueryParamName, queryValue)
				writeErrorResponse(c, http.StatusBadRequest, msg, nil)
				return
			}
		}
	}

	kubeConfig := getKubeConfigPath()

	kubeClient, err := getKubeClient(kubeConfig)
	if err != nil {
		msg := fmt.Sprintf("Failed to get Kubernetes client: %s", err.Error())
		writeErrorResponse(c, http.StatusInternalServerError, msg, nil)
		return
	}

	crdClient, err := getSparkApplicationClient(kubeConfig)
	if err != nil {
		msg := fmt.Sprintf("Failed to get SparkApplication client: %s", err.Error())
		writeErrorResponse(c, http.StatusInternalServerError, msg, nil)
		return
	}

	app, err := crdClient.SparkoperatorV1beta2().SparkApplications(config.SparkApplicationNamespace).Get(context.TODO(), id, metav1.GetOptions{})
	if err != nil {
		msg := fmt.Sprintf("Failed to get SparkApplication %s: %s", id, err.Error())
		writeErrorResponse(c, http.StatusOK, msg, nil)
		return
	}

	podName := ""
	if executorId <= -1 {
		podName = app.Status.DriverInfo.PodName
	} else {
		podName = fmt.Sprintf("%s-exec-%d", id, executorId)
	}
	if podName == "" {
		msg := fmt.Sprintf("Unable to fetch log as the name of the pod for SparkApplication %s is empty", id)
		writeErrorResponse(c, http.StatusOK, msg, nil)
		return
	}

	glog.Infof("Getting log for Spark application %s pod name: %s", id, podName)

	request := kubeClient.CoreV1().Pods(config.SparkApplicationNamespace).GetLogs(podName, &apiv1.PodLogOptions{Follow: follow})
	reader, err := request.Stream(context.TODO())
	if err != nil {
		msg := fmt.Sprintf("Failed to get log for %s: %v", podName, err.Error())
		writeErrorResponse(c, http.StatusOK, msg, nil)
		return
	}
	if reader == nil {
		msg := fmt.Sprintf("Failed to get log for %s: pod not found", podName)
		writeErrorResponse(c, http.StatusOK, msg, nil)
		return
	}

	glog.Infof("Streaming log for Spark application %s, pod: %s", id, podName)
	clientDisconnected := c.Stream(func(w io.Writer) bool {
		glog.Infof("Sending log for Spark application %s, pod: %s", id, podName)
		str := "Getting log for " + podName + "\n"
		w.Write([]byte(str))
		if _, err := io.Copy(w, reader); err != nil {
			str := fmt.Sprintf("\nFailed to get log for %s: %s", podName, err.Error())
			w.Write([]byte(str))
		}
		w.(http.Flusher).Flush()
		reader.Close()
		glog.Infof("Finished sending log for Spark application %s, pod: %s", id, podName)
		return false
	})
	if clientDisconnected {
		reader.Close()
	}
}

func DeleteSubmission(c *gin.Context, config *ApiConfig) {
	id := c.Param("id")

	crdClient, err := createSparkApplicationClient()
	if err != nil {
		writeErrorResponse(c, http.StatusInternalServerError, "", err)
		return
	}

	err = crdClient.SparkoperatorV1beta2().SparkApplications(config.SparkApplicationNamespace).Delete(context.TODO(), id, metav1.DeleteOptions{})
	if err != nil {
		msg := fmt.Sprintf("Failed to delete SparkApplication %s: %s", id, err.Error())
		writeErrorResponse(c, http.StatusInternalServerError, msg, nil)
		return
	}

	response := v1.DeleteSubmissionResponse{
		SubmissionId: id,
		Message:      "Application deleted",
	}

	c.IndentedJSON(http.StatusOK, response)
}

func createSparkApplication(namespace string, app *v1beta2.SparkApplication, crdClient crdclientset.Interface) error {
	v1beta2.SetSparkApplicationDefaults(app)
	if err := validateSpec(app.Spec); err != nil {
		return err
	}

	if _, err := crdClient.SparkoperatorV1beta2().SparkApplications(namespace).Create(
		context.TODO(),
		app,
		metav1.CreateOptions{},
	); err != nil {
		return err
	}

	glog.Infof("Created SparkApplication %s in namespace %s", app.Name, namespace)

	return nil
}

func validateSpec(spec v1beta2.SparkApplicationSpec) error {
	if spec.Image == nil && (spec.Driver.Image == nil || spec.Executor.Image == nil) {
		return fmt.Errorf("spark image is not set, please specify image field in the root or inside driver/executor")
	}

	return nil
}
