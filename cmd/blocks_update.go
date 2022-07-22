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
var blocksUpdateCmd = &cobra.Command{
	Use:   "update BLOCK_NAME",
	Short: "Update a block",
	Long:  ``,
	//Args:  cobra.ExactArgs(1), // https://github.com/spf13/cobra/blob/master/user_guide.md
	Run: func(cmd *cobra.Command, args []string) {
		workspace.load().Flush()
		if len(args) == 0 {
			log.WithFields(log.Fields{
				"workspace": workspace.Name,
			}).Fatalf("No blocks given")
		}

		err := workspace.UpdateBlocks(args)
		if err != nil {
			log.WithFields(log.Fields{
				"workspace": workspace.Name,
			}).Fatal(err)
		}
	},
}

func init() {
	blocksCmd.AddCommand(blocksUpdateCmd)
}