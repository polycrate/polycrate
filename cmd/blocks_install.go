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
	"github.com/spf13/viper"
)

// installCmd represents the install command
var pluginsInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Plugins",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		loadWorkspace()
		if len(args) == 0 {
			for _, plugin := range viper.GetStringSlice("stack.plugins") {
				blocksCmd.Run(cmd, []string{plugin, "install"})
			}
		}
		// log.Info("Installing Cloudstack Plugins")
		// log.Info("Using Kubeconfig at ", kubeconfig)
		// loadStackfile()
		// pluginCommand = "install"
		// if len(args) == 1 {
		// 	installComponent(args[0])
		// } else {
		// 	installComponents()
		// }
	},
}

func init() {
	blocksCmd.AddCommand(pluginsInstallCmd)
}
