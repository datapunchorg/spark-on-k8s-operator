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
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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

	result, err = CheckLocalFile("foo/dir1")
	assert.Nil(t, err)
	assert.Equal(t, "foo/dir1", result)

	result, err = CheckLocalFile("./foo/file1.txt")
	assert.Nil(t, err)
	assert.Equal(t, "./foo/file1.txt", result)

	result, err = CheckLocalFile("/foo/file1.txt")
	assert.Nil(t, err)
	assert.Equal(t, "/foo/file1.txt", result)

	result, err = CheckLocalFile("/foo/dir1")
	assert.Nil(t, err)
	assert.Equal(t, "/foo/dir1", result)

	result, err = CheckLocalFile("http://foo/file1.txt")
	assert.NotNil(t, err)
	assert.Equal(t, "", result)
}

func TestZipDirAndSaveInDir(t *testing.T) {
	rootDir, err := ioutil.TempDir("", "test")
	assert.Nil(t, err)
	dir1, err := ioutil.TempDir(rootDir, "dir1")
	assert.Nil(t, err)

	dirName := filepath.Base(dir1)
	zipFileName := dirName + ".zip"

	zipFilePath, err := ZipDirAndSaveInDir(dir1, rootDir)
	assert.Nil(t, err)
	assert.True(t, strings.HasSuffix(zipFilePath, "/"+zipFileName))

	stat, err := os.Stat(zipFilePath)
	assert.Nil(t, err)
	assert.False(t, stat.IsDir())
	assert.True(t, stat.Size() > 0)

	dir2, err := ioutil.TempDir(dir1, "dir2")
	assert.Nil(t, err)

	file1, err := ioutil.TempFile(dir1, "file1_*.txt")
	file1.Close()
	assert.Nil(t, err)
	err = ioutil.WriteFile(file1.Name(), []byte("file1 content"), fs.ModePerm)
	assert.Nil(t, err)

	file2, err := ioutil.TempFile(dir2, "file2_*.txt")
	file2.Close()
	assert.Nil(t, err)
	err = ioutil.WriteFile(file2.Name(), []byte("file2 content"), fs.ModePerm)
	assert.Nil(t, err)

	file3, err := ioutil.TempFile(dir2, "file3_*.txt")
	file3.Close()
	assert.Nil(t, err)
	err = ioutil.WriteFile(file3.Name(), []byte("file3 content"), fs.ModePerm)
	assert.Nil(t, err)

	zipFilePath, err = ZipDirAndSaveInDir(dir1, rootDir)
	assert.Nil(t, err)
	assert.True(t, strings.HasSuffix(zipFilePath, "/"+zipFileName))
}
