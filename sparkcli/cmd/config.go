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
	"bytes"
	"errors"
	"fmt"
	"k8s.io/apimachinery/pkg/util/yaml"
	"os"
	"os/user"
	"path/filepath"
	yaml2 "sigs.k8s.io/yaml"
)

type Config struct {
	ApiVersion string `json:"apiVersion" yaml:"apiVersion"`
	Kind string `json:"kind" yaml:"kind"`
	Clusters []ClusterConfig `json:"clusters" yaml:"clusters"`
	Contexts []ContextConfig `json:"contexts" yaml:"contexts"`
	CurrentContext string `json:"current-context" yaml:"current-context"`
	Users []UserConfig `json:"users" yaml:"users"`
}

type ClusterConfig struct {
	Cluster ClusterDetail `json:"cluster" yaml:"cluster"`
	Name string `json:"name" yaml:"name"`
}

type ClusterDetail struct {
	Server string `json:"server" yaml:"server"`
}

type ContextConfig struct {
	Context ContextDetail `json:"context" yaml:"context"`
	Name string `json:"name" yaml:"name"`
}

type ContextDetail struct {
	Cluster string `json:"cluster" yaml:"cluster"`
	User string `json:"user" yaml:"user"`
}

type UserConfig struct {
	User UserDetail `json:"user" yaml:"user"`
	Name string `json:"name" yaml:"name"`
}

type UserDetail struct {
	Password string `json:"password" yaml:"password"`
}

type ServerCredential struct {
	Server string
	User string
	Password string
}

func GetDefaultConfigFilePath() (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %s", err.Error())
	}
	homeDir := user.HomeDir
	if homeDir == "" {
		return "", fmt.Errorf("failed to get current user home directory (empty value)")
	}
	sparkcliDir := filepath.Join(homeDir, ".sparkcli")
	err = os.MkdirAll(sparkcliDir, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("failed to create dir %s: %s", sparkcliDir, err.Error())
	}
	return filepath.Join(sparkcliDir, "config"), nil
}

func NewConfig() Config {
	return Config{
		Kind: "SparkcliConfig",
		ApiVersion: "v1",
		Clusters: make([]ClusterConfig, 0, 10),
		Contexts: make([]ContextConfig, 0, 10),
	}
}

func (t *Config) GetCurrentCredential() (ServerCredential, error) {
	currentContext := t.CurrentContext
	if currentContext == "" {
		return ServerCredential{}, fmt.Errorf("cannot get current credential due to current context not set")
	}

	var foundContextConfig *ContextConfig = nil
	for i := range t.Contexts {
		if t.Contexts[i].Name == currentContext {
			foundContextConfig = &t.Contexts[i]
			break
		}
	}
	if foundContextConfig == nil {
		return ServerCredential{}, fmt.Errorf("cannot get current credential due to details not found for context %s", currentContext)
	}

	var foundClusterConfig *ClusterConfig = nil
	for i := range t.Clusters {
		if t.Clusters[i].Name == foundContextConfig.Context.Cluster {
			foundClusterConfig = &t.Clusters[i]
			break
		}
	}
	if foundClusterConfig == nil {
		return ServerCredential{}, fmt.Errorf("cannot get current credential due to details not found for cluster %s", foundContextConfig.Context.Cluster)
	}

	var foundUserConfig *UserConfig = nil
	for i := range t.Users {
		if t.Users[i].Name == foundContextConfig.Context.User {
			foundUserConfig = &t.Users[i]
			break
		}
	}
	if foundUserConfig == nil {
		return ServerCredential{}, fmt.Errorf("cannot get current credential due to details not found for user %s", foundContextConfig.Context.User)
	}

	return ServerCredential{
		Server: foundClusterConfig.Cluster.Server,
		User: foundUserConfig.Name,
		Password: foundUserConfig.User.Password,
	}, nil
}

func (t *Config) GetCredentialByServer(server string) (ServerCredential, error) {
	var foundClusterConfig *ClusterConfig = nil
	for i := range t.Clusters {
		if t.Clusters[i].Cluster.Server == server {
			foundClusterConfig = &t.Clusters[i]
			break
		}
	}
	if foundClusterConfig == nil {
		return ServerCredential{}, fmt.Errorf("did not find cluster for server %s", server)
	}

	var foundContextConfig *ContextConfig = nil
	for i := range t.Contexts {
		if t.Contexts[i].Context.Cluster == foundClusterConfig.Name {
			foundContextConfig = &t.Contexts[i]
			break
		}
	}
	if foundContextConfig == nil {
		return ServerCredential{}, fmt.Errorf("did not find context for cluster %s", foundClusterConfig.Name)
	}

	var foundUserConfig *UserConfig = nil
	for i := range t.Users {
		if t.Users[i].Name == foundContextConfig.Context.User {
			foundUserConfig = &t.Users[i]
			break
		}
	}
	if foundUserConfig == nil {
		return ServerCredential{}, fmt.Errorf("did not find password for user %s", foundContextConfig.Context.User)
	}

	return ServerCredential{
		Server: foundClusterConfig.Cluster.Server,
		User: foundUserConfig.Name,
		Password: foundUserConfig.User.Password,
	}, nil
}

func (t *Config) UpdateCurrentUserPassword(server string, user string, password string)  {
	var foundClusterConfig *ClusterConfig = nil
	for i := range t.Clusters {
		if t.Clusters[i].Name == server {
			foundClusterConfig = &t.Clusters[i]
			break
		}
	}
	clusterDetail := ClusterDetail{
		Server: server,
	}
	if foundClusterConfig == nil {
		t.Clusters = append(t.Clusters, ClusterConfig{
			Cluster: clusterDetail,
			Name: server,
		})
	} else {
		foundClusterConfig.Cluster = clusterDetail
	}

	var foundContextConfig *ContextConfig = nil
	for i := range t.Contexts {
		if t.Contexts[i].Name == server {
			foundContextConfig = &t.Contexts[i]
			break
		}
	}
	contextDetail := ContextDetail{
		Cluster: server,
		User: user,
	}
	if foundContextConfig == nil {
		t.Contexts = append(t.Contexts, ContextConfig{
			Name: server,
			Context: contextDetail,
		})
	} else {
		foundContextConfig.Context = contextDetail
	}

	t.CurrentContext = server

	var foundUserConfig *UserConfig = nil
	for i := range t.Users {
		if t.Users[i].Name == user {
			foundUserConfig = &t.Users[i]
			break
		}
	}
	userDetail := UserDetail{
		Password: password,
	}
	if foundUserConfig == nil {
		t.Users = append(t.Users, UserConfig{
			Name: user,
			User: userDetail,
		})
	} else {
		foundUserConfig.User = userDetail
	}
}

func (t *Config) LoadIfExists(filePath string) error {
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		return nil
	}

	bytesData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read filePath %s: %s", filePath, err.Error())
	}
	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(bytesData), len(bytesData))
	err = decoder.Decode(t)
	if err != nil {
		return fmt.Errorf("failed to unmarshal data from filePath %s: %s", filePath, err.Error())
	}
	return nil
}

// TODO add file lock to prevent concurrent write to config file

func (t *Config) SaveToFile(filePath string) error {
	WaitLockFile(filePath, 5 * 1000, true)
	defer UnlockFile(filePath)

	bytes, err := yaml2.Marshal(*t)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %s", err.Error())
	}
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %s", filePath, err.Error())
	}
	defer f.Close()
	_, err = f.WriteString(string(bytes))
	if err != nil {
		return fmt.Errorf("failed to write file %s: %s", filePath, err.Error())
	}
	return nil
}
