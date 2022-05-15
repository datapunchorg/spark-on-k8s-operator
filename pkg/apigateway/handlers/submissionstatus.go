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
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
)

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
