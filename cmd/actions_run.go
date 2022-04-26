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
var actionsRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run an Action",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		workspace.load()
		if workspace.Flush() != nil {
			log.Fatal(workspace.Flush)
		}

		// block, action, err := workspace.resolveActionAddress(args[0])

		// if block != nil {
		// 	workspace.setCurrentBlock(block)
		// 	if action != nil {
		// 		workspace.setCurrentAction(action)
		// 		action.Run()
		// 	} else {
		// 		log.Fatal(err)
		// 	}
		// } else {
		// 	log.Fatal(err)
		// }
		workspace.RunAction(args[0])

	},
}

func init() {
	actionsCmd.AddCommand(actionsRunCmd)
}