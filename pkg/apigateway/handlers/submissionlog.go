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
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"io"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"strconv"
)

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
