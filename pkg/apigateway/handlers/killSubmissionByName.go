package handlers

import (
	"encoding/json"
	"fmt"
	v1 "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apigateway/apis/v1"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
)

func KillSubmissionByName(c *gin.Context, config *ApiConfig) {
	var request v1.KillSubmissionByNameRequest

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

	writeErrorResponse(c, http.StatusInternalServerError, "This endpoint is not implemented", nil)
}

