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

// createCmd represents the create command
var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "clone",
	Long:  `Clone an existing cloudstack`,
	Run: func(cmd *cobra.Command, args []string) {
		repository := args[0]
		var path string
		if len(args) > 1 {
			path = args[1]
		} else {
			path = ""
		}
		var branch string
		if len(args) > 2 {
			branch = args[2]
		} else {
			branch = ""
		}
		err := cloneRepository(repository, path, branch, "")
		CheckErr(err)
	},
}

func init() {
	rootCmd.AddCommand(cloneCmd)

}
