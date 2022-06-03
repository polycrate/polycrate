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
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var blockVersion string
var download bool = false

// installCmd represents the install command
var blocksInstallCmd = &cobra.Command{
	Use:   "install BLOCK1 BLOCK2:0.0.1",
	Short: "Install Blocks",
	Long:  ``,
	//Args: cobra.RangeArgs(1, 2), // https://github.com/spf13/cobra/blob/master/user_guide.md
	Run: func(cmd *cobra.Command, args []string) {
		workspace.load().Flush()
		if len(args) == 0 {
			log.WithFields(log.Fields{
				"workspace": workspace.Name,
			}).Warnf("No blocks given")
			os.Exit(0)
		}

		err := workspace.InstallBlocks(args)
		if err != nil {
			log.Fatal(err)
		}

	},
}

func init() {
	blocksCmd.AddCommand(blocksInstallCmd)

}
