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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"strconv"
)

func ListSubmissions(c *gin.Context, config *ApiConfig) {
	limit := int64(100)
	limitQueryValue, ok := c.GetQuery("limit")
	if ok && limitQueryValue != "" {
		parsedInt, err := strconv.ParseInt(limitQueryValue, 10, 64)
		if err != nil {
			msg := fmt.Sprintf("Invalid limit query parameter value %s: %s", limitQueryValue, err.Error())
			writeErrorResponse(c, http.StatusBadRequest, msg, nil)
			return
		}
		limit = parsedInt
	}

	crdClient, err := createSparkApplicationClient()
	if err != nil {
		writeErrorResponse(c, http.StatusInternalServerError, "", err)
		return
	}

	appList, err := crdClient.SparkoperatorV1beta2().SparkApplications(config.SparkApplicationNamespace).List(context.TODO(), metav1.ListOptions{
		Limit: limit,
	})
	if err != nil {
		msg := fmt.Sprintf("Failed to get SparkApplication: %s", err.Error())
		writeErrorResponse(c, http.StatusInternalServerError, msg, nil)
		return
	}

	summaryList := make([]v1.SparkApplicationSubmissionSummary, 0, limit)
	for _, app := range appList.Items {
		state := v1beta2.UnknownState
		if app.Status.AppState.State != "" {
			state = app.Status.AppState.State
		}

		id := app.Name
		appId := app.Status.SparkApplicationID

		summary := v1.SparkApplicationSubmissionSummary{
			SubmissionId: id,
			State:        string(state),
			RecentAppId:  appId,
		}
		summaryList = append(summaryList, summary)
	}

	response := v1.ListSubmissionsResponse{
		Items: summaryList,
	}

	c.IndentedJSON(http.StatusOK, response)
}
