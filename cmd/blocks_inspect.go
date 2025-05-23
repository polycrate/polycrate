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

// installCmd represents the install command
var blocksInspectCmd = &cobra.Command{
	Use:   "inspect BLOCK_NAME",
	Short: "Inspect a block",
	Long:  ``,
	Args:  cobra.ExactArgs(1), // https://github.com/spf13/cobra/blob/master/user_guide.md
	Run: func(cmd *cobra.Command, args []string) {
		_w := cmd.Flags().Lookup("workspace").Value.String()

		tx := polycrate.Transaction().SetCommand(cmd).SetJob(func(tx *PolycrateTransaction) error {
			workspace, err := polycrate.LoadWorkspace(tx, _w, true)
			if err != nil {
				return err
			}

			block, err := workspace.GetBlock(args[0])
			if err != nil {
				return err
			}

			if block.Kind == "k8sappinstance" {
				err := block.ValidateConfig()
				if err != nil {
					return err
				}
			}

			block.Inspect()
			return nil
		})

		err := tx.Run()
		if err != nil {
			tx.Log.Fatal(err)
		}
		defer tx.Stop()

	},
}

func init() {
	blocksCmd.AddCommand(blocksInspectCmd)
}
