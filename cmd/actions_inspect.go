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
var actionsInspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Inspect an Action",
	Long:  ``,
	Args:  cobra.ExactArgs(1), // https://github.com/spf13/cobra/blob/master/user_guide.md
	Run: func(cmd *cobra.Command, args []string) {
		workspace.load().Flush()

		action := workspace.GetActionFromIndex(args[0])
		if action != nil {
			action.Inspect()
		} else {
			log.Fatalf("Action not found: %s", args[0])
		}
	},
}

func init() {
	actionsCmd.AddCommand(actionsInspectCmd)
}
