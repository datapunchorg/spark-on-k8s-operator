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

package v1

import "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apis/sparkoperator.k8s.io/v1beta2"

type SparkApplicationSubmissionRequest struct {
	v1beta2.SparkApplicationSpec `json:",inline"`
	ApplicationDescription       string `json:"applicationDescription"`
}

type SparkApplicationSubmissionResponse struct {
	SubmissionId string `json:"submissionId"`
}

type UploadFileResponse struct {
	Url string `json:"url"`
}

type SubmissionStatusResponse struct {
	SubmissionId string `json:"submissionId"`
	State        string `json:"state"`
	RecentAppId  string `json:"recentAppId"`
}

type DeleteSubmissionResponse struct {
	SubmissionId string `json:"submissionId"`
	Message      string `json:"message"`
}

type Response struct {
	Message string `json:"message"`
}

type GetSparkApplicationsResponse struct {
	Items []SparkApplicationSummary `json:"items"`
}

type SparkApplicationSummary struct {
	SubmissionId string `json:"submissionId"`
	State        string `json:"state"`
	RecentAppId  string `json:"recentAppId"`
}
