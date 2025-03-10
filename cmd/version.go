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

var short bool
var latest bool

// installCmd represents the install command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version info",
	Long:  `Show version info`,
	Args:  cobra.ExactArgs(0), // https://github.com/spf13/cobra/blob/master/user_guide.md
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version)

		if !short {
			if commit != "" && date != "" {
				fmt.Printf("Commit %s from %s\n", commit, date)
			}
		}

		if latest {
			latest_stable_version, err := polycrate.GetStableVersion(context.TODO())
			if err == nil {
				fmt.Printf("Latest stable version: %s", latest_stable_version)
			}
		}
	},
}

func init() {

	versionCmd.Flags().BoolVarP(&latest, "latest", "", false, "Get the latest version, too")
	versionCmd.Flags().BoolVarP(&short, "short", "", false, "Only dump the snapshot, do not run anything")
	rootCmd.AddCommand(versionCmd)
}
