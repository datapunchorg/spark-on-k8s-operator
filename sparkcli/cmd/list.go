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
)

var Limit int64
var State string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List application submissions",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		client := NewBasicAuthClient(ServerUrl, User, Password)

		responseStr, _, err := client.ListSubmissions(Limit, State)
		if err != nil {
			ExitWithError(fmt.Sprintf("Failed to get application submissions: %s", err.Error()))
		}

		if OutputFile != "" {
			WriteOutputFileExitOnError(OutputFile, responseStr)
		}

		log.Printf("Response: %s", responseStr)
	},
}

func init() {
	addOutputFlag(listCmd)

	listCmd.Flags().Int64VarP(&Limit, "limit", "", 0,
		"limit of max number of items returned by the server")

	listCmd.Flags().StringVarP(&State, "state", "", "",
		"state of spark application to list")
}
