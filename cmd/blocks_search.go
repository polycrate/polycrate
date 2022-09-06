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

var blocksSearchCmd = &cobra.Command{
	Use:        "search",
	Short:      "Search blocks in the registry",
	Long:       ``,
	Args:       cobra.ExactArgs(1),
	Deprecated: "please see https://docs.polycrate.io",
	Run: func(cmd *cobra.Command, args []string) {
		// blockName := args[0]
		// blocks, err := config.Registry.SearchBlock(blockName)
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// for _, block := range blocks {

		// 	fmt.Printf("%s (latest: %s)\n", block.Title["rendered"], block.Releases[0].Version)
		// }

	},
}

func init() {
	blocksCmd.AddCommand(blocksSearchCmd)
}
