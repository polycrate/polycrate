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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/imdario/mergo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var pruneBlock bool

// installCmd represents the install command
var blocksCmd = &cobra.Command{
	Use:   "block",
	Short: "List blocks in the workspace",
	Aliases: []string{
		"blocks",
	},
	Long: ``,
	Args: cobra.ExactArgs(0),
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
	Path          string `yaml:"path,omitempty" mapstructure:"path,omitempty" json:"path,omitempty"`
	LocalPath     string `yaml:"localpath,omitempty" mapstructure:"localpath,omitempty" json:"localpath,omitempty"`
	ContainerPath string `yaml:"containerpath,omitempty" mapstructure:"containerpath,omitempty" json:"containerpath,omitempty"`
}
type BlockKubeconfig struct {
	Path          string `yaml:"path,omitempty" mapstructure:"path,omitempty" json:"path,omitempty"`
	From          string `yaml:"from,omitempty" mapstructure:"from,omitempty" json:"from,omitempty"`
	Filename      string `yaml:"filename,omitempty" mapstructure:"filename,omitempty" json:"filename,omitempty"`
	exists        bool
	LocalPath     string `yaml:"localpath,omitempty" mapstructure:"localpath,omitempty" json:"localpath,omitempty"`
	ContainerPath string `yaml:"containerpath,omitempty" mapstructure:"containerpath,omitempty" json:"containerpath,omitempty"`
}
type BlockInventory struct {
	Path          string `yaml:"path,omitempty" mapstructure:"path,omitempty" json:"path,omitempty"`
	From          string `yaml:"from,omitempty" mapstructure:"from,omitempty" json:"from,omitempty"`
	Filename      string `yaml:"filename,omitempty" mapstructure:"filename,omitempty" json:"filename,omitempty"`
	exists        bool
	LocalPath     string `yaml:"localpath,omitempty" mapstructure:"localpath,omitempty" json:"localpath,omitempty"`
	ContainerPath string `yaml:"containerpath,omitempty" mapstructure:"containerpath,omitempty" json:"containerpath,omitempty"`
}
type BlockArtifacts struct {
	Path          string `yaml:"path,omitempty" mapstructure:"path,omitempty" json:"path,omitempty"`
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
	err         error
}

func (b *Block) Flush() *Block {
	if b.err != nil {
		log.Fatal(b.err)
	}
	return b
}

func (b *Block) Resolve() *Block {
	if !b.resolved {
		log.WithFields(log.Fields{
			"block":     b.Name,
			"workspace": workspace.Name,
		}).Debugf("Resolving block %s", b.Name)

		// Check if a "from:" stanza is given and not empty
		// This means that the loadedBlock should inherit from another Block
		if b.From != "" {
			// a "from:" stanza is given
			log.WithFields(log.Fields{
				"block":      b.Name,
				"dependency": b.From,
				"workspace":  workspace.Name,
			}).Debugf("Dependency detected")

			// Try to load the referenced Block
			dependency := workspace.getBlockByName(b.From)

			if dependency == nil {
				log.WithFields(log.Fields{
					"block":      b.Name,
					"dependency": b.From,
					"workspace":  workspace.Name,
				}).Errorf("Dependency not found in the workspace")

				b.err = fmt.Errorf("Block '%s' not found in the Workspace. Please check the 'from' stanza of Block", b.From, b.Name)
				return b
			}

			log.WithFields(log.Fields{
				"block":      b.Name,
				"dependency": b.From,
				"workspace":  workspace.Name,
			}).Debugf("Dependency loaded")

			dep := *dependency

			// Check if the dependency Block has already been resolved
			// If not, re-queue the loaded Block so it can be resolved in another iteration
			if !dep.resolved {
				// Needed Block from 'from' stanza is not yet resolved
				log.WithFields(log.Fields{
					"block":               b.Name,
					"dependency":          b.From,
					"workspace":           workspace.Name,
					"dependency_resolved": dep.resolved,
				}).Debugf("Dependency not resolved")
				b.resolved = false
				b.err = DependencyNotResolved
				return b
			}

			// Merge the dependency Block into the loaded Block
			// We do NOT OVERWRITE existing values in the loaded Block because we must assume
			// That this is configuration that has been explicitly set by the user
			// The merge works like "loading defaults" for the loaded Block
			err := b.MergeIn(dep)
			if err != nil {
				b.err = err
				return b
			}

			// Handle Workdir
			b.Workdir.LocalPath = dep.Workdir.LocalPath
			b.Workdir.ContainerPath = dep.Workdir.ContainerPath

			log.WithFields(log.Fields{
				"block":      b.Name,
				"dependency": b.From,
				"workspace":  workspace.Name,
			}).Debugf("Dependency resolved")

		} else {
			log.WithFields(log.Fields{
				"block":     b.Name,
				"workspace": workspace.Name,
			}).Debugf("Block has no dependencies")

		}

		// Register the Block to the Index
		b.address = b.Name
		workspace.registerBlock(b)

		// Update Artifacts, Kubeconfig & Inventory
		err := b.LoadArtifacts()
		if err != nil {
			b.err = err
			return b
		}
		b.LoadInventory()
		b.LoadKubeconfig()

		// Update Action addresses for all blocks not covered by dependencies
		if len(b.Actions) > 0 {
			log.WithFields(log.Fields{
				"block":     b.Name,
				"workspace": workspace.Name,
			}).Debugf("Updating action addresses")

			for _, action := range b.Actions {

				existingAction := b.getActionByName(action.Name)

				actionAddress := strings.Join([]string{b.Name, existingAction.Name}, ".")
				if existingAction.address != actionAddress {
					existingAction.address = actionAddress
					log.WithFields(log.Fields{
						"block":     b.Name,
						"action":    existingAction.Name,
						"workspace": workspace.Name,
						"address":   actionAddress,
					}).Debugf("Updated action address")
				}

				if existingAction.Block != b.Name {
					existingAction.Block = b.Name
					log.WithFields(log.Fields{
						"block":     b.Name,
						"action":    existingAction.Name,
						"workspace": workspace.Name,
						"address":   actionAddress,
					}).Debugf("Updated action block")
				}

				// Register the Action to the Index
				workspace.registerAction(existingAction)
			}
		}

	}
	b.resolved = true
	return b
}

func (b *Block) Load() *Block {
	blockConfigFilePath := filepath.Join(b.Workdir.LocalPath, workspace.Config.BlocksConfig)

	blockConfigObject := viper.New()
	blockConfigObject.SetConfigType("yaml")
	blockConfigObject.SetConfigFile(blockConfigFilePath)

	log.WithFields(log.Fields{
		"workspace": workspace.Name,
		"path":      b.Workdir.LocalPath,
	}).Debugf("Loading installed block")

	if err := blockConfigObject.MergeInConfig(); err != nil {
		b.err = err
		return b
	}

	if err := blockConfigObject.UnmarshalExact(&b); err != nil {
		b.err = err
		return b
	}
	if err := b.validate(); err != nil {
		b.err = err
		return b
	}

	// Set Block vars
	b.Workdir.ContainerPath = filepath.Join(workspace.ContainerPath, workspace.Config.BlocksRoot, filepath.Base(b.Workdir.LocalPath))

	if local {
		b.Workdir.Path = b.Workdir.LocalPath
	} else {
		b.Workdir.Path = b.Workdir.ContainerPath
	}

	log.WithFields(log.Fields{
		"block":     b.Name,
		"workspace": workspace.Name,
	}).Debugf("Loaded block")

	return b
}

func (b *Block) Reload() *Block {
	// Update Artifacts, Kubeconfig & Inventory
	err := b.LoadArtifacts()
	if err != nil {
		b.err = err
		return b
	}
	b.LoadInventory()
	b.LoadKubeconfig()
}

func (c *Block) getInventoryPath() string {
	log.Debugf("Remote inventory: %s", c.Inventory.From)

	if c.Inventory.From != "" {
		// Take the inventory from another Block
		log.Debugf("Remote inventory: %s", c.Inventory.From)
		inventorySourceBlock := workspace.getBlockByName(c.Inventory.From)
		if inventorySourceBlock != nil {
			if inventorySourceBlock.Inventory.exists {
				return inventorySourceBlock.Inventory.Path
			}
		}
	} else {
		if c.Inventory.exists {
			return c.Inventory.Path
		} else {
			return "/etc/ansible/hosts"
		}
	}
	return ""
}

func (c *Block) getKubeconfigPath() string {
	if c.Kubeconfig.From != "" {
		log.Debugf("Loading Kubeconfig from block %s", c.Kubeconfig.From)
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
		} else {
			log.Error("No kubeconfig found")
		}
	} else {
		log.Debugf("Not Loading Kubeconfig from other block")
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

	var blockInventoryFile string
	if c.Inventory.Filename != "" {
		blockInventoryFile = filepath.Join(c.Artifacts.LocalPath, c.Inventory.Filename)
	} else {
		blockInventoryFile = filepath.Join(c.Artifacts.LocalPath, "inventory.yml")
	}

	// log.WithFields(log.Fields{
	// 	"path":      blockInventoryFile,
	// 	"block":     c.Name,
	// 	"workspace": workspace.Name,
	// }).Debugf("Discovering inventory")

	if _, err := os.Stat(blockInventoryFile); !os.IsNotExist(err) {
		// File exists
		c.Inventory.exists = true
		c.Inventory.LocalPath = blockInventoryFile

		if c.Inventory.Filename != "" {
			c.Inventory.ContainerPath = filepath.Join(c.Artifacts.ContainerPath, c.Inventory.Filename)
		} else {
			c.Inventory.ContainerPath = filepath.Join(c.Artifacts.ContainerPath, "inventory.yml")
		}

		if local {
			c.Inventory.Path = c.Inventory.LocalPath
		} else {
			c.Inventory.Path = c.Inventory.ContainerPath
		}

		log.WithFields(log.Fields{
			"path":      blockInventoryFile,
			"block":     c.Name,
			"workspace": workspace.Name,
		}).Debugf("Inventory loaded")
	} else {
		c.Inventory.exists = false
	}

}

func (c *Block) LoadKubeconfig() {
	// Locate "kubeconfig.yml" in blockArtifactsDir
	var blockKubeconfigFile string
	if c.Kubeconfig.Filename != "" {
		blockKubeconfigFile = filepath.Join(c.Artifacts.LocalPath, c.Kubeconfig.Filename)
	} else {
		blockKubeconfigFile = filepath.Join(c.Artifacts.LocalPath, "kubeconfig.yml")
	}

	if _, err := os.Stat(blockKubeconfigFile); !os.IsNotExist(err) {
		// File exists
		c.Kubeconfig.exists = true
		c.Kubeconfig.LocalPath = blockKubeconfigFile

		if c.Kubeconfig.Filename != "" {
			c.Kubeconfig.ContainerPath = filepath.Join(c.Artifacts.ContainerPath, c.Kubeconfig.Filename)
		} else {
			c.Kubeconfig.ContainerPath = filepath.Join(c.Artifacts.ContainerPath, "kubeconfig.yml")
		}

		if local {
			c.Kubeconfig.Path = c.Kubeconfig.LocalPath
		} else {
			c.Kubeconfig.Path = c.Kubeconfig.ContainerPath
		}

		log.WithFields(log.Fields{
			"workspace": workspace.Name,
			"block":     c.Name,
			"path":      c.Kubeconfig.Path,
		}).Debugf("Kubeconfig loaded")
	} else {
		c.Kubeconfig.exists = false
	}

}

func (c *Block) LoadArtifacts() error {
	// e.g. $HOME/.polycrate/workspaces/workspace-1/artifacts/blocks/block-1
	c.Artifacts.LocalPath = filepath.Join(workspace.LocalPath, workspace.Config.ArtifactsRoot, workspace.Config.BlocksRoot, c.Name)
	// e.g. /workspace/artifacts/blocks/block-1
	c.Artifacts.ContainerPath = filepath.Join(workspace.ContainerPath, workspace.Config.ArtifactsRoot, workspace.Config.BlocksRoot, c.Name)

	if local {
		c.Artifacts.Path = c.Artifacts.LocalPath
	} else {
		c.Artifacts.Path = c.Artifacts.ContainerPath
	}

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
	// e.g. $HOME/.polycrate/workspaces/workspace-1/artifacts/blocks/block-1
	if _, err := os.Stat(c.Workdir.LocalPath); os.IsNotExist(err) {
		log.WithFields(log.Fields{
			"workspace": workspace.Name,
			"block":     c.Name,
			"path":      c.Workdir.LocalPath,
		}).Debugf("Block directory does not exist")
	} else {
		log.WithFields(log.Fields{
			"workspace": workspace.Name,
			"block":     c.Name,
			"path":      c.Workdir.LocalPath,
		}).Debugf("Removing block directory")

		err := os.RemoveAll(c.Workdir.LocalPath)
		if err != nil {
			return err
		}

		if prune {
			log.WithFields(log.Fields{
				"workspace": workspace.Name,
				"block":     c.Name,
				"path":      c.Artifacts.LocalPath,
			}).Debugf("Pruning artifacts")

			err := os.RemoveAll(c.Artifacts.LocalPath)
			if err != nil {
				return err
			}
		}
	}
	log.WithFields(log.Fields{
		"workspace": workspace.Name,
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
