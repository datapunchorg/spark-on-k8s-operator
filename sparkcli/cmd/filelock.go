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
	"errors"
	"fmt"
	"os"
	"time"
)

const lockFileSuffix = ".lock"

func LockFile(filePath string) error {
	lockFilePath := filePath + lockFileSuffix
	lockFileNotExists := false
	if _, err := os.Stat(lockFilePath); errors.Is(err, os.ErrNotExist) {
		lockFileNotExists = true
	}
	if !lockFileNotExists {
		return fmt.Errorf("lock file %s already exists or has error", lockFilePath)
	}
	f, err := os.Create(lockFilePath)
	if err != nil {
		return fmt.Errorf("failed to create lock file %s: %s", lockFilePath, err.Error())
	}
	f.Close()
	return nil
}

func UnlockFile(filePath string) error {
	lockFilePath := filePath + lockFileSuffix
	lockFileNotExists := false
	if _, err := os.Stat(lockFilePath); errors.Is(err, os.ErrNotExist) {
		lockFileNotExists = true
	}
	if lockFileNotExists {
		return nil
	}
	err := os.Remove(lockFilePath)
	if err != nil {
		return fmt.Errorf("failed to delete lock file %s: %s", lockFilePath, err.Error())
	}
	return nil
}

func WaitLockFile(filePath string, maxWaitMillis int64, forceUnlock bool) error {
	currentTime := time.Now()
	endTime := currentTime.Add(time.Duration(maxWaitMillis)*time.Millisecond)
	for !currentTime.After(endTime) {
		err := LockFile(filePath)
		if err == nil {
			return nil
		}
		time.Sleep(500*time.Millisecond)
		currentTime = time.Now()
	}
	if forceUnlock {
		err := UnlockFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to unlock file %s after waiting max %d milliseconds", filePath, maxWaitMillis)
		}
		err = LockFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to lock file %s after waiting max %d milliseconds and unlocking it", filePath, maxWaitMillis)
		}
		return nil
	}
	return fmt.Errorf("failed to lock file %s after max waiting %d milliseconds", filePath, maxWaitMillis)
}
