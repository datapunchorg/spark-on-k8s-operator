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
)

func writeErrorResponse(context *gin.Context, httpCode int, message string, err error)  {
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
