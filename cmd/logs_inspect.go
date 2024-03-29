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

var logsInspectCmd = &cobra.Command{
	Use:    "inspect",
	Hidden: true,
	Short:  "Show workspace log",
	Long:   `Show workspace log`,
	Args:   cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		_w := cmd.Flags().Lookup("workspace").Value.String()

		tx := polycrate.Transaction().SetCommand(cmd).SetJob(func(tx *PolycrateTransaction) error {
			workspace, err := polycrate.LoadWorkspace(tx, _w, true)
			if err != nil {
				tx.Log.Fatal(err)
			}

			log, err := workspace.GetLog(args[0])
			if err != nil {
				return err
			}
			log.Inspect()
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
	logsCmd.AddCommand(logsInspectCmd)
}
