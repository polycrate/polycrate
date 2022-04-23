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
	"path/filepath"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var stackCmd = &cobra.Command{
	Use:   "stack",
	Short: "Show stack",
	Long:  `Show stack`,
	Run: func(cmd *cobra.Command, args []string) {
		showWorkspace()
	},
}

func init() {
	rootCmd.AddCommand(stackCmd)
}

type ImageConfig struct {
	Reference string `mapstructure:"reference" json:"reference" validate:"required"`
	Version   string `mapstructure:"version" json:"version" validate:"required"`
}

type Metadata struct {
	Name        string            `mapstructure:"name" json:"name" validate:"required"`
	Description string            `mapstructure:"description" json:"description"`
	Labels      map[string]string `mapstructure:"labels" json:"labels"`
	Alias       []string          `mapstructure:"alias" json:"alias"`
}

type WorkspaceConfig struct {
	Image         ImageConfig            `mapstructure:"image" json:"image" validate:"required"`
	BlocksRoot    string                 `mapstructure:"blocksroot" json:"blocksroot" validate:"required"`
	WorkflowsRoot string                 `mapstructure:"workflowsroot" json:"workflowsroot" validate:"required"`
	StateRoot     string                 `mapstructure:"stateroot" json:"stateroot" validate:"required"`
	ContainerRoot string                 `mapstructure:"containerroot" json:"containerroot" validate:"required"`
	SshPrivateKey string                 `mapstructure:"sshprivatekey" json:"sshprivatekey" validate:"required"`
	SshPublicKey  string                 `mapstructure:"sshpublickey" json:"sshpublickey" validate:"required"`
	Globals       map[string]interface{} `mapstructure:"globals,remain" json:"globals,remain"`
}

type Workspace struct {
	Metadata  Metadata        `mapstructure:"metadata" json:"metadata" validate:"required"`
	Config    WorkspaceConfig `mapstructure:"config" json:"config"`
	Blocks    []Block         `mapstructure:"blocks" json:"blocks" validate:"dive,required"`
	Workflows []Workflow      `mapstructure:"workflows,omitempty" json:"workflows,omitempty"`
}

func (c *Workspace) loadWorkspaceConfig() error {
	workspaceConfig.SetConfigType("yaml")
	workspaceConfig.SetConfigFile(workspaceConfigFilePath)

	err := workspaceConfig.MergeInConfig()
	if err != nil {
		log.Warn("Error loading workspaceConfig: " + err.Error())

		workspaceConfig.SetDefault("metadata.name", filepath.Base(cwd))
		workspaceConfig.SetDefault("metadata.description", "Ad-hoc Workspace in "+cwd)
	}

	return nil
}

func (c *Workspace) loadBlocks() error {
	return nil
}

func (c *Workspace) unmarshal() error {
	config := defaultDecoderConfig(&c)
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}
	return decoder.Decode(workspaceConfig)
}

func (c *Workspace) unmarshalWorkspaceConfig() error {
	err := workspaceConfig.Unmarshal(&c)

	if err != nil {
		return err
	}

	err = c.Validate()
	return err
}

func (c *Workspace) load() error {
	log.Debugf("Loading Workspace")

	workspaceConfigFilePath = filepath.Join(workspaceDir, workspaceConfigFile)
	workspaceContainerConfigFilePath = filepath.Join(workspaceContainerDir, workspaceConfigFile)

	blocksDir = filepath.Join(workspaceDir, blocksRoot)
	blocksContainerDir = filepath.Join(workspaceContainerDir, blocksRoot)

	blockContainerDir = filepath.Join(blocksContainerDir, blockName)
	blockContainerConfigFilePath = filepath.Join(blockContainerDir, blockConfigFile)

	workspaceConfig.BindPFlag("config.image.version", rootCmd.Flags().Lookup("image-version"))
	workspaceConfig.BindPFlag("config.image.reference", rootCmd.Flags().Lookup("image-ref"))
	workspaceConfig.BindPFlag("config.blocksroot", rootCmd.Flags().Lookup("blocks-root"))
	workspaceConfig.BindPFlag("config.workflowsroot", rootCmd.Flags().Lookup("workflows-root"))
	workspaceConfig.BindPFlag("config.stateroot", rootCmd.Flags().Lookup("state-root"))
	workspaceConfig.BindPFlag("config.containerroot", rootCmd.Flags().Lookup("container-root"))
	workspaceConfig.BindPFlag("config.remoteroot", rootCmd.Flags().Lookup("remote-root"))
	workspaceConfig.BindPFlag("config.sshprivatekey", rootCmd.Flags().Lookup("ssh-private-key"))
	workspaceConfig.BindPFlag("config.sshpublickey", rootCmd.Flags().Lookup("ssh-public-key"))

	workspaceConfig.SetEnvPrefix(envPrefix)
	workspaceConfig.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	workspaceConfig.AutomaticEnv()

	// Load Workspace config (e.g. workspace.yml)
	if err := c.loadWorkspaceConfig(); err != nil {
		log.Fatal(err)
	}

	// Unmarshal + Validate Workspace config
	if err := c.unmarshalWorkspaceConfig(); err != nil {
		log.Fatal(err)
	}

	// Find all Blocks in the Workspace
	if err := c.discoverBlocks(); err != nil {
		log.Fatal(err)
	}

	// Load all Blocks in the Workspace
	if err := c.loadBlockConfigs(); err != nil {
		log.Fatal(err)
	}

	// Resolve Block dependencies
	if err := c.resolveBlockDependencies(); err != nil {
		log.Fatal(err)
	}

	log.Debugf("Workspace ready")
	return nil
}

func (c *Workspace) resolveBlockDependencies() error {
	missing := len(c.Blocks)

	log.Debug("Resolving Block dependencies")

	// Iterate over all Blocks in the Workspace
	// Until nothing is missing anymore
	for missing != 0 {
		for i := 0; i < len(c.Blocks); i++ {
			loadedBlock := &c.Blocks[i]

			// Block has not been resolved yet
			if !loadedBlock.resolved {

				log.Debugf("Trying to resolve Block '%s'", loadedBlock.Metadata.Name)

				if loadedBlock.From != "" {
					// a "from:" stanza is given
					log.Debugf("Block has dependency: '%s'", loadedBlock.From)

					// Try to load the referenced Block
					dependency := c.getBlockByName(loadedBlock.From)

					if dependency == nil {
						// There's no Block to load from
						return goErrors.New("Block '" + loadedBlock.From + "' not found in the Workspace. Please check the 'from' stanza of Block " + loadedBlock.Metadata.Name)
					}

					if !dependency.resolved {
						// Needed Block from 'from' stanza is not yet resolved
						log.Debugf("Postponing Block '%s' because its dependency '%s' is not yet resolved", loadedBlock.Metadata.Name, dependency.Metadata.Name)
						loadedBlock.resolved = false
						continue
					}

					err := loadedBlock.MergeIn(dependency)
					if err != nil {
						return err
					}
					// opts := conjungo.NewOptions()
					// opts.Overwrite = false // do not overwrite existing values in workspaceConfig
					// if err := conjungo.Merge(loadedBlock, dependency, opts); err != nil {
					// 	return err
					// }
					loadedBlock.resolved = true
					loadedBlock.parent = dependency
					missing--
					log.Debugf("Resolved Block '%s' from dependency '%s'", loadedBlock.Metadata.Name, loadedBlock.From)
				} else {
					loadedBlock.resolved = true
					missing--
					log.Debugf("Resolved Block'%s'", loadedBlock.Metadata.Name)
				}
				log.Debugf("%d Blocks left", missing)
			}
		}
	}
	return nil
}

func (c *Workspace) Validate() error {
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
		return goErrors.New("Error validating Workspace")
	}
	return nil
}

func (c *Workspace) listBlocks() {
	for _, block := range c.Blocks {
		fmt.Println(block.Metadata.Name)
	}
}

func (c *Workspace) getBlockByName(blockName string) *Block {

	//for _, block := range c.Blocks {
	for i := 0; i < len(c.Blocks); i++ {
		block := &c.Blocks[i]
		if block.Metadata.Name == blockName {
			return block
		}
	}
	return nil
}

func (c *Workspace) print() {
	printObject(&c)
}

func (c *Workspace) loadBlockConfigs() error {
	log.Debugf("Loading Blocks")

	for _, blockPath := range blockPaths {
		blockConfigFilePath := filepath.Join(blockPath, blockConfigFile)

		blockConfigObject := viper.New()
		blockConfigObject.SetConfigType("yaml")
		blockConfigObject.SetConfigFile(blockConfigFilePath)

		log.Debug("Loading ", blockConfigFile, " from "+blockPath)
		if err := blockConfigObject.MergeInConfig(); err != nil {
			return err
		}

		var loadedBlock Block
		if err := blockConfigObject.UnmarshalExact(&loadedBlock); err != nil {
			return err
		}
		if err := loadedBlock.Validate(); err != nil {
			return err
		}
		log.Debugf("Loaded Block '%s'", loadedBlock.Metadata.Name)

		// Check if Block exists
		existingBlock := c.getBlockByName(loadedBlock.Metadata.Name)

		if existingBlock != nil {
			// Block exists
			log.Debugf("Found existing Block '%s' in the Workspace. Merging with loaded Block.", existingBlock.Metadata.Name)

			err := existingBlock.MergeIn(&loadedBlock)
			if err != nil {
				return err
			}

		} else {
			c.Blocks = append(c.Blocks, loadedBlock)
		}
	}
	return nil

}

func (c *Workspace) discoverBlocks() error {
	log.Debugf("Starting Block Discovery at %s", blocksDir)

	// This function adds all valid Blocks to the list of
	err := filepath.WalkDir(blocksDir, walkBlocksDir)
	if err != nil {
		return err
	}

	return nil
}
