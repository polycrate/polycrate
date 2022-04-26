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

var triggerCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a Workflow or an Action",
	Long:  `Run a Workflow or an Action`,
	Run: func(cmd *cobra.Command, args []string) {
		//bootstrapEnvVars()
		//bootstrapMounts()
		if len(args) > 2 {
			log.Fatal("'run' accepts a maximum of 2 arguments")
		} else if len(args) == 1 {
			// 1 Arg means we're running a Workflow
			// Delegate this to the runWorkflow Command
			runWorkflowCmd.Run(cmd, args)
		} else if len(args) == 2 {
			// 2 Args means we're running an Action
			// Determine Block
			block := args[0]

			currentBlock = workspace.getBlockByName(block)

			if currentBlock != nil {
				log.Infof("Found Block '%s'", currentBlock.Metadata.Name)
			} else {
				log.Fatalf("Block '%s' not found", block)
			}

			// Determine Action
			action := args[1]
			currentAction = currentBlock.getActionByName(action)

			if currentAction != nil {
				log.Infof("Found Action '%s' in Block '%s'", currentAction.Metadata.Name, currentBlock.Metadata.Name)
			} else {
				log.Fatalf("Action '%s' not found in Block '%s'", action, currentBlock.Metadata.Name)
			}
		} else {
			log.Fatalf("No arguments given. What do you expect me to do?")
		}
	},
}

func init() {
	rootCmd.AddCommand(triggerCmd)
}
