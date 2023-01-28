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
	"github.com/spf13/cobra"
)

//var workspaceVersion string

// installCmd represents the install command
var workspaceInstallCmd = &cobra.Command{
	Hidden:     true,
	Deprecated: "don't use this",
	Use:        "install BLOCK1 BLOCK2:0.0.1",
	Short:      "Install Blocks",
	Long:       ``,
	Args:       cobra.ExactArgs(1),
	//Args: cobra.RangeArgs(1, 2), // https://github.com/spf13/cobra/blob/master/user_guide.md
	Run: func(cmd *cobra.Command, args []string) {

		// workspaceName, workspaceVersion, err := registry.resolveArg(args[0])
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// Check registry workspace index
		// registryWorkspace, err := registry.GetWorkspace(workspaceName)
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// // Check local workspace index
		// if localWorkspaceIndex[workspaceName] != "" {
		// 	// We found a workspace with that name in the index
		// 	path := localWorkspaceIndex[workspaceName]
		// 	log.WithFields(log.Fields{
		// 		"workspace": workspaceName,
		// 		"path":      path,
		// 	}).Fatalf("Workspace is already installed")
		// }

		// if registryWorkspace != nil {
		// 	// Compile installation path
		// 	workspaceInstallDir := filepath.Join(polycrateWorkspaceDir, workspaceName)

		// 	err := registryWorkspace.Install(workspaceInstallDir, workspaceVersion)
		// 	if err != nil {
		// 		log.Fatal(err)
		// 	}
		// }

	},
}

func init() {
	//workspaceCmd.AddCommand(workspaceInstallCmd)

}
