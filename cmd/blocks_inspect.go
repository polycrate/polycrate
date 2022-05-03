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
var blocksInspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Inspect a Block",
	Long:  ``,
	Args:  cobra.ExactArgs(1), // https://github.com/spf13/cobra/blob/master/user_guide.md
	Run: func(cmd *cobra.Command, args []string) {
		workspace.load()
		if workspace.Flush() != nil {
			log.Fatal(workspace.Flush)
		}

		block := workspace.GetBlockFromIndex(args[0])
		if block != nil {
			block.Inspect()
		} else {
			log.Fatalf("Block not found: %s", args[0])
		}
	},
}

func init() {
	blocksCmd.AddCommand(blocksInspectCmd)
}
