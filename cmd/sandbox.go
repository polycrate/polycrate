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

var sandboxCmd = &cobra.Command{
	Use:   "sandbox",
	Short: "Play with the Polycrate container",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Starting sandbox container at ", workspaceDir)
		runCommand := []string{"/bin/bash"}
		interactive = true
		//bootstrapEnvVars()
		RunContainer(
			workspace.Config.Image.Reference,
			workspace.Config.Image.Version,
			runCommand,
		)

		// if err != nil {
		// 	log.Error("Plugin ", block, " failed with exit code ", exitCode, ": ", err.Error())
		// } else {
		// 	log.Info("Plugin ", block, " succeeded with exit code ", exitCode, ": OK")
		// }
		// Run container with /bin/bash, Stackfile and all Env Vars
		// Means to restructure the callPlugin command
		// new func: getEnvironment(local = false),
		// new func: createScript()
		// new func: getMounts(local = false), respect --dev-dir
		// new func: getPorts(local = false)
		// new func: getWorkdir(local = false)
	},
}

func init() {
	rootCmd.AddCommand(sandboxCmd)
}
