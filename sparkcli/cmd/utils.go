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
	"archive/zip"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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

func ZipDirAndSaveInDir(sourceDir string, targetDir string) (string, error) {
	zipBasePath := ""
	suffix := "/..."
	if strings.HasSuffix(sourceDir, suffix) {
		zipBasePath = ""
		sourceDir = sourceDir[0 : len(sourceDir)-len(suffix)]
	} else {
		zipBasePath = filepath.Base(sourceDir)
	}

	err := os.MkdirAll(targetDir, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("failed to create or check target dir %s: %s", targetDir, err.Error())
	}

	dirName := filepath.Base(sourceDir)
	zipFileName := dirName + ".zip"
	zipFilePath := filepath.Join(targetDir, zipFileName)
	outFile, err := os.Create(zipFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to create zip file %s: %s", zipFilePath, err.Error())
	}
	defer outFile.Close()

	writer := zip.NewWriter(outFile)
	err = addZipFiles(writer, sourceDir, zipBasePath)
	if err != nil {
		return "", fmt.Errorf("failed to read dir %s and add to zip file %s: %s", sourceDir, zipFilePath, err.Error())
	}
	err = writer.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close zip file %s: %s", zipFilePath, err.Error())
	}
	return zipFilePath, nil
}

func addZipFiles(writer *zip.Writer, fileBasePath string, zipBasePath string) error {
	if zipBasePath != "" {
		if !strings.HasSuffix(zipBasePath, string(os.PathSeparator)) {
			zipBasePath += string(os.PathSeparator)
		}
		_, err := writer.Create(zipBasePath)
		if err != nil {
			return fmt.Errorf("failed to create zip dir entry %s: %s", zipBasePath, err.Error())
		}
	}
	files, err := ioutil.ReadDir(fileBasePath)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %s", fileBasePath, err.Error())
	}

	if zipBasePath == "" && len(files) == 0 {
		return fmt.Errorf("cannot write zip file, dir %s is empty and zip base path is empty", fileBasePath)
	}

	for _, file := range files {
		filePath := filepath.Join(fileBasePath, file.Name())
		zipPath := filepath.Join(zipBasePath, file.Name())
		if !file.IsDir() {
			data, err := ioutil.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %s", filePath, err.Error())
			}
			f, err := writer.Create(zipPath)
			if err != nil {
				return fmt.Errorf("failed to create zip file entry %s: %s", filePath, err.Error())
			}
			_, err = f.Write(data)
			if err != nil {
				return fmt.Errorf("failed to write zip file entry %s: %s", filePath, err.Error())
			}
		} else if file.IsDir() {
			newFileBase := filePath + "/"
			newZipPath := zipPath + "/"
			err = addZipFiles(writer, newFileBase, newZipPath)
			if err != nil {
				return fmt.Errorf("failed to add zip files %s into zip %s: %s", newFileBase, newZipPath, err.Error())
			}
		}
	}
	return nil
}
