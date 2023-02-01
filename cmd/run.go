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

// installCmd represents the install command
var runCmd = &cobra.Command{
	Use:   "run block|workflow [action]",
	Short: "Run a Workflow or an Action",
	Long: `
To run a Workflow, use this command with 1 argument - the Workflow name. 
To run an Action, use this command with 2 arguments - the Block name and the Action name.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			// Run a Worlflow
			err := runWorkflowCmd.RunE(cmd, args)
			if err != nil {
				log.Fatal(err)
			}
		} else if len(args) == 2 {
			//action := strings.Join([]string{args[0], args[1]}, ".")
			actionsRunCmd.Run(cmd, []string{args[0], args[1]})
		}

	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		// workspace.SaveRevision().Flush()
		//workspace.Sync().Flush()

	},
	Args: cobra.RangeArgs(1, 2), // https://github.com/spf13/cobra/blob/master/user_guide.md
}

func init() {
	rootCmd.AddCommand(runCmd)
}
