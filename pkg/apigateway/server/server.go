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

package server

import (
	"fmt"
	"github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apigateway/handlers"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/util/yaml"
	"os"
)

func Run(config Config) {
	port := config.Port

	urlPrefix := config.UrlPrefix
	if urlPrefix == "" {
		urlPrefix = DefaultUrlPrefix
	}

	s3Region := config.S3Region
	if s3Region == "" {
		s3Region = DefaultS3Region
	}

	s3Bucket := config.S3Bucket
	if s3Bucket == "" {
		glog.Fatal("S3 bucket is not specified")
	}

	s3Root := config.S3Root

	extraConfig := ExtraConfig{}
	if config.ConfigFile != "" {
		file, err := os.Open(config.ConfigFile)
		if err != nil {
			glog.Warningf("Failed to open config file %s: %s", config.ConfigFile, err.Error())
		} else {
			defer file.Close()
			decoder := yaml.NewYAMLOrJSONDecoder(file, 200)
			err = decoder.Decode(&extraConfig)
			if err != nil {
				glog.Warningf("Failed to parse config file %s: %s", config.ConfigFile, err.Error())
			} else {
				glog.Infof("Loaded config from %s", config.ConfigFile)
			}
		}
	}

	glog.Infof("Running API gateway for Spark application namespace %s", config.SparkApplicationNamespace)

	router := gin.Default()

	apiConfig := handlers.ApiConfig{
		SparkApplicationNamespace: config.SparkApplicationNamespace,
		SparkUIServiceUrlFormat:   config.SparkUIServiceUrlFormat,
		SparkUIBaseProxyPrefix:    fmt.Sprintf("/%s/v1/sparkui", config.UrlPrefix),
		S3Region:                  s3Region,
		S3Bucket:                  s3Bucket,
		S3Root:                    s3Root,
		SubmissionConfig:          extraConfig.SubmissionConfig,
	}

	router.GET("/", handlers.HealthCheck)
	router.GET("/index.html", handlers.HealthCheck)
	router.GET("/index.htm", handlers.HealthCheck)
	router.GET("/health", handlers.HealthCheck)
	router.GET("/healthcheck", handlers.HealthCheck)

	userName := config.UserName

	apiRootPath := fmt.Sprintf("%s/v1", urlPrefix)

	var group *gin.RouterGroup
	if userName == "" {
		glog.Infof("API gateway will be accessible without user name")
		group = router.Group(apiRootPath, func(context *gin.Context) {
		})
	} else {
		glog.Infof("API gateway will be accessible with required user name %s and matching password", userName)
		userPassword := config.UserPassword
		if userPassword == "" {
			userPassword = uuid.New().String()
			glog.Infof(
				"******************************\n"+
					"API gateway is set to require %s as user name, but empty value as user password. Generated value %s as required password.\n"+
					"******************************",
				userName, userPassword)
		}
		group = router.Group(apiRootPath, gin.BasicAuth(gin.Accounts{
			userName: userPassword,
		}))
	}

	group.POST("/s3/upload",
		func(context *gin.Context) {
			handlers.UploadFile(context, s3Region, s3Bucket, s3Root)
		})

	group.POST("/submissions",
		func(context *gin.Context) {
			handlers.PostSubmission(context, &apiConfig)
		})

	group.GET("/submissions/:id/status",
		func(context *gin.Context) {
			handlers.GetSubmissionStatus(context, &apiConfig)
		})

	group.GET("/submissions/:id/log",
		func(context *gin.Context) {
			handlers.GetLog(context, &apiConfig)
		})

	group.DELETE("/submissions/:id",
		func(context *gin.Context) {
			handlers.DeleteSubmission(context, &apiConfig)
		})

	group.GET("/sparkui/*path",
		func(context *gin.Context) {
			handlers.ServeSparkUI(context, &apiConfig, apiRootPath + "/sparkui")
		})

	group.GET("/health", handlers.HealthCheck)

	router.Run(fmt.Sprintf(":%d", port))
}
