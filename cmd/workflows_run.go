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
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var workflowName string
var stepName string
var stepIndex int

var runWorkflowCmd = &cobra.Command{
	Use:   "run",
	Short: "Run Workflow",
	Long:  `Run Workflow`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		workflowName = args[0]
		_w := cmd.Flags().Lookup("workspace").Value.String()

		cmdKey := ContextKey("cmd")
		ctx := context.WithValue(context.Background(), cmdKey, cmd)
		ctx, cancel, err := polycrate.NewTransaction(ctx)
		defer polycrate.StopTransaction(ctx, cancel)

		if err != nil {
			log.Fatal(err)
		}

		err = cmdRunWorkflow(ctx, _w, args, stepName)
		if err != nil {
			log.Fatal(err)
			return err
		}

		return nil
	},
}

func cmdRunWorkflow(ctx context.Context, s string, args []string, stepName string) error {
	ctx, workspace, err := polycrate.GetWorkspaceWithContext(ctx, s)
	if err != nil {
		return err
	}

	ctx, err = workspace.RunWorkflowWithContext(ctx, args[0], stepName)
	if err != nil {
		return err
	}
	return nil
}

func init() {
	runWorkflowCmd.Flags().StringVar(&stepName, "step", "", "The name of the step to be run")
	runWorkflowCmd.Flags().IntVar(&stepIndex, "step-index", -1, "The index of the step to be executed. Currently no-op")

	workflowsCmd.AddCommand(runWorkflowCmd)
}

// func runWorkflow(pipeline string) {
// 	// Check if pipeline exists
// 	log.Info("Running pipeline ", pipeline)
// 	if workspace.Workflows[0].Steps != nil {
// 		for _, step := range workspace.Workflows[0].Steps {
// 			log.Info("Running step ", step.Name)
// 			var err error
// 			//pluginCallExitCode, err = callPlugin(step.Block, step.Action)
// 			CheckErr(err)
// 		}
// 	} else {
// 		log.Fatal("No steps defined")
// 	}
// }
