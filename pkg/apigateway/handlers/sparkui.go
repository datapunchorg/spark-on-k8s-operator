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
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func ServeSparkUI(c *gin.Context, config *ApiConfig) {
	id := c.Param("id")

	proxy, err := newReverseProxy(id, config.SparkApplicationNamespace)
	if err != nil {
		msg := fmt.Sprintf("Failed to create reverse proxy for %s: %s", id, err.Error())
		writeErrorResponse(c, http.StatusInternalServerError, msg, nil)
		return
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

func newReverseProxy(submissionId string, appNamespace string) (*httputil.ReverseProxy, error) {
	urlStr := fmt.Sprintf("http://%s-ui-svc.%s.svc.cluster.local:4040/", submissionId, appNamespace)
	url, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse target Spark UI url %s: %s", urlStr, err.Error())
	}
	director := func(req *http.Request) {
		log.Printf("Reverse proxy, serving url %s for requested url %s", url, req.URL)
		req.URL = url
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
	}
	return &httputil.ReverseProxy{Director: director}, nil
}