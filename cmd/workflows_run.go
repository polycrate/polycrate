/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var runPipelineCmd = &cobra.Command{
	Use:   "run",
	Short: "Run pipeline",
	Long:  `Run pipeline`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 && args[0] != "" {
			pipeline = args[0]
		} else {
			pipeline = "default"
		}
		loadWorkspace()

		runPipeline(pipeline)
	},
}

func init() {
	pipelinesCmd.AddCommand(runPipelineCmd)
}

func runPipeline(pipeline string) {
	// Check if pipeline exists
	log.Info("Running pipeline ", pipeline)
	if workspace.Workflows[0].Steps != nil {
		for _, step := range workspace.Workflows[0].Steps {
			log.Info("Running step ", step.Display)
			var err error
			pluginCallExitCode, err = callPlugin(step.Trigger, step.Trigger)
			CheckErr(err)
		}
	} else {
		log.Fatal("No steps defined")
	}
}
