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
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var workspaceCmd = &cobra.Command{
	Use:   "workspace",
	Short: "Manage the Workspace",
	Long:  `Manage the Workspace`,
	Run: func(cmd *cobra.Command, args []string) {
		showWorkspace()
	},
}

func init() {
	rootCmd.AddCommand(workspaceCmd)
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
	Metadata        Metadata        `mapstructure:"metadata" json:"metadata" validate:"required"`
	Config          WorkspaceConfig `mapstructure:"config" json:"config"`
	Blocks          []Block         `mapstructure:"blocks" json:"blocks" validate:"dive,required"`
	Workflows       []Workflow      `mapstructure:"workflows,omitempty" json:"workflows,omitempty"`
	currentBlock    *Block
	currentAction   *Action
	currentWorkflow *Workflow
	currentStep     *Step
	index           *WorkspaceIndex
	env             map[string]string
	mounts          map[string]string
	err             error
}

type WorkspaceIndex struct {
	Actions   map[string]*Action   `mapstructure:"actions" json:"actions"`
	Steps     map[string]*Step     `mapstructure:"steps" json:"steps"`
	Blocks    map[string]*Block    `mapstructure:"blocks" json:"blocks"`
	Workflows map[string]*Workflow `mapstructure:"workflows" json:"workflows"`
}

type WorkspaceSnapshot struct {
	Workspace *Workspace
	Action    *Action
	Block     *Block
	Workflow  *Workflow
	Step      *Step
	Env       *map[string]string
	Mounts    *map[string]string
}

func (c *Workspace) Snapshot() {
	snapshot := WorkspaceSnapshot{
		Workspace: c,
		Action:    c.currentAction,
		Block:     c.currentBlock,
		Workflow:  currentWorkflow,
		Step:      c.currentStep,
		Env:       &c.env,
		Mounts:    &c.mounts,
	}
	printObject(snapshot)
}

func (c *Workspace) Flush() error {
	if c.err != nil {
		return c.err
	}
	return nil
}

func (c *Workspace) RunAction(address string) {
	// Find action in index and report
	action := c.LookupAction(address)

	if action != nil {
		workspace.registerCurrentAction(action)
		workspace.registerCurrentBlock(action.block)

		if snapshot {
			c.Snapshot()
		} else {
			action.Run()
		}
	} else {
		log.Fatal("Cannot find Action with address " + address)
	}
}

func (c *Workspace) resolveActionAddress(actionAddress string) (*Block, *Action, error) {
	s := strings.Split(actionAddress, ".")

	if len(s) < 2 {
		err := goErrors.New("Action Address malformed: " + actionAddress)
		return nil, nil, err
	} else {
		blockName := s[0]
		actionName := s[1]

		block := workspace.getBlockByName(blockName)

		if block != nil {
			action := block.getActionByName(actionName)

			if action != nil {
				return block, action, nil
			} else {
				err := goErrors.New("Action not found: " + actionName)
				return nil, nil, err
			}
		} else {
			err := goErrors.New("Block not found: " + blockName)
			return nil, nil, err
		}
		err := goErrors.New("Can't resolve Action Address: " + actionAddress)
		return nil, nil, err
	}
}

func (c *Workspace) registerBlock(block *Block) {
	c.index.Blocks[block.address] = block
}
func (c *Workspace) registerWorkflow(workflow *Workflow) {
	c.index.Workflows[workflow.address] = workflow
}
func (c *Workspace) registerAction(action *Action) {
	c.index.Actions[action.address] = action
}
func (c *Workspace) registerStep(step *Step) {
	c.index.Steps[step.address] = step
}

func (c *Workspace) LookupBlock(address string) *Block {
	return c.index.Blocks[address]
}
func (c *Workspace) LookupAction(address string) *Action {
	return c.index.Actions[address]
}
func (c *Workspace) ActivateAction(address string) *Action {
	return c.index.Actions[address]
}
func (c *Workspace) LookupWorkflow(address string) *Workflow {
	return c.index.Workflows[address]
}
func (c *Workspace) LookupStep(address string) *Step {
	return c.index.Steps[address]
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

func (c *Workspace) unmarshalWorkspaceConfig() error {
	err := workspaceConfig.Unmarshal(&c)

	if err != nil {
		return err
	}

	err = c.Validate()
	return err
}

func (c *Workspace) load() {
	log.Debugf("Loading Workspace")

	workspaceConfigFilePath = filepath.Join(workspaceDir, workspaceConfigFile)
	workspaceContainerConfigFilePath = filepath.Join(workspaceContainerDir, workspaceConfigFile)

	blocksDir = filepath.Join(workspaceDir, blocksRoot)
	blocksContainerDir = filepath.Join(workspaceContainerDir, blocksRoot)

	blockContainerDir = filepath.Join(blocksContainerDir, blockName)

	workspaceConfig.BindPFlag("config.image.version", rootCmd.Flags().Lookup("image-version"))
	workspaceConfig.BindPFlag("config.image.reference", rootCmd.Flags().Lookup("image-ref"))
	workspaceConfig.BindPFlag("config.blocksroot", rootCmd.Flags().Lookup("blocks-root"))
	workspaceConfig.BindPFlag("config.workflowsroot", rootCmd.Flags().Lookup("workflows-root"))
	workspaceConfig.BindPFlag("config.stateroot", rootCmd.Flags().Lookup("artifacts-root"))
	workspaceConfig.BindPFlag("config.containerroot", rootCmd.Flags().Lookup("container-root"))
	workspaceConfig.BindPFlag("config.remoteroot", rootCmd.Flags().Lookup("remote-root"))
	workspaceConfig.BindPFlag("config.sshprivatekey", rootCmd.Flags().Lookup("ssh-private-key"))
	workspaceConfig.BindPFlag("config.sshpublickey", rootCmd.Flags().Lookup("ssh-public-key"))

	workspaceConfig.SetEnvPrefix(envPrefix)
	workspaceConfig.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	workspaceConfig.AutomaticEnv()

	// Load Workspace config (e.g. workspace.poly)
	if err := c.loadWorkspaceConfig(); err != nil {
		log.Fatal(err)
	}

	// Unmarshal + Validate Workspace config
	if err := c.unmarshalWorkspaceConfig(); err != nil {
		log.Fatal(err)
	}

	// Bootstrap Index
	c.bootstrapIndex()

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

	// Bootstrap env vars
	c.bootstrapEnvVars()

	// Bootstrap container mounts
	c.bootstrapMounts()

	log.Debugf("Workspace ready")
}

func (c *Workspace) resolveBlockDependencies() error {
	missing := len(c.Blocks)

	log.Debug("Resolving Block dependencies")

	// Iterate over all Blocks in the Workspace
	// Until nothing is "missing" anymore
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

					// Handle Actions
					// Iterate over the loaded Block's Actions
					// and check if it exists in the existing Block
					// If not, add it
					for _, loadedAction := range dependency.Actions {
						log.Debugf("Analyzing Action '%s'", loadedAction.Metadata.Name)
						existingAction := loadedBlock.getActionByName(loadedAction.Metadata.Name)

						if existingAction != nil {
							log.Debugf("Found existing Action '%s'", existingAction.Metadata.Name)
							// An Action with the same name exists
							// We merge!
							err := existingAction.MergeIn(&loadedAction)
							if err != nil {
								return err
							}
						} else {
							log.Debugf("No existing Action found. Adding '%s'", loadedAction.Metadata.Name)
							loadedBlock.Actions = append(loadedBlock.Actions, loadedAction)
						}
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

				// Register the Block to the Index
				loadedBlock.address = strings.Join([]string{loadedBlock.Metadata.Name}, ".")
				c.registerBlock(loadedBlock)

				// Update Artifacts, Kubeconfig & Inventory
				loadedBlock.artifacts.localPath = filepath.Join(workspaceDir, artifactsRoot, blocksRoot, loadedBlock.Metadata.Name)
				loadedBlock.artifacts.containerPath = filepath.Join(workspaceContainerDir, artifactsRoot, blocksRoot, loadedBlock.Metadata.Name)

				loadedBlock.LoadInventory()
				loadedBlock.LoadKubeconfig()

				// Update Action addresses
				for _, action := range loadedBlock.Actions {
					action.address = strings.Join([]string{loadedBlock.Metadata.Name, action.Metadata.Name}, ".")
					action.block = loadedBlock

					// Register the Action to the Index
					c.registerAction(&action)
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

func (c *Workspace) listWorkflows() {
	for workflow := range c.Workflows {
		fmt.Println(workflow)
	}
}

func (c *Workspace) listActions() {
	for _, block := range c.Blocks {
		for _, action := range block.Actions {
			fmt.Printf("%s.%s\n", block.Metadata.Name, action.Metadata.Name)
		}
	}
}

func (c *Workspace) bootstrapIndex() {
	c.index = &WorkspaceIndex{}
	c.index.Actions = make(map[string]*Action)
	c.index.Blocks = make(map[string]*Block)
	c.index.Workflows = make(map[string]*Workflow)
	c.index.Steps = make(map[string]*Step)
}

func (c *Workspace) bootstrapMounts() []string {
	// Prepare mounts map
	c.mounts = map[string]string{}

	// Mount the Workspace
	c.registerMount(workspaceDir, workspaceContainerDir)

	// Mount the Docker Socket if possible
	dockerSocket := "/var/run/docker.sock"
	if _, err := os.Stat(dockerSocket); !os.IsNotExist(err) {
		c.registerMount(dockerSocket, dockerSocket)
	}

	// if extraMounts != nil {

	// 	mounts = append(mounts, extraMounts...)
	// }
	// if devDir != "" {
	// 	mounts = append(mounts, strings.Join([]string{devDir, "/polycrate"}, ":"))
	// }
	for _, extraMount := range extraMounts {
		// Split by =
		p := strings.Split(extraMount, ":")

		if len(p) == 2 {
			c.registerMount(p[0], p[1])
		} else {
			c.err = goErrors.New("Illegal value for mount found: " + extraMount)
			break
		}
	}
	return mounts
}

func (c *Workspace) bootstrapEnvVars() {
	// Prepare env map
	c.env = map[string]string{}

	var _force string
	if force {
		_force = "1"
	} else {
		_force = "0"
	}
	c.registerEnvVar("ANSIBLE_DISPLAY_SKIPPED_HOSTS", "no")
	c.registerEnvVar("ANSIBLE_DISPLAY_OK_HOSTS", "yes")
	c.registerEnvVar("ANSIBLE_HOST_KEY_CHECKING", "no")
	c.registerEnvVar("ANSIBLE_ROLES_PATH", "/root/.ansible/roles:/usr/share/ansible/roles:/etc/ansible/roles")
	c.registerEnvVar("ANSIBLE_COLLECTIONS_PATH", "/root/.ansible/collections:/usr/share/ansible/collections:/etc/ansible/collections")
	c.registerEnvVar("ANSIBLE_VERBOSITY", logLevel)
	c.registerEnvVar("ANSIBLE_CALLBACKS_ENABLED", "timer,profile_tasks,profile_roles")
	c.registerEnvVar("POLYCRATE_CLI_VERSION", version)
	c.registerEnvVar("POLYCRATE_IMAGE_VERSION", imageVersion)
	c.registerEnvVar("POLYCRATE_IMAGE_REFERENCE", imageRef)
	c.registerEnvVar("POLYCRATE_FORCE", _force)
	c.registerEnvVar("POLYCRATE_VERSION", polycrateVersion)
	c.registerEnvVar("IN_CI", "true")
	c.registerEnvVar("TERM", "xterm-256color")

	if local {
		// Not in container
		c.registerEnvVar("POLYCRATE_WORKSPACE", workspaceDir)
		c.registerEnvVar("POLYCRATE_STACKFILE", workspaceConfigFilePath)
	} else {
		// In container
		c.registerEnvVar("POLYCRATE_WORKSPACE", workspaceContainerDir)
		c.registerEnvVar("POLYCRATE_STACKFILE", workspaceContainerConfigFilePath)
	}

	for _, envVar := range extraEnv {
		// Split by =
		p := strings.Split(envVar, "=")

		if len(p) == 2 {
			c.registerEnvVar(p[0], p[1])
		} else {
			c.err = goErrors.New("Illegal value for env var found: " + envVar)
			break
		}
	}
}

func (c *Workspace) registerEnvVar(key string, value string) {
	c.env[key] = value
}

func (c *Workspace) registerMount(host string, container string) {
	c.mounts[host] = container
}

func (c *Workspace) registerCurrentBlock(block *Block) {

	c.registerEnvVar("POLYCRATE_BLOCK", block.Metadata.Name)
	c.currentBlock = block
}
func (c *Workspace) registerCurrentAction(action *Action) {

	c.registerEnvVar("POLYCRATE_ACTION", action.Metadata.Name)
	c.currentAction = action
}
func (c *Workspace) registerCurrentWorkflow(workflow *Workflow) {

	c.registerEnvVar("POLYCRATE_WORKFLOW", workflow.Metadata.Name)
	c.currentWorkflow = workflow
}
func (c *Workspace) registerCurrentStep(step *Step) {

	c.registerEnvVar("POLYCRATE_STEP", step.Metadata.Name)
	c.currentStep = step
}

func (c *Workspace) getActionByPath(actionPath string) *Action {
	// Validate actionPath
	s := strings.Split(actionPath, ".")
	blockName := s[0]
	actionName := s[1]

	block := workspace.getBlockByName(blockName)

	if block != nil {
		action := block.getActionByName(actionName)

		if action != nil {
			return action
		}
	}
	return nil
}

func (c *Workspace) getBlockByPath(actionPath string) *Block {
	// Validate actionPath
	s := strings.Split(actionPath, ".")
	blockName := s[0]

	block := workspace.getBlockByName(blockName)

	if block != nil {
		return block
	}
	return nil
}

func (c *Workspace) getBlockByName(blockName string) *Block {
	for i := 0; i < len(c.Blocks); i++ {
		block := &c.Blocks[i]
		if block.Metadata.Name == blockName {
			return block
		}
	}
	return nil
}
func (c *Workspace) getWorkflowByName(workflowName string) *Workflow {

	for i := 0; i < len(c.Workflows); i++ {
		workflow := &c.Workflows[i]
		if workflow.Metadata.Name == workflowName {
			return workflow
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

		// Set Block vars
		loadedBlock.workdir.localPath = blockPath
		loadedBlock.workdir.containerPath = filepath.Join(workspaceContainerDir, blocksRoot, loadedBlock.Metadata.Name)

		// Check if Block exists
		existingBlock := c.getBlockByName(loadedBlock.Metadata.Name)

		if existingBlock != nil {
			// Block exists
			log.Debugf("Found existing Block '%s' in the Workspace. Merging with loaded Block.", existingBlock.Metadata.Name)

			err := existingBlock.MergeIn(&loadedBlock)
			if err != nil {
				return err
			}

			// Handle Actions
			// Iterate over the loaded Block's Actions
			// and check if it exists in the existing Block
			// If not, add it
			for _, loadedAction := range loadedBlock.Actions {
				log.Debugf("Analyzing Action '%s'", loadedAction.Metadata.Name)
				existingAction := existingBlock.getActionByName(loadedAction.Metadata.Name)

				if existingAction != nil {
					log.Debugf("Found existing Action '%s'", existingAction.Metadata.Name)
					// An Action with the same name exists
					// We merge!
					err := existingAction.MergeIn(&loadedAction)
					if err != nil {
						return err
					}
				} else {
					log.Debugf("No existing Action found. Adding '%s'", loadedAction.Metadata.Name)
					existingBlock.Actions = append(existingBlock.Actions, loadedAction)
				}
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
