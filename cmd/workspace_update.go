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
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var workspaceUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update workspace dependencies",
	Long:  ``,
	Args:  cobra.ExactArgs(0),
	//Args:  cobra.ExactArgs(1), // https://github.com/spf13/cobra/blob/master/user_guide.md
	RunE: func(cmd *cobra.Command, args []string) error {
		// if len(args) == 0 {
		// 	log.WithFields(log.Fields{
		// 		"workspace": workspace.Name,
		// 	}).Fatalf("No blocks given")
		// }
		ctx, cancelFunc := context.WithCancel(context.Background())
		ctx, err := polycrate.StartTransaction(ctx, cancelFunc)
		if err != nil {
			log.Fatal(err)
		}

		log := polycrate.GetContextLogger(ctx)

		workspace, err := polycrate.LoadWorkspace(ctx, cmd.Flags().Lookup("workspace").Value.String())
		if err != nil {
			log.Fatal(err)
		}

		log = log.WithField("workspace", workspace.Name)
		ctx = polycrate.SetContextLogger(ctx, log)

		err = workspace.UpdateBlocks(ctx, workspace.Dependencies)
		if err != nil {
			log.Fatal(err)
		}

		if err := polycrate.StopTransaction(ctx, cancelFunc); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	workspaceCmd.AddCommand(workspaceUpdateCmd)
}
