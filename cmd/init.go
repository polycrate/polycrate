/*
Copyright © 2021 Fabian Peter <fp@ayedo.de>

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
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var fileUrl string

// installCmd represents the install command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initalize a stack",
	Long:  `Initalize a stack`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create context directory if not exists
		contextDir := workspaceDir
		err := os.MkdirAll(contextDir, os.ModePerm)
		CheckErr(err)

		pluginsDir := filepath.Join(workspaceDir, "plugins")
		err = os.MkdirAll(pluginsDir, os.ModePerm)
		CheckErr(err)

		// Create Stackfile if not exists
		configPath := getConfigPath(workspaceDir)
		if configPath == "" {
			// Load from basic example
			// download to /tmp-asdasdasd
			// load into viper var
			// override values
			// save to Stackfile
			// remove tempfile
			donwloadTempPath := filepath.Join("/tmp", "cloudstack-example-stackfile")
			err := DownloadFile(fileUrl, donwloadTempPath)
			exampleConfig := viper.New()

			// Check overrides
			if len(overrides) > 0 {
				for _, override := range overrides {
					// Split string by =
					kv := strings.Split(override, "=")

					// Override property
					log.Debug("Setting " + kv[0] + " to " + kv[1])
					exampleConfig.Set(kv[0], kv[1])
				}
			}
			exampleConfig.SetConfigType("yaml")
			exampleConfig.SetConfigFile(donwloadTempPath)
			err = exampleConfig.MergeInConfig()
			CheckErr(err)

			exampleConfig.WriteConfigAs(workspaceDir + "/Stackfile")
			log.Info("Created config from " + fileUrl + " at " + workspaceDir + "/Stackfile")

			err = os.Remove(donwloadTempPath)
			CheckErr(err)
		}

		// Create SSH Keys if not exists
		CreateSSHKeyCmd.Run(cmd, []string{})
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVar(&fileUrl, "template", "https://gitlab.com/ayedocloudsolutions/cloudstack/cloudstack/-/raw/main/examples/hcloud-single-node/Stackfile", "Stackfile to init from")
}
