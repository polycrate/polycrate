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
	"fmt"

	"github.com/gosimple/hashdir"
	"github.com/spf13/cobra"
)

var checksumCmd = &cobra.Command{
	Use:   "checksum",
	Short: "Return MD5 Checksum of directory",
	Long:  `Return MD5 Checksum of directory`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]

		tx := polycrate.Transaction()
		tx.SetCommand(cmd)
		defer tx.Stop()

		// Compute block directory hash
		checksum, err := hashdir.Make(path, "md5")
		if err != nil {
			tx.Log.Fatal(err)
		}

		fmt.Println(checksum)

	},
}

func init() {
	rootCmd.AddCommand(checksumCmd)
}
