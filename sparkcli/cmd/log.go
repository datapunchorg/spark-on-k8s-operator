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

var ExecutorId int32
var FollowLogs bool

var logCommand = &cobra.Command{
	Use:   "log <submissionId>",
	Short: "Fetch logs of the application by submission id",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "must specify a submission id")
			return
		}

		submissionId := args[0]

		client := NewBasicAuthClient(ServerUrl, User, Password)
		client.PrintApplicationLog(submissionId, int(ExecutorId), FollowLogs)
	},
}

func init() {
	logCommand.Flags().Int32VarP(&ExecutorId, "executor", "e", -1,
		"id of the executor to fetch logs for")
	logCommand.Flags().BoolVarP(&FollowLogs, "follow", "f", false, "whether to stream the logs")
}
