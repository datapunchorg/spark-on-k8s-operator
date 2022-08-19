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
	"fmt"
	v1 "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apigateway/apis/v1"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"strings"
	"time"
)

const AppNameLabel = "appName"

func writeErrorResponse(context *gin.Context, httpCode int, message string, err error) {
	var str string
	if message == "" && err == nil {
		str = "Unknown Error"
	} else if message != "" && err == nil {
		str = message
	} else if message == "" && err != nil {
		str = err.Error()
	} else {
		str = fmt.Sprintf("%s: %s", message, err.Error())
	}
	glog.Warning(str)
	context.IndentedJSON(httpCode,
		v1.Response{
			Message: str,
		})
}

func addMissingKeysInStringMap(to map[string]string, from map[string]string) {
	for k, v := range from {
		if _, ok := to[k]; !ok {
			to[k] = v
		}
	}
}

func findSparkImageName(sparkImageConfigs []SparkImageConfig, sparkVersion string, sparkType string) (string, bool){
	for _, entry := range sparkImageConfigs {
		if strings.EqualFold(entry.Version, sparkVersion) && strings.EqualFold(entry.Type, sparkType) {
			return entry.Image, true
		}
	}
	return "", false
}

func trimStatusUrlFromSubmissions(url string) (string, error) {
	urlParts := strings.Split(url, "/")
	if !strings.EqualFold(urlParts[len(urlParts) - 1], "status") {
		return "", fmt.Errorf("invalid status url (not ending with status): %s", url)
	}
	if len(urlParts) <= 3 {
		return "", fmt.Errorf("invalid status url (not ending with /submissions/id/status): %s", url)
	}
	return strings.Join(urlParts[0:len(urlParts)-3], "/"), nil
}

func RetryUntilTrue(run func() (bool, error), maxWait time.Duration, retryInterval time.Duration) error {
	currentTime := time.Now()
	startTime := currentTime
	endTime := currentTime.Add(maxWait)
	for !currentTime.After(endTime) {
		result, err := run()
		if err != nil {
			return err
		}
		if result {
			return nil
		}
		if !currentTime.After(endTime) {
			time.Sleep(retryInterval)
		}
		currentTime = time.Now()
	}
	return fmt.Errorf("timed out after running %d seconds while max wait time is %d seconds", int(currentTime.Sub(startTime).Seconds()), int(maxWait.Seconds()))
}
