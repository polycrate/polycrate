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

// func newWorkspaceInspectCmd(args []string) *cobra.Command {
// 	cmd := &cobra.Command{
// 		Use:   "inspect",
// 		Short: "Inspect the workspace",
// 		Long:  ``,
// 		Args:  cobra.ExactArgs(0), // https://github.com/spf13/cobra/blob/master/user_guide.md
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			ctx, _ := context.WithCancel(context.Background())
// 			ctx, err := polycrate.StartTransaction(ctx)
// 			if err != nil {
// 				log.Fatal(err)
// 			}

// 			log := polycrate.GetContextLogger(ctx)

// 			workspace, err := polycrate.LoadWorkspace(ctx, cmd.Flags().Lookup("workspace").Value.String())
// 			if err != nil {
// 				log.Fatal(err)
// 			}

// 			log = log.WithField("workspace", workspace.Name)
// 			ctx = polycrate.SetContextLogger(ctx, log)

// 			workspace.Inspect(ctx)
// 			return nil
// 		},
// 	}

// 	return cmd
// }

// installCmd represents the install command
var transactionCmd = &cobra.Command{
	Use:    "transaction",
	Hidden: true,
	Short:  "Interact with Transactions",
	Long:   ``,
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

		workspace.Inspect(tx.Context)
	},
}

func init() {
	rootCmd.AddCommand(transactionCmd)
}
