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
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func TestLockUnlockFile(t *testing.T) {
	file, err := ioutil.TempFile("", "sparkcli_filelock_test_")
	assert.Nil(t, err)

	err = LockFile(file.Name())
	assert.Nil(t, err)

	err = LockFile(file.Name())
	assert.NotNil(t, err)

	err = UnlockFile(file.Name())
	assert.Nil(t, err)

	err = UnlockFile(file.Name())
	assert.Nil(t, err)

	defer UnlockFile(file.Name())
}

func TestWaitLockFile(t *testing.T) {
	file, err := ioutil.TempFile("", "sparkcli_filelock_test_")
	assert.Nil(t, err)

	filePath := file.Name()

	err = WaitLockFile(filePath, 10, false)
	assert.Nil(t, err)

	err = WaitLockFile(filePath, 10, false)
	assert.NotNil(t, err)

	err = UnlockFile(filePath)
	assert.Nil(t, err)

	err = WaitLockFile(filePath, 10, false)
	assert.Nil(t, err)

	err = WaitLockFile(filePath, 10, false)
	assert.NotNil(t, err)

	err = WaitLockFile(filePath, 10, true)
	assert.Nil(t, err)

	defer UnlockFile(filePath)
}

