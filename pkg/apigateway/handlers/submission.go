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
	"encoding/json"
	"fmt"
	"github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apigateway/apis/v1"
	"github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apis/sparkoperator.k8s.io/v1beta2"
	crdclientset "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/client/clientset/versioned"
	"github.com/golang/glog"
	"github.com/google/uuid"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func PostNewSubmission(c *gin.Context, config *ApiConfig) {
	submissionIdPrefix := "app-"
	submissionId := submissionIdPrefix + strings.ReplaceAll(uuid.New().String(), "-", "")

	CreateSubmission(c, config, submissionId)
}

func PostSubmissionWithId(c *gin.Context, config *ApiConfig) {
	submissionId := c.Param("id")

	overwrite := false
	overwriteQueryParamName := "overwrite"
	if queryValue, ok := c.GetQuery(overwriteQueryParamName); ok {
		if queryValue != "" {
			var err error
			overwrite, err = strconv.ParseBool(queryValue)
			if err != nil {
				msg := fmt.Sprintf("Invalid query parameter value for %s: %s. It should be true or false.", overwriteQueryParamName, queryValue)
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
		writeErrorResponse(c, http.StatusInternalServerError, "", err)
		return
	}

	glog.Infof("Checking whether SparkApplication %s already exists in namespace %s", submissionId, config.SparkApplicationNamespace)
	existingApp, getAppErr := crdClient.SparkoperatorV1beta2().SparkApplications(config.SparkApplicationNamespace).Get(context.TODO(), submissionId, metav1.GetOptions{})

	if overwrite {
		glog.Infof("Trying to delete SparkApplication %s in namespace %s due to overwrite mode", submissionId, config.SparkApplicationNamespace)
		err := crdClient.SparkoperatorV1beta2().SparkApplications(config.SparkApplicationNamespace).Delete(context.TODO(), submissionId, metav1.DeleteOptions{})
		if err != nil {
			glog.Infof("Failed to delete SparkApplication %s in namespace %s, ignore and continue creating new SparkApplication: %s", submissionId, config.SparkApplicationNamespace, err.Error())
		} else {
			glog.Infof("Deleted existing SparkApplication %s in namespace %s", submissionId, config.SparkApplicationNamespace)
		}
		if getAppErr == nil && existingApp != nil {
			driverPodName := existingApp.Status.DriverInfo.PodName
			if driverPodName != "" {
				err := kubeClient.CoreV1().Pods(existingApp.Namespace).Delete(context.TODO(), driverPodName, metav1.DeleteOptions{})
				if err != nil {
					glog.Infof("Failed to delete existing SparkApplication driver pod %s in namespace %s, ignore and continue creating new SparkApplication: %s", driverPodName, existingApp.Namespace, err.Error())
				} else {
					glog.Infof("Deleted existing SparkApplication driver pod %s in namespace %s", driverPodName, existingApp.Namespace)
				}
			}
			for executorPodName := range existingApp.Status.ExecutorState {
				err := kubeClient.CoreV1().Pods(existingApp.Namespace).Delete(context.TODO(), executorPodName, metav1.DeleteOptions{})
				if err != nil {
					glog.Infof("Failed to delete existing SparkApplication executor pod %s in namespace %s, ignore and continue creating new SparkApplication: %s", executorPodName, existingApp.Namespace, err.Error())
				} else {
					glog.Infof("Deleted existing SparkApplication executor pod %s in namespace %s", driverPodName, existingApp.Namespace)
				}
			}
			sparkUIServiceName := GetDefaultUIServiceName(existingApp)
			err = kubeClient.CoreV1().Services(existingApp.Namespace).Delete(context.TODO(), sparkUIServiceName, metav1.DeleteOptions{})
			if err != nil {
				glog.Infof("Failed to delete SparkApplication UI service %s in namespace %s, ignore and continue creating new SparkApplication: %s", sparkUIServiceName, existingApp.Namespace, err.Error())
			} else {
				glog.Infof("Deleted existing SparkApplication UI service %s in namespace %s", sparkUIServiceName, existingApp.Namespace)
			}
		}
	} else {
		if getAppErr == nil {
			msg := fmt.Sprintf("Cannot create SparkApplication %s since it already exists in namespace %s (created at %s)", submissionId, config.SparkApplicationNamespace, existingApp.CreationTimestamp)
			writeErrorResponse(c, http.StatusBadRequest, msg, nil)
			return
		}
	}

	CreateSubmission(c, config, submissionId)
}

func CreateSubmission(c *gin.Context, config *ApiConfig, submissionId string) {
	var request v1.SparkApplicationSubmissionRequest

	bytes, err := ioutil.ReadAll(c.Request.Body)
	c.Request.Body.Close()
	if err != nil {
		msg := fmt.Sprintf("Failed to read request body: %s", err.Error())
		writeErrorResponse(c, http.StatusInternalServerError, msg, nil)
		return
	}

	err = json.Unmarshal(bytes, &request)
	if err != nil {
		msg := fmt.Sprintf("Bad json request: %s", err.Error())
		writeErrorResponse(c, http.StatusBadRequest, msg, nil)
		return
	}

	//if err := c.BindJSON(&request); err != nil {
	//	msg := fmt.Sprintf("Bad json request: %s", err.Error())
	//	writeErrorResponse(c, http.StatusBadRequest, msg, nil)
	//	return
	//}

	crdClient, err := createSparkApplicationClient()
	if err != nil {
		writeErrorResponse(c, http.StatusInternalServerError, "", err)
		return
	}

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
	sparkConf["spark.ui.port"] = "4040"

	sparkConf["spark.ui.proxyBase"] = fmt.Sprintf("%s/%s", config.SparkUIBaseProxyPrefix, submissionId)
	if !config.SparkUIModifyRedirectUrl {
		sparkConf["spark.ui.proxyRedirectUri"] = "/"
	}

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
