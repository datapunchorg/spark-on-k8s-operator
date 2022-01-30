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
	"github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apigateway/apis/v1"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/golang/glog"
	"github.com/google/uuid"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gin-gonic/gin"
)

func UploadFile(c *gin.Context, s3Region string, s3Bucket string, s3Root string) {
	nameQueryParamName := "name"
	name, ok := c.GetQuery(nameQueryParamName)
	if !ok || name == "" {
		msg := fmt.Sprintf("Missing query string parameter: %s", nameQueryParamName)
		writeErrorResponse(c,http.StatusBadRequest, msg, nil)
		return
	}

	nameHash := fmt.Sprintf("%s%s", name[0:1], name[len(name)-1:])
	key := fmt.Sprintf("%s/%s/%s/%s", s3Root, nameHash, uuid.New().String(), name)

	session, err := session.NewSession(
		&aws.Config{
			Region: aws.String(s3Region),
		})
	if err != nil {
		msg := fmt.Sprintf("Failed to create AWS session in region %s: %s", s3Region, err.Error())
		writeErrorResponse(c,http.StatusInternalServerError, msg, nil)
		return
	}

	glog.Infof("Uploading to s3, bucket: %s, key: %s", s3Bucket, key)

	uploader := s3manager.NewUploader(session)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s3Bucket),
		Key: aws.String(key),
		Body: c.Request.Body,
	})

	if err != nil {
		msg := fmt.Sprintf("Failed to upload to s3, bucket: %s, key: %s, error: %s", s3Bucket, key, err.Error())
		writeErrorResponse(c,http.StatusInternalServerError, msg, nil)
		return
	}

	glog.Infof("Uploaded to s3, bucket: %s, key: %s", s3Bucket, key)

	result := v1.UploadFileResponse{
		Url: "s3a://" + s3Bucket + "/" + key,
	}
	c.IndentedJSON(http.StatusOK, result)
}
