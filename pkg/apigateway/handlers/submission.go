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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
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
