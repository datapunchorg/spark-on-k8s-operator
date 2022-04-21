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

import "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apigateway/handlers"

const (
	DefaultS3Region  = "us-west-1"
	DefaultUrlPrefix = "/sparkapi"
	DefaultS3Root    = "api-gateway-root"
)

type Config struct {
	Port                      int
	UrlPrefix                 string
	UserName                  string
	UserPassword              string
	SparkApplicationNamespace string
	S3Region                  string
	S3Bucket                  string
	S3Root                    string
	ConfigFile                string
}

// ExtraConfig contains content for Config.ConfigFile
type ExtraConfig struct {
	SubmissionConfig   handlers.ApplicationSubmissionConfig `json:"submissionConfig"`
}