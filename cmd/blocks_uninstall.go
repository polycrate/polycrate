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
var blocksUninstallCmd = &cobra.Command{
	Use:   "uninstall BLOCK_NAME",
	Short: "Uninstall a block",
	Long:  ``,
	Args:  cobra.ExactArgs(1), // https://github.com/spf13/cobra/blob/master/user_guide.md
	Run: func(cmd *cobra.Command, args []string) {
		blockName := args[0]
		workspace.load().Flush()

		block := workspace.GetBlockFromIndex(blockName)
		if block != nil {
			log.WithFields(log.Fields{
				"workspace": workspace.Name,
				"block":     block.Name,
			}).Warnf("Uninstalling block")

			block.Uninstall(pruneBlock)
		} else {
			log.WithFields(log.Fields{
				"workspace": workspace.Name,
				"block":     blockName,
			}).Fatalf("Block not found")
		}
	},
}

func init() {
	blocksCmd.AddCommand(blocksUninstallCmd)
}
