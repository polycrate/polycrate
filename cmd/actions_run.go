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
var actionsRunCmd = &cobra.Command{
	Use:   "run 'block.action'",
	Short: "Run an Action",
	Long: `
To run an Action, use this command with 1 argument - the Action address.
The action address is a combination of the Block name and the Action name, joined with a dot/period.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			log.Fatal("Need exactly one argument: Action address (e.g. 'Block.Action')")
		}

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

		if err := workspace.RunAction(ctx, args[0]); err != nil {
			polycrate.ContextExit(ctx, cancelFunc, err)
		}

		polycrate.ContextExit(ctx, cancelFunc, nil)

	},
}

func init() {
	actionsCmd.AddCommand(actionsRunCmd)
}
