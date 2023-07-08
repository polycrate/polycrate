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

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var _inventoryConvertBlock string

// installCmd represents the install command
var inventoryConvertCmd = &cobra.Command{
	Use:    "convert",
	Short:  "Convert inventory to inventory.poly",
	Long:   ``,
	Hidden: true,
	Args:   cobra.ExactArgs(0), // https://github.com/spf13/cobra/blob/master/user_guide.md
	Run: func(cmd *cobra.Command, args []string) {
		_w := cmd.Flags().Lookup("workspace").Value.String()

		tx := polycrate.Transaction()
		tx.SetCommand(cmd)
		defer tx.Stop()

		workspace, err := polycrate.LoadWorkspace(tx, _w, true)
		if err != nil {
			tx.Log.Fatal(err)
		}

		if _inventoryConvertBlock == "" {
			err := fmt.Errorf("no block selected. Use ' --block $BLOCK_NAME' to select an inventory source")
			log.Fatal(err)
		}

		var block *Block
		block, err = workspace.GetBlock(_inventoryConvertBlock)
		if err != nil {
			log.Fatal(err)
		}

		if block != nil {
			err := block.ConvertInventory(tx)
			if err != nil {
				log.Error(err)
			}
		} else {
			err := fmt.Errorf("block does not exist: %s", _inventoryConvertBlock)
			log.Fatal(err)
		}
	},
}

func init() {
	inventoryCmd.AddCommand(inventoryConvertCmd)

	inventoryConvertCmd.PersistentFlags().StringVar(&_inventoryConvertBlock, "block", "inventory", "Block to load inventory from")
}
