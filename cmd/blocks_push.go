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
var blocksPushCmd = &cobra.Command{
	Use:   "push BLOCK",
	Short: "Upload a block to the registry",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		blockName := args[0]
		_w := cmd.Flags().Lookup("workspace").Value.String()

		ctx := context.Background()
		ctx, _, cancel, err := polycrate.NewTransaction(ctx, cmd)
		defer polycrate.StopTransaction(ctx, cancel)
		if err != nil {
			log.Fatal(err)
		}

		log := polycrate.GetContextLogger(ctx)

		ctx, workspace, err := polycrate.PreloadWorkspaceWithContext(ctx, _w, false)
		if err != nil {
			log.Fatal(err)
		}

		err = workspace.PushBlock(ctx, blockName)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	blocksCmd.AddCommand(blocksPushCmd)
}
