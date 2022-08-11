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
	"encoding/json"
	"fmt"
	apigatewayv1 "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apigateway/apis/v1"
	"github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apis/sparkoperator.k8s.io/v1beta2"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
	"time"
)

var SparkWaitAppCompletion = "spark.kubernetes.submission.waitAppCompletion"

var CreationSubmissionId string
var Overwrite bool

var ApplicationName string
var DesiredState string
var MaxWaitSeconds int64

var Image string
var SparkVersion string
var Type string
var DriverCores int32
var DriverMemory string
var NumExecutors int32
var ExecutorCores int32
var ExecutorMemory string

var Class string

var SparkConf []string

var FailureRetries int32

var submitCmd = &cobra.Command{
	Use:   "submit <application file> <application argument> <application argument>",
	Short: "Submit a Spark application",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "must specify an application file")
			return
		}

		client := NewBasicAuthClient(ServerUrl, User, Password)

		applicationFile := args[0]

		var applicationArgs []string
		if len(args) >= 2 {
			applicationArgs = args[1:]
		}

		sparkConfMap := map[string]string{}
		for _, entry := range SparkConf {
			index := strings.Index(entry, "=")
			if index == -1 {
				sparkConfMap[entry] = ""
			} else {
				key := entry[0:index]
				value := entry[index+1:]
				sparkConfMap[key] = value
			}
		}

		localApplicationFile, _ := CheckLocalFile(applicationFile)
		if localApplicationFile != "" {
			log.Printf("Uploading local applicaiton file %s", localApplicationFile)
			fileUrl, err := client.UploadFileToS3(localApplicationFile)
			if err != nil {
				log.Fatalf("Failed to upload applicaiton file %s: %s", localApplicationFile, err.Error())
			}
			applicationFile = fileUrl
			log.Printf("Uploaded file to %s", applicationFile)
		}

		var mainClass *string = nil
		if Class != "" {
			mainClass = &Class
		}

		request := apigatewayv1.SparkApplicationSubmissionRequest{
			ApplicationName: ApplicationName,
			DesiredState:    DesiredState,
			SparkApplicationSpec: v1beta2.SparkApplicationSpec{
				Image:               &Image,
				SparkVersion:        SparkVersion,
				Type:                v1beta2.SparkApplicationType(Type),
				MainClass:           mainClass,
				MainApplicationFile: &applicationFile,
				Arguments:           applicationArgs,
				Driver: v1beta2.DriverSpec{
					SparkPodSpec: v1beta2.SparkPodSpec{
						Cores:  &DriverCores,
						Memory: &DriverMemory,
					},
				},
				Executor: v1beta2.ExecutorSpec{
					Instances: &NumExecutors,
					SparkPodSpec: v1beta2.SparkPodSpec{
						Cores:  &ExecutorCores,
						Memory: &ExecutorMemory,
					},
				},
				SparkConf:      sparkConfMap,
				FailureRetries: &FailureRetries,
			},
		}

		submissionId := ""
		var err error
		if CreationSubmissionId == "" {
			if Overwrite {
				ExitWithError(fmt.Sprintf("Cannot overwrite Spark application without --id argument"))
			}
			submissionId, err = client.SubmitApplication(request)
			if err != nil {
				ExitWithError(fmt.Sprintf("Failed to submit application: %s", err.Error()))
			}
		} else {
			submissionId, err = client.SubmitApplicationWithId(request, CreationSubmissionId, Overwrite)
			if err != nil {
				ExitWithError(fmt.Sprintf("Failed to submit application with id %s: %s", CreationSubmissionId, err.Error()))
			}
		}

		log.Printf("Submitted application, submission id: %s", submissionId)

		if OutputFile != "" {
			out := map[string]string{
				"submissionId": submissionId,
			}
			bytes, err := json.Marshal(out)
			if err != nil {
				ExitWithError(fmt.Sprintf("Failed to marshal output to json: %s", err.Error()))
			}
			WriteOutputFileExitOnError(OutputFile, string(bytes))
		}

		if waitAppCompletion(sparkConfMap) {
			maxWaitSeconds := MaxWaitSeconds
			sleepInterval := 10 * time.Second
			startTime := time.Now()
			expireTime := startTime.Add(time.Duration(maxWaitSeconds) * time.Second)
			applicationFinished := false
			for time.Now().Before(expireTime) {
				// TODO add retry when getting status returns error
				statusResponseStr, statusResponse, err := client.GetApplicationStatus(submissionId)
				if err != nil {
					log.Fatalf("Failed to get status for application %s: %s", submissionId, err.Error())
				}
				state := statusResponse.State
				if strings.EqualFold(state, "COMPLETED") ||
					strings.EqualFold(state, "FAILED") ||
					strings.EqualFold(state, "SUBMISSION_FAILED") ||
					strings.EqualFold(state, "SUCCESS") ||
					strings.EqualFold(state, "SUCCEEDED") ||
					strings.EqualFold(state, "CANCELLED") {
					log.Printf("Application %s finished: %s", submissionId, statusResponseStr)
					applicationFinished = true
					break
				} else {
					str := fmt.Sprintf("Waiting until application %s finished (current state: %s)", submissionId, state)
					log.Printf(str)
					time.Sleep(sleepInterval)
				}
			}

			if !applicationFinished {
				log.Fatalf("Application %s not finished", submissionId)
			}
		}

		log.Printf("You could check application log by running: sparkcli --insecure --url %s log %s", ServerUrl, submissionId)
	},
}

func init() {
	addOutputFlag(submitCmd)

	submitCmd.Flags().StringVarP(&CreationSubmissionId, "id", "", "",
		"the unique id to submit Spark application, please only provide this argument when submitting a Spark streaming application and want to avoid running same application duplicately")

	submitCmd.Flags().BoolVarP(&Overwrite, "overwrite", "", false,
		"whether to overwrite (delete) old Spark application if it is already running")

	submitCmd.Flags().StringVarP(&ApplicationName, "application-name", "", "",
		"the name of the Spark application")
	submitCmd.Flags().StringVarP(&DesiredState, "desired-state", "", "",
		"the desired state of the Spark application")
	submitCmd.Flags().Int64VarP(&MaxWaitSeconds, "max-wait-seconds", "", 24*60*60,
		"")

	submitCmd.Flags().StringVarP(&Class, "class", "", "",
		"the main class of the Spark application")

	submitCmd.Flags().StringVarP(&Image, "image", "", "",
		"the name of the Spark image")
	submitCmd.Flags().StringVarP(&SparkVersion, "spark-version", "", "",
		"the Spark version")
	submitCmd.Flags().StringVarP(&Type, "type", "", "",
		"the application type")
	submitCmd.Flags().StringArrayVarP(&SparkConf, "conf", "", []string{},
		"Spark config")

	submitCmd.Flags().Int32VarP(&DriverCores, "driver-cores", "", 1,
		"")
	submitCmd.Flags().StringVarP(&DriverMemory, "driver-memory", "", "1g",
		"")

	submitCmd.Flags().Int32VarP(&NumExecutors, "num-executors", "", 1,
		"")
	submitCmd.Flags().Int32VarP(&ExecutorCores, "executor-cores", "", 1,
		"")
	submitCmd.Flags().StringVarP(&ExecutorMemory, "executor-memory", "", "1g",
		"")

	submitCmd.Flags().Int32VarP(&FailureRetries, "failure-retries", "", 0,
		"")

	// Make flag parsing stop after the first non-flag arg
	submitCmd.Flags().SetInterspersed(false)
}

func waitAppCompletion(sparkConf map[string]string) bool {
	value, ok := sparkConf[SparkWaitAppCompletion]
	// default value
	wait := true
	if ok {
		wait = !strings.EqualFold(value, "false")
	}
	return wait
}
