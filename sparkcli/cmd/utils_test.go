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
	"testing"
)

func TestCheckLocalFile(t *testing.T) {
	result, err := CheckLocalFile("file1.txt")
	assert.Nil(t, err)
	assert.Equal(t, "file1.txt", result)

	result, err = CheckLocalFile("./file1.txt")
	assert.Nil(t, err)
	assert.Equal(t, "./file1.txt", result)

	result, err = CheckLocalFile("foo/file1.txt")
	assert.Nil(t, err)
	assert.Equal(t, "foo/file1.txt", result)

	result, err = CheckLocalFile("./foo/file1.txt")
	assert.Nil(t, err)
	assert.Equal(t, "./foo/file1.txt", result)

	result, err = CheckLocalFile("/foo/file1.txt")
	assert.Nil(t, err)
	assert.Equal(t, "/foo/file1.txt", result)

	result, err = CheckLocalFile("http://foo/file1.txt")
	assert.NotNil(t, err)
	assert.Equal(t, "", result)
}
