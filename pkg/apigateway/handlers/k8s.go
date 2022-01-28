/*
Copyright spark-on-k8s-operator contributors

The code in this file is copied from https://github.com/GoogleCloudPlatform/spark-on-k8s-operator
and modified.

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
	"fmt"
	crdclientset "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/client/clientset/versioned"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

func getKubeConfigPath() string {
	var kubeConfigEnv = os.Getenv("KUBECONFIG")
	if len(kubeConfigEnv) == 0 {
		return os.Getenv("HOME") + "/.kube/config"
	}
	return kubeConfigEnv
}

func buildConfig(kubeConfig string) (*rest.Config, error) {
	// Check if kubeConfig exist
	if _, err := os.Stat(kubeConfig); os.IsNotExist(err) {
		// Try InClusterConfig for sparkctl running in a pod
		config, err := rest.InClusterConfig()
		if err != nil {
			return nil, err
		}

		return config, nil
	}

	return clientcmd.BuildConfigFromFlags("", kubeConfig)
}

func getKubeClient(kubeConfig string) (clientset.Interface, error) {
	config, err := buildConfig(kubeConfig)
	if err != nil {
		return nil, err
	}
	return getKubeClientForConfig(config)
}

func getKubeClientForConfig(config *rest.Config) (clientset.Interface, error) {
	return clientset.NewForConfig(config)
}

func getSparkApplicationClient(kubeConfig string) (crdclientset.Interface, error) {
	config, err := buildConfig(kubeConfig)
	if err != nil {
		return nil, err
	}
	return getSparkApplicationClientForConfig(config)
}

func createSparkApplicationClient() (crdclientset.Interface, error) {
	kubeConfig := getKubeConfigPath()
	crdClient, err := getSparkApplicationClient(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get SparkApplication client: %s", err.Error())
	}
	return crdClient, nil
}

func getSparkApplicationClientForConfig(config *rest.Config) (crdclientset.Interface, error) {
	return crdclientset.NewForConfig(config)
}


