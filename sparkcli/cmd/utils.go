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
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"
)

func CheckLocalFile(file string) (string, error) {
	fileUrl, err := url.Parse(file)
	if err != nil {
		return "", err
	}

	if strings.EqualFold(fileUrl.Scheme, "file") {
		return fileUrl.Path, nil
	} else if fileUrl.Scheme == "" {
		return file, nil
	}

	return "", fmt.Errorf("not local file: %s", file)
}

func ReadHttpResponse(response *http.Response) ([]byte, error) {
	defer response.Body.Close()

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return responseBytes, nil
}

func InsecureTLS(insecureSkipVerify bool) {
	if insecureSkipVerify {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: insecureSkipVerify}
		log.Printf("Skip SSL certificate verification")
	}
}

func ErrorBadHttpStatus(url string, response *http.Response) error {
	bytes, err := ioutil.ReadAll(response.Body)
	str := ""
	if err != nil {
		str = err.Error()
	} else {
		str = string(bytes)
	}
	return fmt.Errorf("got bad response status %s from %s, response body: %s", response.Status, url, str)
}

func WriteOutputFileExitOnError(filePath string, fileContent string) {
	f, err := os.Create(filePath)
	if err != nil {
		ExitWithError(fmt.Sprintf("Failed to create output file %s: %s", filePath, err.Error()))
	}
	defer f.Close()
	_, err = f.WriteString(fileContent)
	if err != nil {
		ExitWithError(fmt.Sprintf("Failed to write output file %s: %s", filePath, err.Error()))
	}
}

func ExitWithError(str string) {
	fmt.Fprintln(os.Stderr, str)
	os.Exit(1)
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

func GetObjectTypeName(obj interface{}) string {
	t := reflect.TypeOf(obj)
	if t == nil {
		return "nil"
	}
	if t.Kind() == reflect.Ptr {
		return "*" + t.Elem().Name()
	} else {
		return t.Name()
	}
}
