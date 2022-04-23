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
	goErrors "errors"
	"os"
	"strings"

	"github.com/InVisionApp/conjungo"
	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var pluginCallExitCode int

// installCmd represents the install command
var blocksCmd = &cobra.Command{
	Use:   "blocks",
	Short: "Control Polycrate Blocks",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		// loadStatefile()
		//discoverKubernetesDistro()
		saveRuntimeStackfile()

		// TODO: verifyPlugin(plugin, pluginCommand)
		var err error
		pluginCallExitCode, err = callPlugin(blockName, actionName)
		CheckErr(err)
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		if err := loadWorkspace(); err != nil {
			log.Fatal(err)
		}
		blockName = args[0]
		actionName = args[1]
	},
	// PersistentPreRun: func(cmd *cobra.Command, args []string) {
	// 	if err := loadStatefile(); err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	addHistoryItem(cmd, "in progress")
	// },
	// PersistentPostRun: func(cmd *cobra.Command, args []string) {
	// 	updateHistoryItemStatus(strconv.Itoa(pluginCallExitCode))
	// 	writeHistory()
	// },
}

func init() {
	rootCmd.AddCommand(blocksCmd)
}

type Block struct {
	Metadata Metadata               `mapstructure:"metadata" json:"metadata" validate:"required"`
	Actions  []Action               `mapstructure:"actions,omitempty" json:"actions,omitempty"`
	Config   map[string]interface{} `mapstructure:"config,omitempty,remain" json:"config,omitempty,remain"`
	From     string                 `mapstructure:"from,omitempty" json:"from,omitempty"`
	Template bool                   `mapstructure:"template,omitempty" json:"template,omitempty"`
	Version  string                 `mapstructure:"version" json:"version"`
	resolved bool
	parent   *Block
}

func (c *Block) Validate() error {
	validate := validator.New()

	err := validate.Struct(c)

	if err != nil {

		// this check is only needed when your code could produce
		// an invalid value for validation such as interface with nil
		// value most including myself do not usually have code like this.
		if _, ok := err.(*validator.InvalidValidationError); ok {
			log.Error(err)
			return nil
		}

		for _, err := range err.(validator.ValidationErrors) {
			log.Error("Configuration option `" + strings.ToLower(err.Namespace()) + "` failed to validate: " + err.Tag())
		}

		// from here you can create your own error messages in whatever language you wish
		return goErrors.New("Error validating Block")
	}

	// if _, err := os.Stat(blockDir); os.IsNotExist(err) {
	// 	return goErrors.New("Block not found at: " + blockDir)
	// }
	// log.Debug("Found Block at " + blockDir)

	return nil
}

func (c *Block) _Validate() error {
	if _, err := os.Stat(blockDir); os.IsNotExist(err) {
		return goErrors.New("Block not found at: " + blockDir)
	}
	log.Debug("Found Block at " + blockDir)

	return nil
}

func (c *Block) MergeIn(block *Block) error {
	opts := conjungo.NewOptions()
	opts.Overwrite = false // do not overwrite existing values in workspaceConfig
	if err := conjungo.Merge(c, block, opts); err != nil {
		return err
	}
	return nil
}

func (c *Block) Inspect() {
	printObject(c)
}

// func (c *Block) Download(pluginName string) error {
// 	pluginDir := filepath.Join(workspaceDir, "plugins", pluginName)
// 	_, err := os.Stat(pluginDir)
// 	if os.IsNotExist(err) || (!os.IsNotExist(err) && force) {

// 		if c.Source.Type != "" {
// 			if c.Source.Type == "git" {
// 				downloadDirName := strings.Join([]string{"cloudstack", "plugin", pluginName}, "-")
// 				downloadDir := filepath.Join("/tmp", downloadDirName)
// 				log.Debug("Downloading to " + downloadDir)

// 				_, err := os.Stat(downloadDir)
// 				if os.IsNotExist(err) {
// 					log.Debug("Temporary download directory does not exist. Cloning from git repository " + c.Source.Git.Repository)
// 					err := cloneRepository(c.Source.Git.Repository, downloadDir, c.Source.Git.Branch, c.Source.Git.Tag)
// 					if err != nil {
// 						return err
// 					}
// 				} else {
// 					log.Debug("Temporary download directory already exists")
// 				}

// 				// Remove .git directory
// 				if c.Source.Git.Path == "" || c.Source.Git.Path == "/" {
// 					pluginGitDirectory := filepath.Join(downloadDir, ".git")

// 					if _, err := os.Stat(pluginGitDirectory); !os.IsNotExist(err) {
// 						var args = []string{"-rf", pluginGitDirectory}
// 						log.Debug("Running command: rm ", strings.Join(args, " "))
// 						err = exec.Command("rm", args...).Run()
// 						//log.Debug("Removed .git directory at " + pluginGitDirectory)
// 						CheckErr(err)
// 					}
// 				}

// 				// Create plugin dir
// 				log.Debug("Attempting to create plugin directory at " + pluginDir)
// 				err = os.MkdirAll(pluginDir, os.ModePerm)
// 				if err != nil {
// 					return err
// 				}

// 				// Copy contents of dependency.path to /context/dependency.name
// 				copyArgs := []string{"-r", filepath.Join(downloadDir, c.Source.Git.Path) + "/.", pluginDir}
// 				log.Debug("Running command: cp ", strings.Join(copyArgs, " "))
// 				_, err = exec.Command("cp", copyArgs...).Output()
// 				if err != nil {
// 					return err
// 				}

// 				// Delete downloadDir
// 				err = os.RemoveAll(downloadDir)
// 				if err != nil {
// 					return err
// 				}

// 				log.Info("Successfully downloaded ", pluginName)
// 			}
// 		}
// 	} else {
// 		log.Warn("Plugin ", pluginName, " already exists. Use --force to download anyways. CAUTION: this will override your current plugin directory")
// 	}
// 	return nil
// }
