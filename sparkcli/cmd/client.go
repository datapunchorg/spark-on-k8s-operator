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

package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	apigatewayv1 "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apigateway/apis/v1"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Client struct {
	serverUrl  string
	credential UserCredential
}

func NewBasicAuthClient(serverUrl string, user string, password string) *Client {
	return &Client{
		serverUrl: serverUrl,
		credential: UserCredential{
			Name:     user,
			Password: password,
		},
	}
}

func (c *Client) UploadFileToS3(filePath string) (string, error) {
	file, err := os.Open(filePath)
	defer file.Close()

	if err != nil {
		return "", fmt.Errorf("cannot open file %s: %s", filePath, err.Error())
	}
	fileInfo, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("cannot get info for file %s: %s", filePath, err.Error())
	}
	fileSize := fileInfo.Size()

	fileName := filepath.Base(filePath)
	encodedFileName := url.QueryEscape(fileName)
	url := fmt.Sprintf("%s/s3/upload?name=%s", c.serverUrl, encodedFileName)

	req, err := http.NewRequest(http.MethodPost, url, file)
	if err != nil {
		return "", fmt.Errorf("failed to create post request for %s: %s", url, err.Error())
	}
	req.Header.Set("Content-Length", fmt.Sprintf("%d", fileSize))
	req.Header.Set("Content-Type", "application/octet-stream")
	req.SetBasicAuth(c.credential.Name, c.credential.Password)

	log.Printf("Sending file %s to %s", filePath, url)

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to post to %s: %s", url, err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", ErrorBadHttpStatus(url, response)
	}

	responseBytes, err := ReadHttpResponse(response)
	if err != nil {
		return "", fmt.Errorf("failed to read response data from %s: %s", url, err)
	}

	responseStruct := apigatewayv1.UploadFileResponse{}
	err = json.Unmarshal(responseBytes, &responseStruct)
	if err != nil {
		return "", fmt.Errorf("failed to parse response from %s: %s, response: %s", url, err.Error(), string(responseBytes))
	}

	if responseStruct.Url == "" {
		return "", fmt.Errorf("failed to upload file, response: %s", string(responseBytes))
	}

	return responseStruct.Url, nil
}

func (c *Client) SubmitApplication(request apigatewayv1.SparkApplicationSubmissionRequest) (string, error) {
	url := fmt.Sprintf("%s/submissions", c.serverUrl)
	return c.submitApplicationImpl(request, url)
}

func (c *Client) SubmitApplicationWithId(request apigatewayv1.SparkApplicationSubmissionRequest, submissionId string) (string, error) {
	url := fmt.Sprintf("%s/submissions/%s", c.serverUrl, submissionId)
	return c.submitApplicationImpl(request, url)
}

func (c *Client) submitApplicationImpl(request apigatewayv1.SparkApplicationSubmissionRequest, url string) (string, error) {
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return "",
			fmt.Errorf("failed to serialize request to json: %s", err.Error())
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(requestBytes))
	if err != nil {
		return "",
			fmt.Errorf("failed to create post request for %s: %s", url, err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.credential.Name, c.credential.Password)

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return "",
			fmt.Errorf("failed to post %s: %s", url, err.Error())
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", ErrorBadHttpStatus(url, response)
	}

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return "",
			fmt.Errorf("failed to read response data for %s: %s", url, err.Error())
	}

	responseStruct := apigatewayv1.SparkApplicationSubmissionResponse{}
	err = json.Unmarshal(responseBytes, &responseStruct)
	if err != nil {
		return "", fmt.Errorf("failed to parse response from %s: %s, response: %s", url, err.Error(), string(responseBytes))
	}

	if responseStruct.SubmissionId == "" {
		return "", fmt.Errorf("failed to submit application, response: %s", string(responseBytes))
	}

	return responseStruct.SubmissionId, nil
}

func (c *Client) GetApplicationStatus(submissionId string) (string, apigatewayv1.SubmissionStatusResponse, error) {
	result := apigatewayv1.SubmissionStatusResponse{}

	url := fmt.Sprintf("%s/submissions/%s/status", c.serverUrl, submissionId)

	req, err := http.NewRequest(http.MethodGet, url, bytes.NewReader([]byte{}))
	if err != nil {
		return "", result,
			fmt.Errorf("failed to create get request for %s: %s", url, err.Error())
	}
	req.SetBasicAuth(c.credential.Name, c.credential.Password)

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", result,
			fmt.Errorf("failed to get %s: %s", url, err.Error())
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", result, ErrorBadHttpStatus(url, response)
	}

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return "", result,
			fmt.Errorf("failed to read response data for %s: %s", url, err.Error())
	}

	responseStr := string(responseBytes)

	responseStruct := apigatewayv1.SubmissionStatusResponse{}
	err = json.Unmarshal(responseBytes, &responseStruct)
	if err != nil {
		return responseStr, result,
			fmt.Errorf("failed to parse response from %s: %s, response: %s", url, err.Error(), responseStr)
	}

	if responseStruct.SubmissionId == "" {
		return responseStr, result,
			fmt.Errorf("failed to get application status, response: %s", responseStr)
	}

	return responseStr, responseStruct, nil
}

func (c *Client) ListSubmissions(limit int64) (string, apigatewayv1.ListSubmissionsResponse, error) {
	result := apigatewayv1.ListSubmissionsResponse{}

	url := fmt.Sprintf("%s/submissions", c.serverUrl)

	if limit > 0 {
		url += fmt.Sprintf("?limit=%d", limit)
	}

	req, err := http.NewRequest(http.MethodGet, url, bytes.NewReader([]byte{}))
	if err != nil {
		return "", result,
			fmt.Errorf("failed to create get request for %s: %s", url, err.Error())
	}
	req.SetBasicAuth(c.credential.Name, c.credential.Password)

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", result,
			fmt.Errorf("failed to get %s: %s", url, err.Error())
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", result, ErrorBadHttpStatus(url, response)
	}

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return "", result,
			fmt.Errorf("failed to read response data for %s: %s", url, err.Error())
	}

	responseStr := string(responseBytes)

	err = json.Unmarshal(responseBytes, &result)
	if err != nil {
		return responseStr, result,
			fmt.Errorf("failed to parse response from %s: %s, response: %s", url, err.Error(), responseStr)
	}

	return responseStr, result, nil
}

func (c *Client) PrintApplicationLog(submissionId string, executorId int, followLogs bool) {
	url := fmt.Sprintf("%s/submissions/%s/log", c.serverUrl, submissionId)

	queryParameters := make([]string, 0, 5)
	if executorId != -1 {
		queryParameters = append(queryParameters, fmt.Sprintf("executor=%d", executorId))
	}
	if followLogs {
		queryParameters = append(queryParameters, fmt.Sprintf("follow=true"))
	}
	if len(queryParameters) > 0 {
		url += "?" + strings.Join(queryParameters, "&")
	}

	req, err := http.NewRequest(http.MethodGet, url, bytes.NewReader([]byte{}))
	if err != nil {
		log.Printf("Failed to create get request for %s: %s", url, err.Error())
		return
	}
	req.SetBasicAuth(c.credential.Name, c.credential.Password)

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Failed to get %s: %s", url, err.Error())
		return
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.Printf("Failed to get %s: %s", url, ErrorBadHttpStatus(url, response).Error())
		return
	}

	reader := bufio.NewReader(response.Body)
	for {
		// TODO optimize reading here by reading in batches
		if c, _, err := reader.ReadRune(); err != nil {
			if err == io.EOF {
				break
			} else {
				log.Printf("Failed to get log: %s", err.Error())
				return
			}
		} else {
			fmt.Printf("%s", string(c))
		}
	}
}

func (c *Client) DeleteApplication(submissionId string) (string, apigatewayv1.DeleteSubmissionResponse, error) {
	result := apigatewayv1.DeleteSubmissionResponse{}

	url := fmt.Sprintf("%s/submissions/%s", c.serverUrl, submissionId)

	req, err := http.NewRequest(http.MethodDelete, url, bytes.NewReader([]byte{}))
	if err != nil {
		return "", result,
			fmt.Errorf("failed to create delete request for %s: %s", url, err.Error())
	}
	req.SetBasicAuth(c.credential.Name, c.credential.Password)

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", result,
			fmt.Errorf("failed to get %s: %s", url, err.Error())
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", result, ErrorBadHttpStatus(url, response)
	}

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return "", result,
			fmt.Errorf("failed to read response data for %s: %s", url, err.Error())
	}

	responseStr := string(responseBytes)

	responseStruct := apigatewayv1.DeleteSubmissionResponse{}
	err = json.Unmarshal(responseBytes, &responseStruct)
	if err != nil {
		return responseStr, result,
			fmt.Errorf("failed to parse response from %s: %s, response: %s", url, err.Error(), responseStr)
	}

	return responseStr, responseStruct, nil
}
