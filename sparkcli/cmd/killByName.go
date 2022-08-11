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

var KillByNameMaxApplicationCount int32

var killByNameCmd = &cobra.Command{
	Use:   "kill-by-name <applicationName>",
	Short: "Kill the application submission by name",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "must specify application name")
			return
		}

		appName := args[0]

		client := NewBasicAuthClient(ServerUrl, User, Password)

		responseStr, _, err := client.KillApplicationByName(appName, int(KillByNameMaxApplicationCount))
		if err != nil {
			log.Fatalf("Failed to kill application: %s", err.Error())
		}

		log.Printf("Response: %s", responseStr)
	},
}

func init() {
	killByNameCmd.Flags().Int32VarP(&KillByNameMaxApplicationCount, "max-count", "", 1,
		"")
}
