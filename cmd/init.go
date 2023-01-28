/*
Copyright © 2021 Fabian Peter <fp@ayedo.de>

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

var fileUrl string

// installCmd represents the install command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initalize a workspace",
	Long:  `Initalize a workspace`,
	Args:  cobra.ExactArgs(0), // https://github.com/spf13/cobra/blob/master/user_guide.md
	Run: func(cmd *cobra.Command, args []string) {
		workspaceCreateCmd.Run(cmd, args)
	},
}

func init() {
	initCmd.Flags().StringVar(&withName, "with-name", "", "The name of the workspace")
	initCmd.Flags().BoolVar(&withSshKeys, "with-ssh-keys", true, "Create SSH keys")
	initCmd.Flags().BoolVar(&withSshKeys, "with-config", true, "Create config file")

	rootCmd.AddCommand(initCmd)

}
