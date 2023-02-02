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
	"fmt"

	"github.com/spf13/cobra"
)

// installCmd represents the install command
var workflowsInspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Inspect a Workflow",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		_w := cmd.Flags().Lookup("workspace").Value.String()

		ctx := context.Background()
		ctx, cancel, err := polycrate.NewTransaction(ctx, cmd)
		defer polycrate.StopTransaction(ctx, cancel)

		log := polycrate.GetContextLogger(ctx)

		ctx, workspace, err := polycrate.GetWorkspaceWithContext(ctx, _w, true)
		if err != nil {
			log.Fatal(err)
		}

		var workflow *Workflow
		ctx, workflow, err = workspace.GetWorkflowWithContext(ctx, args[0])
		if err != nil {
			log.Fatal(err)
		}

		if workflow != nil {
			workflow.Inspect()
		} else {
			err := fmt.Errorf("Workflow not found: %s", args[0])
			log.Fatal(err)
		}
	},
}

func init() {
	workflowsCmd.AddCommand(workflowsInspectCmd)
}
