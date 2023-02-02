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

	"github.com/spf13/cobra"
)

var loadCmd = &cobra.Command{
	Use:    "load",
	Hidden: true,
	Short:  "Show Workspace",
	Long:   `Show Workspace`,
	Run: func(cmd *cobra.Command, args []string) {
		_w := cmd.Flags().Lookup("workspace").Value.String()

		ctx := context.Background()
		ctx, cancel, err := polycrate.NewTransaction(ctx, cmd)
		defer polycrate.StopTransaction(ctx, cancel)

		log := polycrate.GetContextLogger(ctx)

		ctx, workspace, err := polycrate.GetWorkspaceWithContext(ctx, _w, true)
		if err != nil {
			log.Fatal(err)
		}

		workspace.Inspect(ctx)
	},
}

func init() {
	rootCmd.AddCommand(loadCmd)
}
