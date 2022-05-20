package handlers

import (
	"fmt"
	"github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apis/sparkoperator.k8s.io/v1beta2"
)

func GetDefaultUIServiceName(app *v1beta2.SparkApplication) string {
	return fmt.Sprintf("%s-ui-svc", app.Name)
}

