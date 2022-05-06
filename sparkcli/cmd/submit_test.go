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

func Test_waitAppCompletion(t *testing.T) {
	assert.Equal(t, true, waitAppCompletion(map[string]string{}))
	assert.Equal(t, true, waitAppCompletion(map[string]string{
		"key1": "value1",
	}))
	assert.Equal(t, true, waitAppCompletion(map[string]string{
		"key1": "value1",
		"spark.kubernetes.submission.waitAppCompletion": "",
	}))
	assert.Equal(t, true, waitAppCompletion(map[string]string{
		"key1": "value1",
		"spark.kubernetes.submission.waitAppCompletion": "true",
	}))
	assert.Equal(t, true, waitAppCompletion(map[string]string{
		"key1": "value1",
		"spark.kubernetes.submission.waitAppCompletion": "invalid value",
	}))
	assert.Equal(t, false, waitAppCompletion(map[string]string{
		"key1": "value1",
		"spark.kubernetes.submission.waitAppCompletion": "false",
	}))
	assert.Equal(t, false, waitAppCompletion(map[string]string{
		"key1": "value1",
		"spark.kubernetes.submission.waitAppCompletion": "False",
	}))
}
