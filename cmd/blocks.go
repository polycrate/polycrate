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

	"github.com/go-playground/validator/v10"
	"github.com/imdario/mergo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var pruneBlock bool

// installCmd represents the install command
var blocksCmd = &cobra.Command{
	Use:   "block",
	Short: "Control Polycrate Blocks",
	Aliases: []string{
		"blocks",
	},
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		workspace.load().Flush()
		workspace.ListBlocks().Flush()
	},
}

func init() {
	rootCmd.AddCommand(blocksCmd)
	blocksCmd.PersistentFlags().BoolVar(&pruneBlock, "prune", false, "Prune artifacts")
	blocksCmd.PersistentFlags().StringVarP(&blockVersion, "version", "v", "latest", "Version of the block")
}

type BlockWorkdir struct {
	exists        bool
	LocalPath     string `yaml:"localpath,omitempty" mapstructure:"localpath,omitempty" json:"localpath,omitempty"`
	ContainerPath string `yaml:"containerpath,omitempty" mapstructure:"containerpath,omitempty" json:"containerpath,omitempty"`
}
type BlockKubeconfig struct {
	From          string `yaml:"from,omitempty" mapstructure:"from,omitempty" json:"from,omitempty"`
	exists        bool
	LocalPath     string `yaml:"localpath,omitempty" mapstructure:"localpath,omitempty" json:"localpath,omitempty"`
	ContainerPath string `yaml:"containerpath,omitempty" mapstructure:"containerpath,omitempty" json:"containerpath,omitempty"`
}
type BlockInventory struct {
	From          string `yaml:"from,omitempty" mapstructure:"from,omitempty" json:"from,omitempty"`
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
	Config      map[string]interface{} `yaml:"config,omitempty" mapstructure:"config,omitempty" json:"config,omitempty"`
	From        string                 `yaml:"from,omitempty" mapstructure:"from,omitempty" json:"from,omitempty"`
	Template    bool                   `yaml:"template,omitempty" mapstructure:"template,omitempty" json:"template,omitempty"`
	Version     string                 `yaml:"version,omitempty" mapstructure:"version" json:"version"`
	resolved    bool
	Workdir     BlockWorkdir    `yaml:"workdir,omitempty" mapstructure:"workdir,omitempty" json:"workdir,omitempty"`
	Inventory   BlockInventory  `yaml:"inventory,omitempty" mapstructure:"inventory,omitempty" json:"inventory,omitempty"`
	Kubeconfig  BlockKubeconfig `yaml:"kubeconfig,omitempty" mapstructure:"kubeconfig,omitempty" json:"kubeconfig,omitempty"`
	Artifacts   BlockArtifacts  `yaml:"artifacts,omitempty" mapstructure:"artifacts,omitempty" json:"artifacts,omitempty"`
	address     string
	//err         error
}

func (c *Block) getInventoryPath() string {
	if c.Inventory.From != "" {
		// Take the inventory from another Block
		inventorySourceBlock := workspace.getBlockByName(c.Inventory.From)
		if inventorySourceBlock != nil {
			if inventorySourceBlock.Inventory.exists {
				if local {
					return inventorySourceBlock.Inventory.LocalPath
				} else {
					return inventorySourceBlock.Inventory.ContainerPath
				}
			}
		}
	} else {
		if c.Inventory.exists {
			if local {
				return c.Inventory.LocalPath
			} else {
				return c.Inventory.ContainerPath
			}
		}
	}
	return ""
}

func (c *Block) getKubeconfigPath() string {
	if c.Kubeconfig.From != "" {
		// Take the inventory from another Block
		kubeconfigSourceBlock := workspace.getBlockByName(c.Kubeconfig.From)
		if kubeconfigSourceBlock != nil {
			if kubeconfigSourceBlock.Kubeconfig.exists {
				if local {
					return kubeconfigSourceBlock.Kubeconfig.LocalPath
				} else {
					return kubeconfigSourceBlock.Kubeconfig.ContainerPath
				}
			}
		}
	} else {
		if c.Kubeconfig.exists {
			if local {
				return c.Kubeconfig.LocalPath
			} else {
				return c.Kubeconfig.ContainerPath
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

func (c *Block) validate() error {
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

func (c *Block) MergeIn(block Block) error {
	// if err := mergo.Merge(c, block); err != nil {
	// 	log.Fatal(err)
	// }
	// return nil

	// Name
	if block.Name != "" && c.Name == "" {
		c.Name = block.Name
	}

	// Description
	if block.Description != "" && c.Description == "" {
		c.Description = block.Description
	}

	// Labels
	if block.Labels != nil {
		if err := mergo.Merge(&c.Labels, block.Labels); err != nil {
			return err
		}
	}

	// Alias
	if block.Alias != nil {
		if err := mergo.Merge(&c.Alias, block.Alias); err != nil {
			return err
		}
	}

	// Actions
	if block.Actions != nil {
		for _, sourceAction := range block.Actions {
			// get the corresponding action in the existing block
			destinationAction := c.getActionByName(sourceAction.Name)

			if destinationAction != nil {
				if err := mergo.Merge(destinationAction, sourceAction); err != nil {
					return err
				}

				destinationAction.address = strings.Join([]string{c.Name, sourceAction.Name}, ".")
				destinationAction.Block = c.Name
			} else {
				sourceAction.address = strings.Join([]string{c.Name, sourceAction.Name}, ".")
				sourceAction.Block = c.Name
				c.Actions = append(c.Actions, sourceAction)
			}
		}
	}

	// Config
	if block.Config != nil {
		if err := mergo.Merge(&c.Config, block.Config); err != nil {
			return err
		}
	}

	// From
	if block.From != "" && c.From == "" {
		c.From = block.From
	}

	// Template
	if !block.Template {
		c.Template = block.Template
	}

	// Version
	if block.Version != "" && c.Version == "" {
		c.Version = block.Version
	}

	// Workdir
	if (block.Workdir != BlockWorkdir{} && c.Workdir == BlockWorkdir{}) {
		if err := mergo.Merge(&c.Workdir, block.Workdir); err != nil {
			return err
		}
	}

	// Inventory
	if (block.Inventory != BlockInventory{} && c.Inventory == BlockInventory{}) {
		if err := mergo.Merge(&c.Inventory, block.Inventory); err != nil {
			return err
		}
	}

	// Kubeconfig
	if (block.Kubeconfig != BlockKubeconfig{} && c.Kubeconfig == BlockKubeconfig{}) {
		if err := mergo.Merge(&c.Kubeconfig, block.Kubeconfig); err != nil {
			return err
		}
	}

	// Artifacts
	if (block.Artifacts != BlockArtifacts{} && c.Artifacts == BlockArtifacts{}) {
		if err := mergo.Merge(&c.Artifacts, block.Artifacts); err != nil {
			return err
		}
	}

	return nil
}

func (c *Block) Inspect() {
	printObject(c)
}

func (c *Block) LoadInventory() {
	// Locate "inventory.yml" in blockArtifactsDir
	blockInventoryFile := filepath.Join(c.Artifacts.LocalPath, "inventory.yml")
	log.WithFields(log.Fields{
		"path":      blockInventoryFile,
		"block":     c.Name,
		"workspace": workspace.Name,
	}).Debugf("Loading inventory")

	if _, err := os.Stat(blockInventoryFile); !os.IsNotExist(err) {
		// File exists
		c.Inventory.exists = true
		c.Inventory.LocalPath = blockInventoryFile
		c.Inventory.ContainerPath = filepath.Join(c.Artifacts.ContainerPath, "inventory.yml")
		log.WithFields(log.Fields{
			"path":      blockInventoryFile,
			"block":     c.Name,
			"workspace": workspace.Name,
		}).Debugf("Found block inventory")
	} else {
		c.Inventory.exists = false
	}

}

func (c *Block) LoadKubeconfig() {
	// Locate "kubeconfig.yml" in blockArtifactsDir
	blockKubeconfigFile := filepath.Join(c.Artifacts.LocalPath, "kubeconfig.yml")

	if _, err := os.Stat(blockKubeconfigFile); !os.IsNotExist(err) {
		// File exists
		c.Kubeconfig.exists = true
		c.Kubeconfig.LocalPath = blockKubeconfigFile
		c.Kubeconfig.ContainerPath = filepath.Join(c.Artifacts.ContainerPath, "kubeconfig.yml")
		log.Debug("Found Block Kubeconfig at " + blockKubeconfigFile)
	} else {
		c.Kubeconfig.exists = false
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

func (c *Block) Uninstall(prune bool) error {
	log.WithFields(log.Fields{
		"workspace": c.Name,
		"block":     c.Name,
		"version":   c.Version,
	}).Debugf("Successfully uninstalled block from workspace")

	// e.g. $HOME/.polycrate/workspaces/workspace-1/artifacts/blocks/block-1
	if _, err := os.Stat(c.Workdir.LocalPath); os.IsNotExist(err) {
		log.WithFields(log.Fields{
			"workspace": workspace.Name,
			"block":     c.Name,
			"path":      c.Workdir.LocalPath,
		}).Debugf("Block directory does not exist")
	} else {
		err := os.RemoveAll(c.Workdir.LocalPath)
		if err != nil {
			return err
		}
		log.WithFields(log.Fields{
			"workspace": workspace.Name,
			"block":     c.Name,
			"path":      c.Workdir.LocalPath,
		}).Debugf("Block directory removed")

		if prune {
			log.WithFields(log.Fields{
				"workspace": workspace.Name,
				"block":     c.Name,
				"path":      c.Workdir.LocalPath,
			}).Debugf("Pruning artifacts")

			err := os.RemoveAll(c.Artifacts.LocalPath)
			if err != nil {
				return err
			}
			log.WithFields(log.Fields{
				"workspace": workspace.Name,
				"block":     c.Name,
				"path":      c.Artifacts.LocalPath,
			}).Debugf("Block artifacts directory removed")
		}
	}
	log.WithFields(log.Fields{
		"workspace": c.Name,
		"block":     c.Name,
		"version":   c.Version,
	}).Debugf("Successfully uninstalled block from workspace")
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
