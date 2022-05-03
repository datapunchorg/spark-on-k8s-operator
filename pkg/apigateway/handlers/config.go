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

type ApiConfig struct {
	SparkApplicationNamespace string `json:"sparkApplicationNamespace"`
	S3Region                  string `json:"s3Region"`
	S3Bucket                  string `json:"s3Bucket"`
	S3Root                    string `json:"s3Root"`
	SubmissionConfig          ApplicationSubmissionConfig `json:"submissionConfig"`
}

type ApplicationSubmissionConfig struct {
	SparkImages []SparkImageConfig `json:"sparkImages,omitempty"`
	DefaultSparkVersion string `json:"defaultSparkVersion,omitempty"`
	SparkConf map[string]string `json:"sparkConf,omitempty"`
}

type SparkImageConfig struct {
	Version string `json:"version,omitempty"`
	Type string `json:"type,omitempty"`
	Image string `json:"image,omitempty"`
}
