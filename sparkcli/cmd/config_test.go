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
	"strings"
	"testing"
)

func TestConfigData(t *testing.T) {
	config := NewConfig()
	_, err := config.GetCurrentCredential()
	assert.NotNil(t, err)

	config.UpdateCurrentUserPassword("http://server1/sparkapi/v1", "user1", "password1")

	credential, err := config.GetCurrentCredential()
	assert.Nil(t, err)
	assert.Equal(t, "http://server1/sparkapi/v1", credential.Server)
	assert.Equal(t, "user1", credential.User)
	assert.Equal(t, "password1", credential.Password)

	// run UpdateCurrentUserPassword again with same values
	config.UpdateCurrentUserPassword("http://server1/sparkapi/v1", "user1", "password1")

	credential, err = config.GetCurrentCredential()
	assert.Nil(t, err)
	assert.Equal(t, "http://server1/sparkapi/v1", credential.Server)
	assert.Equal(t, "user1", credential.User)
	assert.Equal(t, "password1", credential.Password)

	// update password
	config.UpdateCurrentUserPassword("http://server1/sparkapi/v1", "user1", "password2")

	credential, err = config.GetCurrentCredential()
	assert.Nil(t, err)
	assert.Equal(t, "http://server1/sparkapi/v1", credential.Server)
	assert.Equal(t, "user1", credential.User)
	assert.Equal(t, "password2", credential.Password)

	// update server
	config.UpdateCurrentUserPassword("http://server2/sparkapi/v1", "user1", "password2")

	credential, err = config.GetCurrentCredential()
	assert.Nil(t, err)
	assert.Equal(t, "http://server2/sparkapi/v1", credential.Server)
	assert.Equal(t, "user1", credential.User)
	assert.Equal(t, "password2", credential.Password)

	// update user
	config.UpdateCurrentUserPassword("http://server2/sparkapi/v1", "user2", "password2")

	credential, err = config.GetCurrentCredential()
	assert.Nil(t, err)
	assert.Equal(t, "http://server2/sparkapi/v1", credential.Server)
	assert.Equal(t, "user2", credential.User)
	assert.Equal(t, "password2", credential.Password)

	// update all values
	config.UpdateCurrentUserPassword("http://server3/sparkapi/v1", "user3", "password3")

	credential, err = config.GetCurrentCredential()
	assert.Nil(t, err)
	assert.Equal(t, "http://server3/sparkapi/v1", credential.Server)
	assert.Equal(t, "user3", credential.User)
	assert.Equal(t, "password3", credential.Password)
}

func TestWriteReadConfigFile(t *testing.T) {
	config := NewConfig()
	err := config.LoadIfExists("not_exist_file.abc")
	assert.Nil(t, err)

	config.UpdateCurrentUserPassword("http://server1/sparkapi/v1", "user1", "password1")

	file, err := ioutil.TempFile("", "sparkcli_config_test_")
	assert.Nil(t, err)

	filePath := file.Name()
	config.SaveToFile(filePath)

	config = NewConfig()
	config.LoadIfExists(filePath)

	credential, err := config.GetCurrentCredential()
	assert.Nil(t, err)
	assert.Equal(t, "http://server1/sparkapi/v1", credential.Server)
	assert.Equal(t, "user1", credential.User)
	assert.Equal(t, "password1", credential.Password)
}

func TestGetDefaultConfigFilePath(t *testing.T) {
	filePath, err := GetDefaultConfigFilePath()
	assert.Nil(t, err)
	assert.True(t, strings.HasSuffix(filePath, "/.sparkcli/config"))
}

func TestGetCredentialByServer(t *testing.T) {
	config := NewConfig()
	credential, err := config.GetCredentialByServer("server1")
	assert.NotNil(t, err)
	assert.Equal(t, "", credential.Server)

	config.UpdateCurrentUserPassword("server1", "user1", "password1")

	credential, err = config.GetCurrentCredential()
	assert.Nil(t, err)
	assert.Equal(t, "server1", credential.Server)
	assert.Equal(t, "user1", credential.User)
	assert.Equal(t, "password1", credential.Password)

	credential, err = config.GetCredentialByServer("server1")
	assert.Nil(t, err)
	assert.Equal(t, "server1", credential.Server)
	assert.Equal(t, "user1", credential.User)
	assert.Equal(t, "password1", credential.Password)

	config.UpdateCurrentUserPassword("server2", "user2", "password2")

	credential, err = config.GetCurrentCredential()
	assert.Nil(t, err)
	assert.Equal(t, "server2", credential.Server)
	assert.Equal(t, "user2", credential.User)
	assert.Equal(t, "password2", credential.Password)

	credential, err = config.GetCredentialByServer("server1")
	assert.Nil(t, err)
	assert.Equal(t, "server1", credential.Server)
	assert.Equal(t, "user1", credential.User)
	assert.Equal(t, "password1", credential.Password)

	credential, err = config.GetCredentialByServer("server2")
	assert.Nil(t, err)
	assert.Equal(t, "server2", credential.Server)
	assert.Equal(t, "user2", credential.User)
	assert.Equal(t, "password2", credential.Password)

	config.UpdateCurrentUserPassword("server1", "user1", "password11")

	credential, err = config.GetCurrentCredential()
	assert.Nil(t, err)
	assert.Equal(t, "server1", credential.Server)
	assert.Equal(t, "user1", credential.User)
	assert.Equal(t, "password11", credential.Password)

	credential, err = config.GetCredentialByServer("server1")
	assert.Nil(t, err)
	assert.Equal(t, "server1", credential.Server)
	assert.Equal(t, "user1", credential.User)
	assert.Equal(t, "password11", credential.Password)

	credential, err = config.GetCredentialByServer("server2")
	assert.Nil(t, err)
	assert.Equal(t, "server2", credential.Server)
	assert.Equal(t, "user2", credential.User)
	assert.Equal(t, "password2", credential.Password)
}