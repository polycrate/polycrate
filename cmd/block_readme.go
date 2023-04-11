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
	"fmt"

	"github.com/spf13/cobra"
)

// installCmd represents the install command
var blocksReadmeCmd = &cobra.Command{
	Use:   "readme BLOCK_NAME",
	Short: "Show the README of a block",
	Long:  ``,
	Args:  cobra.ExactArgs(1), // https://github.com/spf13/cobra/blob/master/user_guide.md
	Run: func(cmd *cobra.Command, args []string) {
		_w := cmd.Flags().Lookup("workspace").Value.String()

		tx := polycrate.Transaction()
		tx.SetCommand(cmd)
		defer tx.Stop()

		workspace, err := polycrate.LoadWorkspace(tx, _w, true)
		if err != nil {
			tx.Log.Fatal(err)
		}

		block, err := workspace.GetBlock(args[0])
		if err != nil {
			tx.Log.Fatal(err)
		}

		if block != nil {
			err := block.README()
			if err != nil {
				tx.Log.Fatal(err)
			}
		} else {
			err := fmt.Errorf("Block not found: %s", args[0])
			tx.Log.Fatal(err)
		}
	},
}

func init() {
	blocksCmd.AddCommand(blocksReadmeCmd)
}
