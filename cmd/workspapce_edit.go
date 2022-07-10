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
	"path/filepath"

	"github.com/spf13/cobra"
)

// installCmd represents the install command
var workspaceEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit the workspace",
	Long:  ``,
	Args:  cobra.ExactArgs(0), // https://github.com/spf13/cobra/blob/master/user_guide.md
	Run: func(cmd *cobra.Command, args []string) {
		workspace.load().Flush()

		workspaceConfigFilePath := filepath.Join(workspace.LocalPath, workspace.Config.WorkspaceConfig)
		// TODO: "code" should be configurable
		if editor == "code" {
			// When VS Code is used, open the whole workspace directory
			// and not only the workspace.poly file
			RunCommand(editor, workspace.LocalPath)
		} else {
			// We need to set interactive to true to forward stdin to the command
			// Otherwise the editor won't be able to receive input and the command
			// will fail with: "Too many errors from stdin"
			interactive = true
			RunCommand(editor, workspaceConfigFilePath)
		}
	},
}

func init() {
	workspaceCmd.AddCommand(workspaceEditCmd)
}
