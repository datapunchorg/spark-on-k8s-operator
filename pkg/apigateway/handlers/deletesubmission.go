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
	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
)

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
