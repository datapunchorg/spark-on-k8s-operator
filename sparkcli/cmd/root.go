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
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var ServerUrl string
var InsecureHttps bool
var User string
var Password string

var rootCmd = &cobra.Command{
	Use:   "sparkcli",
	Short: "sparkcli is the command-line tool to submit Spark application to Spark Operator API Gateway",
	Long: `sparkcli is the command-line tool to submit Spark application to Spark Operator API Gateway. It supports submitting, deleting and 
           checking status of Spark application. It also supports fetching application logs.`,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&ServerUrl, "url", "l", "http://localhost:8080/sparkapi/v1",
		"The API gateway url to which the Spark application is to be submitted")
	rootCmd.PersistentFlags().BoolVarP(&InsecureHttps, "insecure", "k", false,
		"Whether to skip SSL certificate verification")
	rootCmd.PersistentFlags().StringVarP(&User, "user", "u", "",
		"User name to connect to API gateway")
	rootCmd.PersistentFlags().StringVarP(&Password, "password", "p", "",
		"User password to connect to API gateway")
	rootCmd.AddCommand(uploadCmd, submitCmd, statusCmd, logCommand, deleteCmd)
}

func Execute() {
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		InsecureTLS(InsecureHttps)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
	}
}
