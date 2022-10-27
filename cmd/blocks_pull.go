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
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var blocksPullCmd = &cobra.Command{
	Use:   "pull BLOCK:VERSION",
	Short: "Pull a block from the registry",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		workspace.load().Flush()
		blockInfo := args[0]

		fullTag, registryUrl, blockName, blockVersion := mapDockerTag(blockInfo)

		// blockName, blockVersion, err := registry.resolveArg(blockInfo)
		// if err != nil {
		// 	log.Fatal(err)
		// }

		err := workspace.PullBlock(fullTag, registryUrl, blockName, blockVersion)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	blocksCmd.AddCommand(blocksPullCmd)
}
