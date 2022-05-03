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
	"path/filepath"
	"strings"

	"github.com/InVisionApp/conjungo"
	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var blocksCmd = &cobra.Command{
	Use:   "block",
	Short: "Control Polycrate Blocks",
	Aliases: []string{
		"blocks",
	},
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		workspace.load()
		if workspace.Flush() != nil {
			log.Fatal(workspace.Flush)
		}
		workspace.ListBlocks()
	},
}

func init() {
	rootCmd.AddCommand(blocksCmd)
}

type BlockWorkdir struct {
	exists        bool
	LocalPath     string `yaml:"localpath,omitempty" mapstructure:"localpath,omitempty" json:"localpath,omitempty"`
	ContainerPath string `yaml:"containerpath,omitempty" mapstructure:"containerpath,omitempty" json:"containerpath,omitempty"`
}
type BlockKubeconfig struct {
	from          string
	exists        bool
	LocalPath     string `yaml:"localpath,omitempty" mapstructure:"localpath,omitempty" json:"localpath,omitempty"`
	ContainerPath string `yaml:"containerpath,omitempty" mapstructure:"containerpath,omitempty" json:"containerpath,omitempty"`
}
type BlockInventory struct {
	from          string
	exists        bool
	LocalPath     string `yaml:"localpath,omitempty" mapstructure:"localpath,omitempty" json:"localpath,omitempty"`
	ContainerPath string `yaml:"containerpath,omitempty" mapstructure:"containerpath,omitempty" json:"containerpath,omitempty"`
}
type BlockArtifacts struct {
	LocalPath     string `yaml:"localpath,omitempty" mapstructure:"localpath,omitempty" json:"localpath,omitempty"`
	ContainerPath string `yaml:"containerpath,omitempty" mapstructure:"containerpath,omitempty" json:"containerpath,omitempty"`
}

type Block struct {
	Name        string                 `yaml:"name,omitempty" mapstructure:"name,omitempty" json:"name,omitempty" validate:"required,metadata_name"`
	Description string                 `yaml:"description,omitempty" mapstructure:"description,omitempty" json:"description,omitempty"`
	Labels      map[string]string      `yaml:"labels,omitempty" mapstructure:"labels,omitempty" json:"labels,omitempty"`
	Alias       []string               `yaml:"alias,omitempty" mapstructure:"alias,omitempty" json:"alias,omitempty"`
	Actions     []Action               `yaml:"actions,omitempty" mapstructure:"actions,omitempty" json:"actions,omitempty"`
	Config      map[string]interface{} `yaml:"config,omitempty" mapstructure:"config,omitempty,remain" json:"config,omitempty"`
	From        string                 `yaml:"from,omitempty" mapstructure:"from,omitempty" json:"from,omitempty"`
	Template    bool                   `yaml:"template,omitempty" mapstructure:"template,omitempty" json:"template,omitempty"`
	Version     string                 `yaml:"version,omitempty" mapstructure:"version" json:"version"`
	resolved    bool
	Parent      *Block
	Workdir     BlockWorkdir `yaml:"workdir,omitempty" mapstructure:"workdir,omitempty" json:"workdir,omitempty"`
	inventory   BlockInventory
	kubeconfig  BlockKubeconfig
	Artifacts   BlockArtifacts `yaml:"artifacts,omitempty" mapstructure:"artifacts,omitempty" json:"artifacts,omitempty"`
	address     string
	//err         error
}

func (c *Block) getInventoryPath() string {
	if c.inventory.from != "" {
		// Take the inventory from another Block
		inventorySourceBlock := workspace.getBlockByName(c.inventory.from)
		if inventorySourceBlock != nil {
			if inventorySourceBlock.inventory.exists {
				if local {
					return inventorySourceBlock.inventory.LocalPath
				} else {
					return inventorySourceBlock.inventory.ContainerPath
				}
			}
		}
	} else {
		if c.inventory.exists {
			if local {
				return c.inventory.LocalPath
			} else {
				return c.inventory.ContainerPath
			}
		}
	}
	return ""
}

func (c *Block) getKubeconfigPath() string {
	if c.kubeconfig.from != "" {
		// Take the inventory from another Block
		kubeconfigSourceBlock := workspace.getBlockByName(c.kubeconfig.from)
		if kubeconfigSourceBlock != nil {
			if kubeconfigSourceBlock.kubeconfig.exists {
				if local {
					return kubeconfigSourceBlock.kubeconfig.LocalPath
				} else {
					return kubeconfigSourceBlock.kubeconfig.ContainerPath
				}
			}
		}
	} else {
		if c.kubeconfig.exists {
			if local {
				return c.kubeconfig.LocalPath
			} else {
				return c.kubeconfig.ContainerPath
			}
		}
	}
	return ""
}

func (c *Block) getActionByName(actionName string) *Action {

	//for _, block := range c.Blocks {
	for i := 0; i < len(c.Actions); i++ {
		action := &c.Actions[i]
		if action.Name == actionName {
			return action
		}
	}
	return nil
}

func (c *Block) Validate() error {
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
		return goErrors.New("error validating Block")
	}

	// if _, err := os.Stat(blockDir); os.IsNotExist(err) {
	// 	return goErrors.New("Block not found at: " + blockDir)
	// }
	// log.Debug("Found Block at " + blockDir)

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

func (c *Block) LoadInventory() {
	// Locate "inventory.json" in blockArtifactsDir
	blockInventoryFile := filepath.Join(c.Artifacts.LocalPath, "inventory.json")

	if _, err := os.Stat(blockInventoryFile); !os.IsNotExist(err) {
		// File exists
		c.inventory.exists = true
		c.inventory.LocalPath = blockInventoryFile
		c.inventory.ContainerPath = filepath.Join(c.Artifacts.ContainerPath, "inventory.json")
		log.Debug("Found Block Inventory at " + blockInventoryFile)
	} else {
		c.inventory.exists = false
	}

}

func (c *Block) LoadKubeconfig() {
	// Locate "kubeconfig.yml" in blockArtifactsDir
	blockKubeconfigFile := filepath.Join(c.Artifacts.LocalPath, "kubeconfig.yml")

	if _, err := os.Stat(blockKubeconfigFile); !os.IsNotExist(err) {
		// File exists
		c.kubeconfig.exists = true
		c.kubeconfig.LocalPath = blockKubeconfigFile
		c.kubeconfig.ContainerPath = filepath.Join(c.Artifacts.ContainerPath, "kubeconfig.yml")
		log.Debug("Found Block Kubeconfig at " + blockKubeconfigFile)
	} else {
		c.kubeconfig.exists = false
	}

}

func (c *Block) LoadArtifacts() error {
	// e.g. $HOME/.polycrate/workspaces/workspace-1/artifacts/blocks/block-1
	c.Artifacts.LocalPath = filepath.Join(workspace.Path, workspace.Config.ArtifactsRoot, workspace.Config.BlocksRoot, c.Name)
	// e.g. /workspace/artifacts/blocks/block-1
	c.Artifacts.ContainerPath = filepath.Join(workspace.ContainerPath, workspace.Config.ArtifactsRoot, workspace.Config.BlocksRoot, c.Name)

	// Check if the local artifacts directory for this Block exists
	if _, err := os.Stat(c.Artifacts.LocalPath); os.IsNotExist(err) {
		// Directory does not exist
		// We create it
		if err := os.MkdirAll(c.Artifacts.LocalPath, os.ModePerm); err != nil {
			return err
		}
		log.Debugf("Created artifacts directory for block %s at %s", c.Name, c.Artifacts.LocalPath)
	}
	return nil

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

// 				// Copy contents of dependency.Path to /context/dependency.name
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
