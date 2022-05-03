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
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_addMissingKeysInStringMap(t *testing.T) {
	map1 := map[string]string {
		"key1": "value1",
		"key2": "value2",
	}

	addMissingKeysInStringMap(map1, nil)
	assert.Equal(t, "value1", map1["key1"])
	assert.Equal(t, "value2", map1["key2"])

	addMissingKeysInStringMap(map1, map[string]string{})
	assert.Equal(t, "value1", map1["key1"])
	assert.Equal(t, "value2", map1["key2"])

	map2 := map[string]string {
		"key2": "value22",
		"key3": "value33",
	}
	addMissingKeysInStringMap(map1, map2)
	assert.Equal(t, 3, len(map1))
	assert.Equal(t, "value1", map1["key1"])
	assert.Equal(t, "value2", map1["key2"])
	assert.Equal(t, "value33", map1["key3"])
}

func Test_findSparkImageName(t *testing.T) {
	config := []SparkImageConfig {
		{Version: "3.1", Type: "java", Image: "image1"},
		{Version: "3.2", Type: "Python", Image: "image2"},
	}
	foundImage, ok := findSparkImageName(config, "3.0", "java")
	assert.Equal(t, "", foundImage)
	assert.Equal(t, false, ok)
	foundImage, ok = findSparkImageName(config, "3.1", "java")
	assert.Equal(t, "image1", foundImage)
	assert.Equal(t, true, ok)
	foundImage, ok = findSparkImageName(config, "3.2", "python")
	assert.Equal(t, "image2", foundImage)
	assert.Equal(t, true, ok)
}
