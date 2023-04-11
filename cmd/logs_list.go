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

var logsListCmd = &cobra.Command{
	Use:    "list",
	Hidden: true,
	Short:  "Show workspace logs",
	Long:   `Show workspace logs`,
	Run: func(cmd *cobra.Command, args []string) {
		_w := cmd.Flags().Lookup("workspace").Value.String()

		tx := polycrate.Transaction().SetCommand(cmd).SetJob(func(tx *PolycrateTransaction) error {
			workspace, err := polycrate.LoadWorkspace(tx, _w, true)
			if err != nil {
				tx.Log.Fatal(err)
			}

			workspace.ListLogs(tx)
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
	logsCmd.AddCommand(logsListCmd)
}
