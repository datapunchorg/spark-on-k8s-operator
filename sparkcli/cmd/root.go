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
	"log"
	"os"
)

var ServerUrl string
var InsecureHttps bool
var User string
var Password string
var IgnoreCredentialCache bool

var OutputFile string

var rootCmd = &cobra.Command{
	Use:   "sparkcli",
	Short: "sparkcli is the command-line tool to submit Spark application to Spark Operator API Gateway",
	Long: `sparkcli is the command-line tool to submit Spark application to Spark Operator API Gateway. It supports submitting, deleting and 
           checking status of Spark application. It also supports fetching application logs.`,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&ServerUrl, "url", "l", "",
		"The API gateway url to which the Spark application is to be submitted")
	rootCmd.PersistentFlags().BoolVarP(&InsecureHttps, "insecure", "k", false,
		"Whether to skip SSL certificate verification")
	rootCmd.PersistentFlags().StringVarP(&User, "user", "u", "",
		"User name to connect to API gateway")
	rootCmd.PersistentFlags().StringVarP(&Password, "password", "p", "",
		"User password to connect to API gateway")
	rootCmd.PersistentFlags().BoolVarP(&IgnoreCredentialCache, "ignore-credential-cache", "", false,
		"Do not use credential from values in local config")
	rootCmd.AddCommand(uploadCmd, submitCmd, statusCmd, logCommand, deleteCmd, listCmd)
}

func Execute() {
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		InsecureTLS(InsecureHttps)

		if !IgnoreCredentialCache {
			filePath, err := GetDefaultConfigFilePath()
			if err != nil {
				log.Fatalf("Do not save credential due to failing to get default config file path: %s", err.Error())
			}

			config := NewConfig()
			err = config.LoadIfExists(filePath)
			if err != nil {
				log.Fatalf("Failed to load config from file %s: %s", filePath, err.Error())
			}

			if ServerUrl != "" && User != "" && Password != "" {
				config.UpdateCurrentUserPassword(ServerUrl, User, Password)
				err = config.SaveToFile(filePath)
				if err != nil {
					log.Fatalf("Failed to save credential to file %s: %s", filePath, err.Error())
				}
			}

			if ServerUrl == "" && User == "" && Password == "" {
				credential, err := config.GetCurrentCredential()
				if err != nil {
					log.Fatalf("Failed to get current credential")
				}

				if ServerUrl == "" {
					ServerUrl = credential.Server
				}
				if User == "" {
					User = credential.User
				}
				if Password == "" {
					Password = credential.Password
				}
				log.Printf("No server information found in the command arguments, use credential from current context, server: %s, user: %s", ServerUrl, User)
			} else if ServerUrl != "" && (User == "" || Password == "") {
				credential, err := config.GetCredentialByServer(ServerUrl)
				if err != nil {
					log.Fatalf("Failed to get credential for server %s", ServerUrl)
				}
				if User == "" {
					User = credential.User
					log.Printf("Server %s provided in the command arguments, search local config and use user: %s", ServerUrl, User)
				}
				if Password == "" {
					Password = credential.Password
					log.Printf("Server %s provided in the command arguments, search local config and use password for user: %s", ServerUrl, User)
				}
			}
		}

		if ServerUrl == "" {
			log.Fatalf("Please provide server information using --url argument, e.g. --url http://server:port/sparkapi/v1")
		}
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
	}
}

func addOutputFlag(cmd *cobra.Command)  {
	cmd.Flags().StringVarP(&OutputFile, "output", "o", "",
		"the file to write output information")
}
