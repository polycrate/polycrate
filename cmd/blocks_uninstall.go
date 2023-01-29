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
	"fmt"

	"github.com/spf13/cobra"
)

// installCmd represents the install command
var blocksUninstallCmd = &cobra.Command{
	Use:    "uninstall BLOCK1 BLOCK2",
	Short:  "Uninstall blocks",
	Hidden: true,
	Long:   ``,
	//Args:  cobra.ExactArgs(1), // https://github.com/spf13/cobra/blob/master/user_guide.md
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancelFunc := context.WithCancel(context.Background())
		ctx, err := polycrate.StartTransaction(ctx, cancelFunc)
		if err != nil {
			polycrate.ContextExit(ctx, cancelFunc, err)
		}

		log := polycrate.GetContextLogger(ctx)

		workspace, err := polycrate.LoadWorkspace(ctx, cmd.Flags().Lookup("workspace").Value.String())
		if err != nil {
			polycrate.ContextExit(ctx, cancelFunc, err)
		}

		log = log.WithField("workspace", workspace.Name)
		ctx = polycrate.SetContextLogger(ctx, log)

		if len(args) == 0 {
			err := fmt.Errorf("no blocks given")
			polycrate.ContextExit(ctx, cancelFunc, err)
		}

		err = workspace.UninstallBlocks(ctx, args)
		if err != nil {
			polycrate.ContextExit(ctx, cancelFunc, err)
		}

		polycrate.ContextExit(ctx, cancelFunc, nil)
	},
}

func init() {
	blocksCmd.AddCommand(blocksUninstallCmd)
}
