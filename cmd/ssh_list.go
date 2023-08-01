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

var _formatHosts bool

// installCmd represents the install command
var sshListCmd = &cobra.Command{
	Use:    "list",
	Short:  "list hosts to ssh into",
	Hidden: true,
	Long:   ``,
	Args:   cobra.ExactArgs(0), // https://github.com/spf13/cobra/blob/master/user_guide.md
	Run: func(cmd *cobra.Command, args []string) {
		// Get all inventories from workspace
		// convert all inventories from workspace
		// populate global host cache in $ARTIFACTS_DIR/ssh_hosts.json
		pull = false
		_w := cmd.Flags().Lookup("workspace").Value.String()

		tx := polycrate.Transaction()
		tx.SetCommand(cmd)
		defer tx.Stop()

		workspace, err := polycrate.LoadWorkspace(tx, _w, true)
		if err != nil {
			tx.Log.Fatal(err)
		}

		var block *Block

		if _sshBlock == "" {
			_blocks, err := workspace.GetBlocksWithInventory()
			if err != nil {
				tx.Log.Fatal(err)
			}
			block = _blocks[0]
		} else {
			block, err = workspace.GetBlock(_sshBlock)
			if err != nil {
				tx.Log.Fatal(err)
			}
		}

		if block != nil {
			err := block.SSHList(tx, _refreshHosts, _formatHosts)
			if err != nil {
				tx.Log.Fatal(err)
			}
		} else {
			err := fmt.Errorf("block does not exist: %s", _sshBlock)
			tx.Log.Fatal(err)
		}
	},
}

func init() {
	sshCmd.AddCommand(sshListCmd)

	sshListCmd.PersistentFlags().StringVar(&_sshBlock, "block", "", "Block to list hosts from")
	sshListCmd.PersistentFlags().BoolVar(&_refreshHosts, "refresh", false, "Refresh hosts cache")
	sshListCmd.PersistentFlags().BoolVar(&_formatHosts, "format", false, "Output polycrate commands instead of objects")
}
