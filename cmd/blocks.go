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
	"io/ioutil"
	"os"
	"path/filepath"
	"polycrate/cmd/mergo"
	"strings"

	"github.com/go-playground/validator/v10"

	//"github.com/imdario/mergo"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xeipuuv/gojsonschema"
	"golang.org/x/sync/errgroup"
)

var pruneBlock bool
var blockVersion string

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
		_w := cmd.Flags().Lookup("workspace").Value.String()

		tx := polycrate.Transaction().SetCommand(cmd)
		defer tx.Stop()

		workspace, err := polycrate.LoadWorkspace(tx, _w, true)
		if err != nil {
			tx.Log.Fatal(err)
		}

		workspace.ListBlocks()
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
	exists        bool
	Path          string `yaml:"path,omitempty" mapstructure:"path,omitempty" json:"path,omitempty"`
	From          string `yaml:"from,omitempty" mapstructure:"from,omitempty" json:"from,omitempty"`
	Filename      string `yaml:"filename,omitempty" mapstructure:"filename,omitempty" json:"filename,omitempty"`
	LocalPath     string `yaml:"localpath,omitempty" mapstructure:"localpath,omitempty" json:"localpath,omitempty"`
	ContainerPath string `yaml:"containerpath,omitempty" mapstructure:"containerpath,omitempty" json:"containerpath,omitempty"`
}
type BlockInventory struct {
	exists        bool
	Path          string `yaml:"path,omitempty" mapstructure:"path,omitempty" json:"path,omitempty"`
	From          string `yaml:"from,omitempty" mapstructure:"from,omitempty" json:"from,omitempty"`
	Filename      string `yaml:"filename,omitempty" mapstructure:"filename,omitempty" json:"filename,omitempty"`
	LocalPath     string `yaml:"localpath,omitempty" mapstructure:"localpath,omitempty" json:"localpath,omitempty"`
	ContainerPath string `yaml:"containerpath,omitempty" mapstructure:"containerpath,omitempty" json:"containerpath,omitempty"`
}
type BlockArtifacts struct {
	Path          string `yaml:"path,omitempty" mapstructure:"path,omitempty" json:"path,omitempty"`
	LocalPath     string `yaml:"localpath,omitempty" mapstructure:"localpath,omitempty" json:"localpath,omitempty"`
	ContainerPath string `yaml:"containerpath,omitempty" mapstructure:"containerpath,omitempty" json:"containerpath,omitempty"`
}

type Block struct {
	Name        string                      `yaml:"name,omitempty" mapstructure:"name,omitempty" json:"name,omitempty" validate:"required,block_name"`
	Description string                      `yaml:"description,omitempty" mapstructure:"description,omitempty" json:"description,omitempty"`
	Labels      map[string]string           `yaml:"labels,omitempty" mapstructure:"labels,omitempty" json:"labels,omitempty"`
	Alias       []string                    `yaml:"alias,omitempty" mapstructure:"alias,omitempty" json:"alias,omitempty"`
	Actions     []*Action                   `yaml:"actions,omitempty" mapstructure:"actions,omitempty" json:"actions,omitempty"`
	Config      map[interface{}]interface{} `yaml:"config,omitempty" mapstructure:"config,omitempty" json:"config,omitempty"`
	From        string                      `yaml:"from,omitempty" mapstructure:"from,omitempty" json:"from,omitempty"`
	Template    bool                        `yaml:"template,omitempty" mapstructure:"template,omitempty" json:"template,omitempty"`
	Version     string                      `yaml:"version,omitempty" mapstructure:"version" json:"version"`
	Workdir     BlockWorkdir                `yaml:"workdir,omitempty" mapstructure:"workdir,omitempty" json:"workdir,omitempty"`
	Inventory   BlockInventory              `yaml:"inventory,omitempty" mapstructure:"inventory,omitempty" json:"inventory,omitempty"`
	Kubeconfig  BlockKubeconfig             `yaml:"kubeconfig,omitempty" mapstructure:"kubeconfig,omitempty" json:"kubeconfig,omitempty"`
	Artifacts   BlockArtifacts              `yaml:"artifacts,omitempty" mapstructure:"artifacts,omitempty" json:"artifacts,omitempty"`
	resolved    bool
	schema      string
	workspace   *Workspace
	blockConfig viper.Viper
}

type SSHHost struct {
	Ip   string `yaml:"ip,omitempty" mapstructure:"ip,omitempty" json:"ip,omitempty"`
	Port string `yaml:"port,omitempty" mapstructure:"port,omitempty" json:"port,omitempty"`
	User string `yaml:"user,omitempty" mapstructure:"user,omitempty" json:"user,omitempty"`
}

func (b *Block) getHostsFromInventory(tx *PolycrateTransaction) (*viper.Viper, error) {

	workspace := b.workspace

	containerName := slugify([]string{tx.TXID.String(), "ssh"})
	fileName := strings.Join([]string{containerName, "yml"}, ".")
	workspace.registerEnvVar("ANSIBLE_INVENTORY", b.getInventoryPath(tx))
	workspace.registerEnvVar("KUBECONFIG", b.getKubeconfigPath(tx))
	workspace.registerCurrentBlock(b)

	// Create temp file to write output to
	f, err := polycrate.getTempFile(tx.Context, fileName)
	if err != nil {
		return nil, err
	}

	workspace.registerMount(f.Name(), f.Name())

	if _, err := workspace.SaveSnapshot(tx); err != nil {
		return nil, err
	}

	//cmd := "exec $(poly-utils ssh cmd " + hostname + ")"
	cmd := []string{
		"poly-utils",
		"inventory",
		"hosts",
		"--output-file",
		f.Name(),
	}

	tx.Log.Infof("Starting container for inventory conversion")
	err = workspace.RunContainer(tx, containerName, workspace.ContainerPath, cmd)
	if err != nil {
		return nil, err
	}

	// Load hosts from file
	var hosts = viper.New()
	if _, err := os.Stat(f.Name()); !os.IsNotExist(err) {
		hosts.SetConfigType("json")
		hosts.SetConfigFile(f.Name())
	} else {
		return nil, err
	}

	err = hosts.MergeInConfig()
	if err != nil {
		return nil, err
	}

	hosts_cache_path := filepath.Join(b.Artifacts.LocalPath, "hosts_cache.json")
	err = hosts.WriteConfigAs(hosts_cache_path)
	if err != nil {
		return nil, err
	}

	return hosts, nil
}

func (b *Block) SSH(tx *PolycrateTransaction, hostname string, refresh bool) error {
	workspace := b.workspace

	tx.Log.Infof("Starting SSH session")

	host_cache := map[string]SSHHost{}
	hosts_cache_path := filepath.Join(b.Artifacts.LocalPath, "hosts_cache.json")

	var hosts = viper.New()

	if _, err := os.Stat(hosts_cache_path); os.IsNotExist(err) || refresh {
		hosts, err = b.getHostsFromInventory(tx)
		if err != nil {
			return err
		}
	} else {
		tx.Log.Infof("Loading hosts from cache (use --refresh to update hosts list)")
		hosts.SetConfigType("json")
		hosts.SetConfigFile(hosts_cache_path)

		err = hosts.MergeInConfig()
		if err != nil {
			return err
		}

	}

	err := hosts.Unmarshal(&host_cache)
	if err != nil {
		return err
	}

	// Get IP
	//sshHost := hosts.Sub(hostname)
	sshHost := host_cache[hostname]

	//ip := sshHost.GetString("ip")
	ip := sshHost.Ip
	if ip == "" {
		return fmt.Errorf("ip of host %s not found", hostname)
	}
	user := sshHost.User
	if user == "" {
		// set default user
		user = "root"
	}
	port := sshHost.Port
	if port == "" {
		// set default port
		port = "22"
	}

	privateKey := filepath.Join(workspace.LocalPath, workspace.Config.SshPrivateKey)
	if privateKey == "" {
		err := fmt.Errorf("no private key found")
		return err
	}

	tx.Log.Infof("Connecting via SSH")

	//err = connectWithSSH(user, ip, port, privateKey)
	interactive = true
	err = ConnectWithSSH(tx, user, ip, port, privateKey)
	if err != nil {
		return err
	}

	return nil

}

func (b *Block) SSHList(tx *PolycrateTransaction, refresh bool, format bool) error {
	//workspace := b.workspace

	host_cache := map[string]SSHHost{}
	hosts_cache_path := filepath.Join(b.Artifacts.LocalPath, "hosts_cache.json")

	var hosts = viper.New()

	if _, err := os.Stat(hosts_cache_path); os.IsNotExist(err) || refresh {
		hosts, err = b.getHostsFromInventory(tx)
		if err != nil {
			return err
		}
	} else {
		tx.Log.Warnf("Loading hosts from cache (use --refresh to update hosts list)")
		hosts.SetConfigType("json")
		hosts.SetConfigFile(hosts_cache_path)

		err = hosts.MergeInConfig()
		if err != nil {
			return err
		}

	}

	err := hosts.Unmarshal(&host_cache)
	if err != nil {
		return err
	}

	if !format {
		printObject(host_cache)
	} else {
		for host, _ := range host_cache {
			polycrate_command := fmt.Sprintf("polycrate ssh --block %s %s", b.Name, host)
			fmt.Println(polycrate_command)
		}
	}

	return nil

}

func (b *Block) ConvertInventory(tx *PolycrateTransaction) error {
	log := tx.Log.log.WithField("block", b.Name)

	log.Infof("Converting inventory")
	workspace := b.workspace

	containerName := slugify([]string{tx.TXID.String(), "convert-inventory"})

	fileName := strings.Join([]string{containerName, "yml"}, ".")

	workspace.registerEnvVar("ANSIBLE_INVENTORY", b.getInventoryPath(tx))
	workspace.registerEnvVar("KUBECONFIG", b.getKubeconfigPath(tx))
	workspace.registerCurrentBlock(b)

	// Create temp file to write output to
	f, err := polycrate.getTempFile(tx.Context, fileName)
	if err != nil {
		return err
	}

	log.Debugf("Using temp file at %s", f.Name())

	workspace.registerMount(f.Name(), f.Name())

	if _, err := workspace.SaveSnapshot(tx); err != nil {
		return err
	}

	//cmd := "exec $(poly-utils ssh cmd " + hostname + ")"
	cmd := []string{
		"poly-utils",
		"inventory",
		"convert",
		"--output-file",
		f.Name(),
	}
	err = workspace.RunContainer(tx, containerName, workspace.ContainerPath, cmd)
	if err != nil {
		return err
	}

	// Load hosts from file
	var hosts = viper.New()
	if _, err := os.Stat(f.Name()); !os.IsNotExist(err) {
		hosts.SetConfigType("yaml")
		hosts.SetConfigFile(f.Name())
	} else {
		return err
	}

	err = hosts.MergeInConfig()
	if err != nil {
		return err
	}

	// Save to inventory.poly
	converted_inventory_path := filepath.Join(b.Artifacts.LocalPath, "converted-inventory.yml")

	err = hosts.WriteConfigAs(converted_inventory_path)
	if err != nil {
		return err
	}

	return nil

}

// func (b *Block) Resolve() *Block {
// 	if !b.resolved {
// 		log.WithFields(log.Fields{
// 			"block":     b.Name,
// 			"workspace": workspace.Name,
// 		}).Debugf("Resolving block %s", b.Name)

// 		// Check if a "from:" stanza is given and not empty
// 		// This means that the loadedBlock should inherit from another Block
// 		if b.From != "" {
// 			// a "from:" stanza is given
// 			log.WithFields(log.Fields{
// 				"block":      b.Name,
// 				"dependency": b.From,
// 				"workspace":  workspace.Name,
// 			}).Debugf("Dependency detected")

// 			// Try to load the referenced Block
// 			dependency := workspace.getBlockByName(b.From)

// 			if dependency == nil {
// 				log.WithFields(log.Fields{
// 					"block":      b.Name,
// 					"dependency": b.From,
// 					"workspace":  workspace.Name,
// 				}).Errorf("Dependency not found in the workspace")

// 				b.err = fmt.Errorf("Block '%s' not found in the Workspace. Please check the 'from' stanza of Block '%s'", b.From, b.Name)
// 				return b
// 			}

// 			log.WithFields(log.Fields{
// 				"block":      b.Name,
// 				"dependency": b.From,
// 				"workspace":  workspace.Name,
// 			}).Debugf("Dependency loaded")

// 			dep := *dependency

// 			// Check if the dependency Block has already been resolved
// 			// If not, re-queue the loaded Block so it can be resolved in another iteration
// 			if !dep.resolved {
// 				// Needed Block from 'from' stanza is not yet resolved
// 				log.WithFields(log.Fields{
// 					"block":               b.Name,
// 					"dependency":          b.From,
// 					"workspace":           workspace.Name,
// 					"dependency_resolved": dep.resolved,
// 				}).Debugf("Dependency not resolved")
// 				b.resolved = false
// 				b.err = DependencyNotResolved
// 				return b
// 			}

// 			// Merge the dependency Block into the loaded Block
// 			// We do NOT OVERWRITE existing values in the loaded Block because we must assume
// 			// That this is configuration that has been explicitly set by the user
// 			// The merge works like "loading defaults" for the loaded Block
// 			err := b.MergeIn(dep)
// 			if err != nil {
// 				b.err = err
// 				return b
// 			}

// 			// Handle Workdir
// 			b.Workdir.LocalPath = dep.Workdir.LocalPath
// 			b.Workdir.ContainerPath = dep.Workdir.ContainerPath

// 			log.WithFields(log.Fields{
// 				"block":      b.Name,
// 				"dependency": b.From,
// 				"workspace":  workspace.Name,
// 			}).Debugf("Dependency resolved")

// 		} else {
// 			log.WithFields(log.Fields{
// 				"block":     b.Name,
// 				"workspace": workspace.Name,
// 			}).Debugf("Block has no dependencies")

// 		}

// 		// Validate schema
// 		err := b.ValidateSchema()
// 		if err != nil {
// 			b.err = err
// 			return b
// 		}

// 		// Register the Block to the Index
// 		b.address = b.Name
// 		workspace.registerBlock(b)

// 		// Update Artifacts, Kubeconfig & Inventory
// 		err = b.LoadArtifacts()
// 		if err != nil {
// 			b.err = err
// 			return b
// 		}
// 		b.LoadInventory()
// 		b.LoadKubeconfig()

// 		// Update Action addresses for all blocks not covered by dependencies
// 		if len(b.Actions) > 0 {
// 			log.WithFields(log.Fields{
// 				"block":     b.Name,
// 				"workspace": workspace.Name,
// 			}).Debugf("Updating action addresses")

// 			for _, action := range b.Actions {

// 				existingAction := b.getActionByName(action.Name)

// 				actionAddress := strings.Join([]string{b.Name, existingAction.Name}, ".")
// 				if existingAction.address != actionAddress {
// 					existingAction.address = actionAddress
// 					log.WithFields(log.Fields{
// 						"block":     b.Name,
// 						"action":    existingAction.Name,
// 						"workspace": workspace.Name,
// 						"address":   actionAddress,
// 					}).Debugf("Updated action address")
// 				}

// 				if existingAction.Block != b.Name {
// 					existingAction.Block = b.Name
// 					log.WithFields(log.Fields{
// 						"block":     b.Name,
// 						"action":    existingAction.Name,
// 						"workspace": workspace.Name,
// 						"address":   actionAddress,
// 					}).Debugf("Updated action block")
// 				}

// 				err := existingAction.validate()
// 				if err != nil {
// 					b.err = err
// 					return b
// 				}

// 				// Register the Action to the Index
// 				workspace.registerAction(existingAction)
// 			}
// 		}

// 		b.resolved = true
// 	}
// 	return b
// }

// func (b *Block) Load() *Block {
// 	blockConfigFilePath := filepath.Join(b.Workdir.LocalPath, workspace.Config.BlocksConfig)

// 	blockConfigObject := viper.New()
// 	blockConfigObject.SetConfigType("yaml")
// 	blockConfigObject.SetConfigFile(blockConfigFilePath)

// 	log.WithFields(log.Fields{
// 		"workspace": workspace.Name,
// 		"path":      b.Workdir.LocalPath,
// 	}).Debugf("Loading installed block")

// 	if err := blockConfigObject.MergeInConfig(); err != nil {
// 		b.err = err
// 		return b
// 	}

// 	if err := blockConfigObject.UnmarshalExact(&b); err != nil {
// 		b.err = err
// 		return b
// 	}

// 	// Load schema
// 	schemaFilePath := filepath.Join(b.Workdir.LocalPath, "schema.json")
// 	if _, err := os.Stat(schemaFilePath); !os.IsNotExist(err) {
// 		b.schema = schemaFilePath

// 		log.WithFields(log.Fields{
// 			"block":       b.Name,
// 			"workspace":   workspace.Name,
// 			"schema_path": schemaFilePath,
// 		}).Debugf("Loaded schema")
// 	}

// 	if err := b.validate(); err != nil {
// 		b.err = err
// 		return b
// 	}

// 	// Set Block vars
// 	relativeBlockPath, err := filepath.Rel(filepath.Join(workspace.LocalPath, workspace.Config.BlocksRoot), b.Workdir.LocalPath)
// 	if err != nil {
// 		b.err = err
// 		return b
// 	}
// 	b.Workdir.ContainerPath = filepath.Join(workspace.ContainerPath, workspace.Config.BlocksRoot, relativeBlockPath)

// 	if local {
// 		b.Workdir.Path = b.Workdir.LocalPath
// 	} else {
// 		b.Workdir.Path = b.Workdir.ContainerPath
// 	}

// 	log.WithFields(log.Fields{
// 		"block":     b.Name,
// 		"workspace": workspace.Name,
// 	}).Debugf("Loaded block")

// 	return b
// }

func (b *Block) ValidateSchema() error {
	if b.schema == "" {
		return nil
	}

	log.WithFields(log.Fields{
		"block":       b.Name,
		"workspace":   workspace.Name,
		"schema_path": b.schema,
	}).Debugf("Validating schema")
	schemaLoader := gojsonschema.NewReferenceLoader("file://" + b.schema)

	rawData := gojsonschema.NewGoLoader(b.Config)
	result, err := gojsonschema.Validate(schemaLoader, rawData)
	if err != nil {
		return err
	}

	//state := rs.Validate(ctx, b.Config)
	if !result.Valid() {
		for _, desc := range result.Errors() {
			log.WithFields(log.Fields{
				"block":     b.Name,
				"workspace": workspace.Name,
			}).Errorf("%s", desc)
		}
		return fmt.Errorf("failed to validate schema for block '%s'", b.Name)
	}

	return nil
}

func (b *Block) Reload(tx *PolycrateTransaction) error {
	eg := new(errgroup.Group)

	// Update Artifacts, Kubeconfig & Inventory
	eg.Go(func() error {
		err := b.LoadArtifacts(tx)
		if err != nil {
			return err
		}

		err = b.LoadInventory(tx)
		if err != nil {
			return err
		}

		err = b.LoadKubeconfig(tx)
		if err != nil {
			return err
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

func (c *Block) getInventoryPath(tx *PolycrateTransaction) string {

	workspace := c.workspace

	if c.Inventory.From != "" {
		// Take the inventory from another Block
		tx.Log.Debugf("Loading inventory from block %s", c.Inventory.From)
		inventorySourceBlock, err := workspace.GetBlock(c.Inventory.From)
		if err != nil {
			tx.Log.Errorf("Block %s not found", c.Inventory.From)
		}

		if inventorySourceBlock != nil {
			return inventorySourceBlock.Inventory.Path
		} else {
			tx.Log.Errorf("Inventory source '%s' not found", c.Inventory.From)
		}
	} else {
		return c.Inventory.Path
	}
	return "/etc/ansible/hosts"
}

func (b *Block) getKubeconfigPath(tx *PolycrateTransaction) string {
	workspace := b.workspace

	if b.Kubeconfig.From != "" {
		// Take the kubeconfig from another Block
		tx.Log.Debugf("Loading Kubeconfig from block %s", b.Kubeconfig.From)
		kubeconfigSourceBlock, err := workspace.GetBlock(b.Kubeconfig.From)
		if err != nil {
			tx.Log.Errorf("Block %s not found", b.Kubeconfig.From)
		}

		if kubeconfigSourceBlock != nil {
			return kubeconfigSourceBlock.Kubeconfig.Path
		} else {
			tx.Log.Errorf("Kubeconfig source '%s' not found", b.Kubeconfig.From)
		}
	} else {
		return b.Kubeconfig.Path
	}
	return ""
}

func (b *Block) GetAction(name string) (*Action, error) {
	for i := 0; i < len(b.Actions); i++ {
		action := b.Actions[i]
		if action.Name == name {
			return action, nil
		}
	}
	return nil, fmt.Errorf("action not found: %s", name)
}

func (c *Block) getActionByName(actionName string) *Action {

	//for _, block := range c.Blocks {
	for i := 0; i < len(c.Actions); i++ {
		action := c.Actions[i]
		if action.Name == actionName {
			return action
		}
	}
	return nil
}

func (b *Block) validate() error {
	err := validate.Struct(b)

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

	// Schema
	if block.schema != "" && c.schema == "" {
		c.schema = block.schema
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

				destinationAction.Block = c.Name
				destinationAction.block = c
			} else {
				sourceAction.Block = c.Name
				sourceAction.block = c
				c.Actions = append(c.Actions, sourceAction)
			}
		}
	}

	// Config
	if block.Config != nil {
		if polycrate.Config.Experimental.MergeV2 {
			log.Debug("Merging block config using experimental merge method")

			c.Config = mergeMaps(block.Config, c.Config)
		} else {

			if err := mergo.Merge(&c.Config, block.Config); err != nil {
				return err
			}
		}
		// feature flag: merge-v2
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

func (b *Block) README() error {
	readme := filepath.Join(b.Workdir.LocalPath, "README.md")
	if _, err := os.Stat(readme); os.IsNotExist(err) {
		return err
	}
	file, err := os.Open(readme)
	if err != nil {
		return err
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	r, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	fmt.Println(string(r))

	return nil
}

func (b *Block) Defaults() {
	printObject(b)
}

func (b *Block) Inspect() {
	printObject(b)
}

func (c *Block) LoadInventory(tx *PolycrateTransaction) error {
	// Locate "inventory.yml" in blockArtifactsDir

	workspace := c.workspace

	var localInventoryFile string
	var containerInventoryFile string
	if c.Inventory.Filename != "" {
		localInventoryFile = filepath.Join(c.Artifacts.LocalPath, c.Inventory.Filename)
		containerInventoryFile = filepath.Join(c.Artifacts.ContainerPath, c.Inventory.Filename)
	} else {
		localInventoryFile = filepath.Join(c.Artifacts.LocalPath, "inventory.yml")
		containerInventoryFile = filepath.Join(c.Artifacts.ContainerPath, "inventory.yml")
	}

	if _, err := os.Stat(localInventoryFile); !os.IsNotExist(err) {
		// File exists
		c.Inventory.exists = true
		c.Inventory.LocalPath = localInventoryFile
		c.Inventory.ContainerPath = containerInventoryFile
	} else {
		// Check if workspace inventory exists
		if workspace.Inventory.exists {
			c.Inventory.exists = true
			c.Inventory.LocalPath = workspace.Inventory.LocalPath
			c.Inventory.ContainerPath = workspace.Inventory.ContainerPath

			tx.Log.Debugf("Using workspace inventory at %s for block", c.Inventory.LocalPath)
		} else {
			c.Inventory.exists = false
			c.Inventory.LocalPath = "/etc/ansible/hosts"
			c.Inventory.ContainerPath = "/etc/ansible/hosts"
			tx.Log.Debugf("Using fallback inventory at /etc/ansible/hosts for block")
		}
	}

	if local {
		c.Inventory.Path = c.Inventory.LocalPath
	} else {
		c.Inventory.Path = c.Inventory.ContainerPath
	}

	return nil
}

func (c *Block) LoadKubeconfig(tx *PolycrateTransaction) error {
	// Locate "kubeconfig.yml" in blockArtifactsDir

	workspace := c.workspace

	var localKubeconfigFile string
	var containerKubeconfigFile string
	if c.Kubeconfig.Filename != "" {
		localKubeconfigFile = filepath.Join(c.Artifacts.LocalPath, c.Kubeconfig.Filename)
		containerKubeconfigFile = filepath.Join(c.Artifacts.ContainerPath, c.Kubeconfig.Filename)
	} else {
		localKubeconfigFile = filepath.Join(c.Artifacts.LocalPath, "kubeconfig.yml")
		containerKubeconfigFile = filepath.Join(c.Artifacts.ContainerPath, "kubeconfig.yml")
	}

	if _, err := os.Stat(localKubeconfigFile); !os.IsNotExist(err) {
		// File exists
		c.Kubeconfig.exists = true
		c.Kubeconfig.LocalPath = localKubeconfigFile
		c.Kubeconfig.ContainerPath = containerKubeconfigFile
	} else {
		// Check if workspace inventory exists
		if workspace.Kubeconfig.exists {
			c.Kubeconfig.exists = true
			c.Kubeconfig.LocalPath = workspace.Kubeconfig.LocalPath
			c.Kubeconfig.ContainerPath = workspace.Kubeconfig.ContainerPath
		} else {
			c.Kubeconfig.exists = false
		}
	}
	if local {
		c.Kubeconfig.Path = c.Kubeconfig.LocalPath
	} else {
		c.Kubeconfig.Path = c.Kubeconfig.ContainerPath
	}

	return nil
}

func (b *Block) LoadArtifacts(tx *PolycrateTransaction) error {
	// e.g. $HOME/.polycrate/workspaces/workspace-1/artifacts/blocks/block-1

	b.Artifacts.LocalPath = filepath.Join(b.workspace.LocalPath, b.workspace.Config.ArtifactsRoot, b.workspace.Config.BlocksRoot, b.Name)
	// e.g. /workspace/artifacts/blocks/block-1
	b.Artifacts.ContainerPath = filepath.Join(b.workspace.ContainerPath, b.workspace.Config.ArtifactsRoot, b.workspace.Config.BlocksRoot, b.Name)

	if local {
		b.Artifacts.Path = b.Artifacts.LocalPath
	} else {
		b.Artifacts.Path = b.Artifacts.ContainerPath
	}

	// Check if the local artifacts directory for this Block exists
	if _, err := os.Stat(b.Artifacts.LocalPath); os.IsNotExist(err) {
		// Directory does not exist
		// We create it
		if err := os.MkdirAll(b.Artifacts.LocalPath, os.ModePerm); err != nil {
			return err
		}
		tx.Log.Debugf("Created artifacts directory for block %s at %s", b.Name, b.Artifacts.LocalPath)
	}
	return nil

}

func (c *Block) Uninstall(tx *PolycrateTransaction, prune bool) error {
	// e.g. $HOME/.polycrate/workspaces/workspace-1/artifacts/blocks/block-1

	if _, err := os.Stat(c.Workdir.LocalPath); os.IsNotExist(err) {
		tx.Log.Debug("Block directory does not exist")
	} else {
		tx.Log.Debug("Removing block directory")

		err := os.RemoveAll(c.Workdir.LocalPath)
		if err != nil {
			return err
		}

		if prune {
			tx.Log.Debug("Pruning artifacts")

			err := os.RemoveAll(c.Artifacts.LocalPath)
			if err != nil {
				return err
			}
		}
	}
	tx.Log.Debug("Successfully uninstalled block from workspace")
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
