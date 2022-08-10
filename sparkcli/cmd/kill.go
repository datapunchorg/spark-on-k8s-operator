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
	"log"
	"os"

	"github.com/spf13/cobra"
)

var killCmd = &cobra.Command{
	Use:   "kill <submissionId>",
	Short: "Kill the application submission",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "must specify a submission id")
			return
		}

		submissionId := args[0]

		client := NewBasicAuthClient(ServerUrl, User, Password)

		responseStr, _, err := client.KillApplication(submissionId)
		if err != nil {
			log.Fatalf("Failed to kill application: %s", err.Error())
		}

		log.Printf("Response: %s", responseStr)
	},
}
