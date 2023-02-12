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
	"bytes"
	"context"
	"errors"
	goErrors "errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"text/template"
	"time"

	"golang.org/x/sync/errgroup"

	//"github.com/docker/docker/container"
	"github.com/google/uuid"
	"github.com/jeremywohl/flatten"

	"github.com/go-playground/validator/v10"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

var workspaceCmd = &cobra.Command{
	Use:   "workspace",
	Short: "Manage the workspace",
	Long:  `Manage the workspace`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(workspaceCmd)
}

type ImageConfig struct {
	Reference string `yaml:"reference" mapstructure:"reference" json:"reference" validate:"required"`
	Version   string `yaml:"version" mapstructure:"version" json:"version" validate:"required"`
}

type Metadata struct {
	Name        string            `mapstructure:"name" json:"name" validate:"required,metadata_name"`
	Description string            `mapstructure:"description" json:"description"`
	Labels      map[string]string `mapstructure:"labels" json:"labels"`
	Alias       []string          `mapstructure:"alias" json:"alias"`
}

type WorkspaceContainerStatus struct {
	Name        string    `yaml:"name,omitempty" mapstructure:"name,omitempty" json:"name,omitempty"`
	Transaction uuid.UUID `yaml:"transaction,omitempty" mapstructure:"transaction,omitempty" json:"transaction,omitempty"`
	Running     bool
	Pruned      bool
}

type WorkspaceRevision struct {
	Command     string            `yaml:"command,omitempty" mapstructure:"command,omitempty" json:"command,omitempty"`
	UserEmail   string            `yaml:"user_email,omitempty" mapstructure:"user_email,omitempty" json:"user_email,omitempty"`
	UserName    string            `yaml:"user_name,omitempty" mapstructure:"user_name,omitempty" json:"user_name,omitempty"`
	Date        string            `yaml:"date,omitempty" mapstructure:"date,omitempty" json:"date,omitempty"`
	Transaction uuid.UUID         `yaml:"transaction,omitempty" mapstructure:"transaction,omitempty" json:"transaction,omitempty"`
	Version     string            `yaml:"version,omitempty" mapstructure:"version,omitempty" json:"version,omitempty"`
	ExitCode    int               `yaml:"exit_code,omitempty" mapstructure:"exit_code,omitempty" json:"exit_code,omitempty"`
	Output      string            `yaml:"output,omitempty" mapstructure:"output,omitempty" json:"output,omitempty"`
	Snapshot    WorkspaceSnapshot `yaml:"snapshot,omitempty" mapstructure:"snapshot,omitempty" json:"snapshot,omitempty"`
}

type WorkspaceConfig struct {
	Image      ImageConfig `yaml:"image" mapstructure:"image" json:"image" validate:"required"`
	BlocksRoot string      `yaml:"blocksroot" mapstructure:"blocksroot" json:"blocksroot" validate:"required"`
	// The block configuration file (default: block.poly)
	BlocksConfig    string                 `yaml:"blocksconfig" mapstructure:"blocksconfig" json:"blocksconfig" validate:"required"`
	WorkspaceConfig string                 `yaml:"workspaceconfig" mapstructure:"workspaceconfig" json:"workspaceconfig" validate:"required"`
	WorkflowsRoot   string                 `yaml:"workflowsroot" mapstructure:"workflowsroot" json:"workflowsroot" validate:"required"`
	ArtifactsRoot   string                 `yaml:"artifactsroot" mapstructure:"artifactsroot" json:"artifactsroot" validate:"required"`
	ContainerRoot   string                 `yaml:"containerroot" mapstructure:"containerroot" json:"containerroot" validate:"required"`
	SshPrivateKey   string                 `yaml:"sshprivatekey" mapstructure:"sshprivatekey" json:"sshprivatekey" validate:"required"`
	SshPublicKey    string                 `yaml:"sshpublickey" mapstructure:"sshpublickey" json:"sshpublickey" validate:"required"`
	RemoteRoot      string                 `yaml:"remoteroot" mapstructure:"remoteroot" json:"remoteroot" validate:"required"`
	Dockerfile      string                 `yaml:"dockerfile" mapstructure:"dockerfile,omitempty" json:"dockerfile,omitempty"`
	Globals         map[string]interface{} `yaml:"globals" mapstructure:"globals" json:"globals"`
}

type Workspace struct {
	Name        string            `yaml:"name,omitempty" mapstructure:"name,omitempty" json:"name,omitempty" validate:"required,metadata_name"`
	Description string            `yaml:"description,omitempty" mapstructure:"description,omitempty" json:"description,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty" mapstructure:"labels,omitempty" json:"labels,omitempty"`
	Alias       []string          `yaml:"alias,omitempty" mapstructure:"alias,omitempty" json:"alias,omitempty"`
	// Format: block:version
	Dependencies    []string               `yaml:"dependencies,omitempty" mapstructure:"dependencies,omitempty" json:"dependencies,omitempty"`
	Config          WorkspaceConfig        `yaml:"config,omitempty" mapstructure:"config,omitempty" json:"config,omitempty"`
	Blocks          []*Block               `yaml:"blocks,omitempty" mapstructure:"blocks,omitempty" json:"blocks,omitempty" validate:"dive,required"`
	Workflows       []*Workflow            `yaml:"workflows,omitempty" mapstructure:"workflows,omitempty" json:"workflows,omitempty"`
	ExtraEnv        []string               `yaml:"extraenv,omitempty" mapstructure:"extraenv,omitempty" json:"extraenv,omitempty"`
	ExtraMounts     []string               `yaml:"extramounts,omitempty" mapstructure:"extramounts,omitempty" json:"extramounts,omitempty"`
	Path            string                 `yaml:"path,omitempty" mapstructure:"path,omitempty" json:"path,omitempty"`
	LocalPath       string                 `yaml:"localpath,omitempty" mapstructure:"localpath,omitempty" json:"localpath,omitempty"`
	ContainerPath   string                 `yaml:"containerpath,omitempty" mapstructure:"containerpath,omitempty" json:"containerpath,omitempty"`
	Version         string                 `yaml:"version,omitempty" mapstructure:"version,omitempty" json:"version,omitempty"`
	Identifier      string                 `yaml:"identifier,omitempty" mapstructure:"identifier,omitempty" json:"identifier,omitempty"`
	Meta            map[string]interface{} `yaml:"meta,omitempty" mapstructure:"meta,omitempty" json:"meta,omitempty"`
	SyncOptions     SyncOptions            `yaml:"sync,omitempty" mapstructure:"sync,omitempty" json:"sync,omitempty"`
	loaded          bool
	currentBlock    *Block
	currentAction   *Action
	currentWorkflow *Workflow
	currentStep     *Step
	index           *WorkspaceIndex
	revision        *WorkspaceRevision
	env             map[string]string
	mounts          map[string]string
	err             error
	runtimeDir      string
	installedBlocks []*Block
	//overrides       []string
	synced bool
	//syncLoaded      bool
	syncNeeded      bool
	syncStatus      string
	isGitRepo       bool
	containerStatus WorkspaceContainerStatus
}

type WorkspaceIndex struct {
	Actions   map[string]*Action   `yaml:"actions" mapstructure:"actions" json:"actions"`
	Steps     map[string]*Step     `yaml:"steps" mapstructure:"steps" json:"steps"`
	Blocks    map[string]*Block    `yaml:"blocks" mapstructure:"blocks" json:"blocks"`
	Workflows map[string]*Workflow `yaml:"workflows" mapstructure:"workflows" json:"workflows"`
}

type WorkspaceSnapshot struct {
	Workspace *Workspace        `yaml:"workspace,omitempty" mapstructure:"workspace,omitempty" json:"workspace,omitempty"`
	Action    *Action           `yaml:"action,omitempty" mapstructure:"action,omitempty" json:"action,omitempty"`
	Block     *Block            `yaml:"block,omitempty" mapstructure:"block,omitempty" json:"block,omitempty"`
	Workflow  *Workflow         `yaml:"workflow,omitempty" mapstructure:"workflow,omitempty" json:"workflow,omitempty"`
	Step      *Step             `yaml:"step,omitempty" mapstructure:"step,omitempty" json:"step,omitempty"`
	Env       map[string]string `yaml:"env,omitempty" mapstructure:"env,omitempty" json:"env,omitempty"`
	Mounts    map[string]string `yaml:"mounts,omitempty" mapstructure:"mounts,omitempty" json:"mounts,omitempty"`
}

func (w *Workspace) CreateSshKeys(ctx context.Context) error {
	err := CreateSSHKeys(ctx, w.LocalPath, w.Config.SshPrivateKey, w.Config.SshPublicKey)
	if err != nil {
		return err
	}
	return nil
}

func (w *Workspace) FormatCommand(cmd *cobra.Command) string {

	commandPath := cmd.CommandPath()
	localArgs := cmd.Flags().Args()

	localFlags := []string{}

	cmd.Flags().Visit(func(flag *pflag.Flag) {
		//fmt.Printf("--%s=%s\n", flag.Name, flag.Value)
		localFlags = append(localFlags, fmt.Sprintf("--%s=%s", flag.Name, flag.Value))
	})

	command := strings.Join([]string{
		commandPath,
		strings.Join(localArgs, " "),
		strings.Join(localFlags, " "),
	}, " ")

	return command
}

func (w *Workspace) SaveRevision(ctx context.Context) error {
	log := polycrate.GetContextLogger(ctx)
	log.Debugf("Saving revision")

	f, err := os.OpenFile(filepath.Join(w.LocalPath, "revision.poly"), os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	// Export revision to yaml
	yaml, err := yaml.Marshal(w.revision)
	if err != nil {
		return err
	}

	// Write yaml export to file
	_, err = f.Write(yaml)
	if err != nil {
		return err
	}
	return nil
}

func (c *Workspace) RegisterSnapshotEnv(snapshot WorkspaceSnapshot) *Workspace {
	// create empty map to hold the flattened keys
	var jsonMap map[string]interface{}
	// err := mapstructure.Decode(snapshot, &jsonMap)
	// if err != nil {
	// 	panic(err)
	// }
	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	// marshal the snapshot into json
	jsonData, err := json.Marshal(snapshot)
	if err != nil {
		log.Errorf("Error marshalling: %s", err)
		c.err = err
		return c
	}

	// unmarshal the json into the previously created json map
	// flatten needs this input format: map[string]interface
	// which we achieve by first marshalling the struct (snapshot)
	// and then unmarshalling the resulting bytes into our json structure
	if err := json.Unmarshal(jsonData, &jsonMap); err != nil {
		log.Errorf("Error unmarshalling: %s", err)
		c.err = err
		return c
	}

	// flatten to key_key_0_key=value
	flat, err := flatten.Flatten(jsonMap, "", flatten.UnderscoreStyle)
	if err != nil {
		c.err = err
		return c
	}

	for key := range flat {
		keyString := fmt.Sprintf("%v", flat[key])
		//fmt.Printf("%s=%s\n", strings.ToUpper(key), keyString)
		c.registerEnvVar(strings.ToUpper(key), keyString)
	}

	return c
}

func (c *Workspace) Snapshot(ctx context.Context) {
	snapshot := c.GetSnapshot(ctx)
	printObject(snapshot)
	//convertToEnv(&snapshot)
}

func (w *Workspace) Inspect(ctx context.Context) {
	w.Print()
}

func (w *Workspace) Flush() *Workspace {
	if w.err != nil {
		log.Fatal(w.err)
	}
	return w
}

func (w *Workspace) RunActionWithContext(ctx context.Context, _block string, _action string) (context.Context, error) {
	//address := strings.Join([]string{block, action}, ":")
	// err := w.RunAction(ctx, address)
	// if err != nil {
	// 	return ctx, err
	// }

	var err error

	var block *Block
	ctx, block, err = w.GetBlockWithContext(ctx, _block)
	if err != nil {
		return ctx, err
	}

	var action *Action
	ctx, action, err = block.GetActionWithContext(ctx, _action)
	if err != nil {
		return ctx, err
	}

	if block.Template {
		err = goErrors.New("this is a template block. not running action")
		return ctx, err
	}

	//log := polycrate.GetContextLogger(ctx)

	w.registerCurrentAction(action)
	w.registerCurrentBlock(block)

	// If --snapshot is set, print the snapshot and exit
	if snapshot {
		w.Snapshot(ctx)
	} else {
		ctx, err := action.RunWithContext(ctx)
		if err != nil {
			return ctx, err
		}
	}

	// Reload Block after action execution to update artifacts, inventory and kubeconfig
	err = block.Reload(ctx, w)
	if err != nil {
		return ctx, err
	}

	// Run event handler
	//w.revision.Snapshot = WorkspaceSnapshot{}

	event, err := polycrate.GetContextEvent(ctx)
	if err == nil {
		event.Labels["monk.event.level"] = "Info"

		ctx = polycrate.SetContextEvent(ctx, event)
	}

	return ctx, nil
}

func (w *Workspace) RunAction(ctx context.Context, address string) error {
	log := polycrate.GetContextLogger(ctx)

	// Find action in index and report
	action := w.LookupAction(address)

	if action != nil {

		block := w.GetBlockFromIndex(action.Block)

		log = log.WithField("action", action.Name)
		log = log.WithField("block", block.Name)
		ctx = polycrate.SetContextLogger(ctx, log)

		w.registerCurrentAction(action)
		w.registerCurrentBlock(block)

		log.Debugf("Registering current block")

		log.Debugf("Registering current action")

		if block.Template {
			return goErrors.New("this is a template block. Not running action")
		}

		// Write log here
		if snapshot {
			w.Snapshot(ctx)
		} else {
			_, err := action.RunWithContext(ctx)
			if err != nil {
				return err
			}
			//sync.Log(fmt.Sprintf("Action %s of block %s was successful", action.Name, block.Name)).Flush()
		}

		// Reload Block after action execution to update artifacts, inventory and kubeconfig
		err := block.Reload(ctx, w)
		if err != nil {
			return err
		}

		// Run event handler
		//w.revision.Snapshot = WorkspaceSnapshot{}
		log.Debugf("Running event handler")
		event := NewEvent(ctx)
		event.Action = w.revision.Snapshot.Action.Name
		event.Block = w.revision.Snapshot.Block.Name
		event.Workspace = w.revision.Snapshot.Workspace.Name
		event.Command = w.revision.Command
		event.ExitCode = w.revision.ExitCode
		event.UserEmail = w.revision.UserEmail
		event.UserName = w.revision.UserName
		event.Date = w.revision.Date
		event.Output = w.revision.Output
		event.Labels["monk.event.level"] = "Info"

		if err := polycrate.EventHandler(ctx); err != nil {
			// We're not terminating here to not block further execution
			log.Warnf("Event handler failed: %s", err)
		}

	} else {
		return goErrors.New("cannot find Action with address " + address)
	}

	// After running an action we want to sync the workspace
	w.syncNeeded = true

	return nil
}

func (w *Workspace) ResolveBlock(ctx context.Context, block *Block, workspaceLocalPath string, workspaceContainerPath string) error {
	log := polycrate.GetContextLogger(ctx)
	log = log.WithField("block", block.Name)

	if !block.resolved {
		log.Tracef("Resolving block")

		// Check if a "from:" stanza is given and not empty
		// This means that the loadedBlock should inherit from another Block
		if block.From != "" {
			// a "from:" stanza is given
			log.WithField("dependency", block.From).Tracef("Dependency detected")

			// Try to load the referenced Block
			dependency := w.getBlockByName(block.From)

			if dependency == nil {
				//log.WithField("dependency", block.From).Errorf("Dependency not found")

				err := fmt.Errorf("dependency '%s' not found in the Workspace. Please check the 'from' stanza of block '%s'", block.From, block.Name)
				return err
			}

			log.WithField("dependency", block.From).Tracef("Dependency loaded")

			dep := dependency

			// Check if the dependency Block has already been resolved
			// If not, re-queue the loaded Block so it can be resolved in another iteration
			if !dep.resolved {
				// Needed Block from 'from' stanza is not yet resolved
				log.WithField("dependency", block.From).Tracef("Dependency not resolved. Deferring.")

				block.resolved = false
				err := ErrDependencyNotResolved
				return err
			}

			// Merge the dependency Block into the loaded Block
			// We do NOT OVERWRITE existing values in the loaded Block because we must assume
			// That this is configuration that has been explicitly set by the user
			// The merge works like "loading defaults" for the loaded Block
			err := block.MergeIn(dep)
			if err != nil {
				return err
			}

			// Handle Workdir
			block.Workdir.LocalPath = dep.Workdir.LocalPath
			block.Workdir.ContainerPath = dep.Workdir.ContainerPath

			log.WithField("dependency", block.From).Tracef("Dependency resolved")

		} else {
			log.WithField("dependency", block.From).Tracef("No dependency found")
		}

		// Validate schema
		err := block.ValidateSchema()
		if err != nil {
			return err
		}

		// Register the Block to the Index
		block.address = block.Name
		w.registerBlock(block)

		// Update Artifacts, Kubeconfig & Inventory
		err = block.LoadArtifacts(ctx, w)
		if err != nil {
			return err
		}
		err = block.LoadInventory(ctx, w)
		if err != nil {
			return err
		}
		err = block.LoadKubeconfig(ctx, w)
		if err != nil {
			return err
		}

		// Update Action addresses for all blocks not covered by dependencies
		if len(block.Actions) > 0 {

			for _, action := range block.Actions {
				existingAction := block.getActionByName(action.Name)
				log = log.WithField("action", existingAction.Name)

				actionAddress := strings.Join([]string{block.Name, existingAction.Name}, ".")
				if existingAction.address != actionAddress {
					existingAction.address = actionAddress
				}

				if existingAction.Block != block.Name {
					existingAction.Block = block.Name
				}

				err := existingAction.validate()
				if err != nil {
					return err
				}

				// Register the Action to the Index
				w.registerAction(existingAction)
			}
		}

		block.resolved = true
	}
	return nil
}

func (w *Workspace) PruneContainer(ctx context.Context) error {
	return polycrate.PruneContainer(ctx)
}

func (w *Workspace) RunContainerWithContext(ctx context.Context, name string, workdir string, cmd []string) (context.Context, error) {
	log := polycrate.GetContextLogger(ctx)
	txid := polycrate.GetContextTXID(ctx)

	log.Infof("Starting container")

	containerImage := strings.Join([]string{w.Config.Image.Reference, w.Config.Image.Version}, ":")

	// Check if a Dockerfile is configured in the Workspace
	if w.Config.Dockerfile != "" {
		// Create the filepath
		dockerfilePath := filepath.Join(w.LocalPath, w.Config.Dockerfile)

		// Check if the file exists
		if _, err := os.Stat(dockerfilePath); !os.IsNotExist(err) {
			if build {
				// We need to build and tag this
				log.Debugf("Dockerfile detected: %s", dockerfilePath)

				tag := w.Name + ":" + version
				log.Warnf("Building custom image: %s", tag)

				tags := []string{tag}
				containerImage, err = polycrate.BuildContainer(ctx, w.LocalPath, w.Config.Dockerfile, tags)
				if err != nil {
					return ctx, err
				}
			} else {
				if pull {
					err := polycrate.PullImage(ctx, containerImage)

					if err != nil {
						return ctx, err
					}
				} else {
					log.Debugf("Not pulling/building image")
				}
			}
		} else {
			if pull {
				err := polycrate.PullImage(ctx, containerImage)

				if err != nil {
					return ctx, err
				}
			} else {
				log.Debugf("Not pulling/building image")
			}
		}
	} else {
		if pull {
			err := polycrate.PullImage(ctx, containerImage)

			if err != nil {
				return ctx, err
			}
		} else {
			log.Debugf("Not pulling/building image")
		}
	}

	runCommand := cmd

	// Setup mounts
	mounts := []string{}
	for mount := range w.mounts {
		m := strings.Join([]string{mount, w.mounts[mount]}, ":")
		mounts = append(mounts, m)
	}

	// Setup labels
	labels := []string{}
	labels = append(labels, fmt.Sprintf("polycrate.txid=%s", txid))

	w.containerStatus.Running = true
	w.containerStatus.Transaction = w.revision.Transaction

	// Capture CTRL-C if the container is running and non-interactive
	var cancelHighjack context.CancelFunc
	var sigIntCtx context.Context
	if !interactive {
		sigIntCtx, cancelHighjack = context.WithCancel(ctx)
		defer cancelHighjack()
		polycrate.HighjackSigint(sigIntCtx)

	}
	log.Debugf("Starting container")

	ports := []string{}

	env := w.DumpEnv()

	containerName := polycrate.GetContextTXID(ctx).String()

	exitCode, output, err := polycrate.RunContainer(
		ctx,
		mounts,
		env,
		ports,
		containerName,
		labels,
		workdir,
		containerImage,
		runCommand,
	)

	ctx = polycrate.SetContextOutput(ctx, output)
	ctx = polycrate.SetContextExitCode(ctx, exitCode)

	if err != nil {
		return ctx, err
	}

	log.Debugf("Stopped container. Exit code: %d", exitCode)

	// Update revision
	revision, err := polycrate.GetContextRevision(ctx)
	if err == nil {
		revision.Output = output
		revision.ExitCode = exitCode
		revision.Snapshot = w.GetSnapshot(ctx)

		revisionKey := ContextKey("revision")
		ctx = context.WithValue(ctx, revisionKey, revision)
	}

	w.containerStatus.Running = false

	// Handle container error
	if err != nil {
		return ctx, err
	}

	// Prune container
	w.PruneContainer(ctx)

	return ctx, nil

	// cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	// if err != nil {
	// 	w.err = err
	// 	return w
	// }

	// log.WithFields(log.Fields{
	// 	"workspace": w.Name,
	// 	"build":     build,
	// 	"pull":      pull,
	// }).Debugf("Running container")

	// entrypoint := []string{"/bin/bash", "-c"}
	// runCommand := []string{}

	// if cmd != "" {
	// 	runCommand = append(runCommand, cmd)
	// } else {
	// 	w.err = goErrors.New("no execution script path given. Nothing to do")
	// 	return w
	// }

	// log.WithFields(log.Fields{
	// 	"workspace": w.Name,
	// }).Debugf("Running entrypoint %s", entrypoint)

	// log.WithFields(log.Fields{
	// 	"workspace": w.Name,
	// }).Debugf("Running command %s", runCommand)

	// cc := &container.Config{
	// 	Image:        containerImage,
	// 	Entrypoint:   entrypoint,
	// 	Cmd:          runCommand,
	// 	Tty:          true,
	// 	AttachStderr: true,
	// 	AttachStdout: true,
	// 	AttachStdin:  interactive,
	// 	Labels: map[string]string{
	// 		"polycrate.workspace": w.Name,
	// 		"polycrate.name":      name,
	// 	},
	// 	StdinOnce:  interactive,
	// 	OpenStdin:  interactive,
	// 	Env:        w.DumpEnv(),
	// 	WorkingDir: workdir,
	// }

	// // Setup mounts
	// containerMounts := []mount.Mount{}
	// for containerMount := range w.mounts {
	// 	m := mount.Mount{
	// 		Type:   mount.TypeBind,
	// 		Source: containerMount,
	// 		Target: workspace.mounts[containerMount],
	// 	}
	// 	containerMounts = append(containerMounts, m)
	// }

	// hc := &container.HostConfig{
	// 	Mounts: containerMounts,
	// }

	// // Ignore / Stop Signal channel
	// // if interactive {
	// // 	signal.Ignore(os.Interrupt)
	// // }

	// err = runContainer(cli, cc, hc, name)

	// if err != nil {
	// 	w.err = err
	// 	return w
	// }
	// return w
}

func (w *Workspace) RunWorkflowWithContext(ctx context.Context, name string, stepName string) (context.Context, error) {

	var err error

	var workflow *Workflow
	ctx, workflow, err = w.GetWorkflowWithContext(ctx, name)
	if err != nil {
		return ctx, err
	}

	w.registerCurrentWorkflow(workflow)

	if snapshot {
		w.Snapshot(ctx)
	} else {
		ctx, err := workflow.RunWithContext(ctx, stepName)
		if err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}
func (w *Workspace) RunWorkflow(ctx context.Context, name string) error {

	// Find workflow in index
	workflow := w.LookupWorkflow(name)

	if workflow != nil {
		w.registerCurrentWorkflow(workflow)

		if snapshot {
			w.Snapshot(ctx)
		} else {
			err := workflow.Run(ctx)
			if err != nil {
				return err
			}
		}
	} else {
		return goErrors.New("cannot find Workflow: " + name)
	}
	return nil
}

func (w *Workspace) RunStep(ctx context.Context, workflow string, name string) error {
	log := polycrate.GetContextLogger(ctx)
	log = log.WithField("workspace", workspace.Name)
	ctx = polycrate.SetContextLogger(ctx, log)

	// Find action in index and report
	step := w.LookupStep(name)

	if step != nil {
		if snapshot {
			w.Snapshot(ctx)
		} else {
			err := step.Run(ctx)
			if err != nil {
				return err
			}
		}
	} else {
		return goErrors.New("cannot find step: " + name)
	}
	return nil
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

func (w *Workspace) SoftloadWorkspaceConfig() *Workspace {
	var workspaceConfig = viper.NewWithOptions(viper.KeyDelimiter("::"))
	workspaceConfigFilePath := filepath.Join(w.Path, WorkspaceConfigFile)

	workspaceConfig.SetConfigType("yaml")
	workspaceConfig.SetConfigFile(workspaceConfigFilePath)

	err := workspaceConfig.MergeInConfig()
	if err != nil {
		w.err = err
		return w
	}

	if err := workspaceConfig.Unmarshal(&w); err != nil {
		w.err = err
		return w
	}

	return w
}

func (w *Workspace) LoadConfig() *Workspace {
	log.Warn("LoadConfig is deprecated")
	// This variable holds the configuration loaded from the workspace config file (e.g. workspace.poly)
	var workspaceConfig = viper.NewWithOptions(viper.KeyDelimiter("::"))

	// Match CLI Flags with Config options
	// CLI Flags have precedence

	workspaceConfig.BindPFlag("sync.local.branch.name", rootCmd.Flags().Lookup("sync-local-branch"))
	workspaceConfig.BindPFlag("sync.remote.branch.name", rootCmd.Flags().Lookup("sync-remote-branch"))
	workspaceConfig.BindPFlag("sync.remote.name", rootCmd.Flags().Lookup("sync-remote-name"))
	workspaceConfig.BindPFlag("sync.enabled", rootCmd.Flags().Lookup("sync-enabled"))
	workspaceConfig.BindPFlag("sync.auto", rootCmd.Flags().Lookup("sync-auto"))
	workspaceConfig.BindPFlag("config.registry.url", rootCmd.Flags().Lookup("registry-url"))
	workspaceConfig.BindPFlag("config.registry.baseimage", rootCmd.Flags().Lookup("registry-base-image"))
	workspaceConfig.BindPFlag("config.image.version", rootCmd.Flags().Lookup("image-version"))
	workspaceConfig.BindPFlag("config.image.reference", rootCmd.Flags().Lookup("image-ref"))
	workspaceConfig.BindPFlag("config.blocksroot", rootCmd.Flags().Lookup("blocks-root"))
	workspaceConfig.BindPFlag("config.blocksconfig", rootCmd.Flags().Lookup("blocks-config"))
	workspaceConfig.BindPFlag("config.workflowsroot", rootCmd.Flags().Lookup("workflows-root"))
	workspaceConfig.BindPFlag("config.workspaceconfig", rootCmd.Flags().Lookup("workspace-config"))
	workspaceConfig.BindPFlag("config.artifactsroot", rootCmd.Flags().Lookup("artifacts-root"))
	workspaceConfig.BindPFlag("config.containerroot", rootCmd.Flags().Lookup("container-root"))
	workspaceConfig.BindPFlag("config.remoteroot", rootCmd.Flags().Lookup("remote-root"))
	workspaceConfig.BindPFlag("config.sshprivatekey", rootCmd.Flags().Lookup("ssh-private-key"))
	workspaceConfig.BindPFlag("config.sshpublickey", rootCmd.Flags().Lookup("ssh-public-key"))
	workspaceConfig.BindPFlag("config.dockerfile", rootCmd.Flags().Lookup("dockerfile"))

	workspaceConfig.SetEnvPrefix(EnvPrefix)
	workspaceConfig.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	workspaceConfig.AutomaticEnv()

	// Check if a full path has been given
	workspaceConfigFilePath := filepath.Join(w.LocalPath, w.Config.WorkspaceConfig)

	if _, err := os.Stat(workspaceConfigFilePath); os.IsNotExist(err) {
		// The config file does not exist
		// We try to look in the list of workspaces in $HOME/.polycrate/workspaces

		// Assuming the "path" given is actually the name of a workspace
		workspaceName := w.LocalPath

		log.WithFields(log.Fields{
			"path":      workspaceConfigFilePath,
			"workspace": workspaceName,
		}).Debugf("Workspace config not found. Looking in the local workspace index")

		// Check if workspaceName exists in the local workspace index (see discoverWorkspaces())
		if localWorkspaceIndex[workspaceName] != "" {
			// We found a workspace with that name in the index
			path := localWorkspaceIndex[workspaceName]
			log.WithFields(log.Fields{
				"workspace": workspaceName,
				"path":      path,
			}).Debugf("Found workspace in the local workspace index")

			// Update the workspace config file path with the local workspace path from the index
			w.LocalPath = path
			workspaceConfigFilePath = filepath.Join(w.LocalPath, WorkspaceConfigFile)
		} else {
			w.err = fmt.Errorf("couldn't find workspace config at %s", workspaceConfigFilePath)
			return w
		}
	}

	workspaceConfig.SetConfigType("yaml")
	workspaceConfig.SetConfigFile(workspaceConfigFilePath)

	err := workspaceConfig.MergeInConfig()
	if err != nil {
		w.err = err
		return w
	}

	if err := workspaceConfig.Unmarshal(&w); err != nil {
		w.err = err
		return w
	}

	if err := w.validate(); err != nil {
		w.err = err
		return w
	}

	// set runtime dir
	w.runtimeDir = filepath.Join(polycrateRuntimeDir, w.Name)

	return w
}
func (w *Workspace) LoadConfigFromFile(ctx context.Context, path string, validate bool) error {
	// This variable holds the configuration loaded from the workspace config file (e.g. workspace.poly)
	var workspaceConfig = viper.NewWithOptions(viper.KeyDelimiter("::"))
	log := polycrate.GetContextLogger(ctx)

	// Match CLI Flags with Config options
	// CLI Flags have precedence

	workspaceConfig.BindPFlag("sync.local.branch.name", rootCmd.Flags().Lookup("sync-local-branch"))
	workspaceConfig.BindPFlag("sync.remote.branch.name", rootCmd.Flags().Lookup("sync-remote-branch"))
	workspaceConfig.BindPFlag("sync.remote.name", rootCmd.Flags().Lookup("sync-remote-name"))
	workspaceConfig.BindPFlag("sync.enabled", rootCmd.Flags().Lookup("sync-enabled"))
	workspaceConfig.BindPFlag("sync.auto", rootCmd.Flags().Lookup("sync-auto"))
	workspaceConfig.BindPFlag("config.registry.url", rootCmd.Flags().Lookup("registry-url"))
	workspaceConfig.BindPFlag("config.registry.baseimage", rootCmd.Flags().Lookup("registry-base-image"))
	workspaceConfig.BindPFlag("config.image.version", rootCmd.Flags().Lookup("image-version"))
	workspaceConfig.BindPFlag("config.image.reference", rootCmd.Flags().Lookup("image-ref"))
	workspaceConfig.BindPFlag("config.blocksroot", rootCmd.Flags().Lookup("blocks-root"))
	workspaceConfig.BindPFlag("config.blocksconfig", rootCmd.Flags().Lookup("blocks-config"))
	workspaceConfig.BindPFlag("config.workflowsroot", rootCmd.Flags().Lookup("workflows-root"))
	//workspaceConfig.BindPFlag("config.workspaceconfig", rootCmd.Flags().Lookup("workspace-config"))
	workspaceConfig.BindPFlag("config.artifactsroot", rootCmd.Flags().Lookup("artifacts-root"))
	workspaceConfig.BindPFlag("config.containerroot", rootCmd.Flags().Lookup("container-root"))
	workspaceConfig.BindPFlag("config.remoteroot", rootCmd.Flags().Lookup("remote-root"))
	workspaceConfig.BindPFlag("config.sshprivatekey", rootCmd.Flags().Lookup("ssh-private-key"))
	workspaceConfig.BindPFlag("config.sshpublickey", rootCmd.Flags().Lookup("ssh-public-key"))
	workspaceConfig.BindPFlag("config.dockerfile", rootCmd.Flags().Lookup("dockerfile"))

	// workspaceConfig.SetEnvPrefix(EnvPrefix)
	// workspaceConfig.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	// workspaceConfig.AutomaticEnv()

	// Check if a full path has been given
	workspaceConfigFilePath := filepath.Join(w.LocalPath, w.Config.WorkspaceConfig)

	if _, err := os.Stat(workspaceConfigFilePath); os.IsNotExist(err) {
		// The config file does not exist
		// We try to look in the list of workspaces in $HOME/.polycrate/workspaces

		// Assuming the "path" given is actually the name of a workspace
		workspaceName := w.LocalPath

		log.Debugf("Workspace config not found. Looking in the local workspace index")

		// Check if workspaceName exists in the local workspace index (see discoverWorkspaces())
		if localWorkspaceIndex[workspaceName] != "" {
			// We found a workspace with that name in the index
			path := localWorkspaceIndex[workspaceName]
			log.Debugf("Found workspace in the local workspace index")

			// Update the workspace config file path with the local workspace path from the index
			w.LocalPath = path
			workspaceConfigFilePath = filepath.Join(w.LocalPath, WorkspaceConfigFile)
		} else {
			// err = fmt.Errorf("couldn't find workspace config at %s", workspaceConfigFilePath)
			return ErrWorkspaceConfigNotFound
		}
	}

	workspaceConfig.SetConfigType("yaml")
	workspaceConfig.SetConfigFile(workspaceConfigFilePath)

	err := workspaceConfig.MergeInConfig()
	if err != nil {
		return err
	}

	if err := workspaceConfig.Unmarshal(&w); err != nil {
		return err
	}

	if validate {
		errors, err := w.Validate(ctx)
		if err != nil {
			for errorString := range errors {
				log.Error(errorString)
			}
			return err
		}
	} else {
		errors, err := w.Validate(ctx)
		if err != nil {
			log.Warn("You have validation errors in your workspace")
			for _, errorString := range errors {
				fmt.Printf("%T\n", errorString)
				log.Warn(errorString)
			}
		}
	}

	// set runtime dir
	txid := polycrate.GetContextTXID(ctx)
	w.runtimeDir = filepath.Join(polycrateRuntimeDir, txid.String(), w.Name)

	return nil
}

func (w *Workspace) Save(ctx context.Context) error {
	workspaceConfigFilePath := filepath.Join(w.LocalPath, w.Config.WorkspaceConfig)

	if _, err := os.Stat(workspaceConfigFilePath); !os.IsNotExist(err) {
		return fmt.Errorf("config file already exists: %s", workspaceConfigFilePath)
	}

	var _w Workspace
	_w.Name = w.Name

	yamlBytes, err := yaml.Marshal(_w)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(workspaceConfigFilePath, yamlBytes, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (c *Workspace) updateConfig(path string, value string) *Workspace {
	var sideloadConfig = viper.NewWithOptions(viper.KeyDelimiter("::"))
	//var sideloadStruct Workspace

	// Check if a full path has been given
	workspaceConfigFilePath := filepath.Join(c.LocalPath, workspace.Config.WorkspaceConfig)

	log.WithFields(log.Fields{
		"workspace": c.Name,
		"key":       path,
		"value":     value,
		"path":      workspaceConfigFilePath,
	}).Debugf("Updating workspace config")

	if _, err := os.Stat(workspaceConfigFilePath); os.IsNotExist(err) {
		c.err = fmt.Errorf("couldn't find workspace config at %s: %s", workspaceConfigFilePath, err)
		return c
	}

	yamlFile, err := ioutil.ReadFile(workspaceConfigFilePath)
	if err != nil {
		c.err = err
		return c
	}

	sideloadConfig.SetConfigType("yaml")
	sideloadConfig.SetConfigName("workspace")
	sideloadConfig.SetConfigFile(workspaceConfigFilePath)
	//sideloadConfig.ReadInConfig()

	err = sideloadConfig.ReadConfig(bytes.NewBuffer(yamlFile))
	if err != nil {
		c.err = err
		return c
	}

	// Update here
	sideloadConfig.Set(path, value)

	// if err := sideloadConfig.Unmarshal(&sideloadStruct); err != nil {
	// 	c.err = err
	// 	return c
	// }

	// if err := sideloadStruct.validate(); err != nil {
	// 	c.err = err
	// 	return c
	// }

	// Write back
	s := sideloadConfig.AllSettings()
	bs, err := yaml.Marshal(s)
	if err != nil {
		c.err = err
		return c
	}

	err = ioutil.WriteFile(workspaceConfigFilePath, bs, 0)
	if err != nil {
		c.err = err
		return c
	}
	return c
}

func (w *Workspace) Reload(ctx context.Context, validate bool) (*Workspace, error) {
	log := polycrate.GetContextLogger(ctx)
	log = log.WithField("workspace", w.Name)
	ctx = polycrate.SetContextLogger(ctx, log)

	log.Debug("Reloading workspace")

	return w.Load(ctx, w.LocalPath, validate)
}

func (w *Workspace) HighjackSigint(ctx context.Context) {
	log := polycrate.GetContextLogger(ctx)
	log.Debugf("Starting signal handler")
	signals := make(chan os.Signal, 1)

	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signals

		log.Debugf("Received CTRL-C")
		w.PruneContainer(ctx)

		//signalHandler(s)
		//workspace.Sync().Flush()

		log.Fatalf("ctrl-c received")

	}()
}

func (w *Workspace) Preload(ctx context.Context, path string, validate bool) (*Workspace, error) {
	var err error

	// Reset blocks
	w.Blocks = []*Block{}

	// Check if this is a git repo
	w.isGitRepo = GitIsRepo(ctx, path)

	// Load Workspace config (e.g. workspace.poly)
	err = w.LoadConfigFromFile(ctx, path, validate)
	if err != nil {
		return nil, err
	}

	// Setup Logger
	log := polycrate.GetContextLogger(ctx)
	log = log.WithField("workspace", w.Name)
	ctx = polycrate.SetContextLogger(ctx, log)
	txid := polycrate.GetContextTXID(ctx)

	// Set workspace.Path depending on --local
	w.ContainerPath = filepath.Join([]string{w.Config.ContainerRoot}...)
	if local {
		w.Path = w.LocalPath
	} else {
		w.Path = w.ContainerPath
	}

	// Bootstrap the workspace index
	if w.index == nil {
		w.index = &WorkspaceIndex{}
		w.index.Actions = make(map[string]*Action)
		w.index.Blocks = make(map[string]*Block)
		w.index.Workflows = make(map[string]*Workflow)
		w.index.Steps = make(map[string]*Step)
	}

	// Find all blocks in the workspace
	log.Debugf("Searching for installed blocks")
	blocksDir := filepath.Join(w.LocalPath, w.Config.BlocksRoot)
	if err := w.FindInstalledBlocks(ctx, blocksDir); err != nil {
		return nil, err
	}

	// Load all discovered blocks in the workspace
	log.Debugf("Loading installed blocks")
	if err := w.LoadInstalledBlocks(ctx); err != nil {
		return nil, err
	}

	// Bootstrap revision data
	// TODO: deprecate
	w.revision = &WorkspaceRevision{}
	w.revision.Date = time.Now().Format(time.RFC3339)
	w.revision.Command = w.FormatCommand(globalCmd)
	w.revision.Transaction = txid
	w.revision.Version = w.Version

	userInfo := polycrate.GetUserInfo(ctx)
	w.revision.UserEmail = userInfo["email"]
	w.revision.UserName = userInfo["name"]

	return w, nil
}

func (w *Workspace) Load(ctx context.Context, path string, validate bool) (*Workspace, error) {
	var err error

	w, err = w.Preload(ctx, path, validate)
	if err != nil {
		return nil, err
	}

	// Resolve block dependencies
	log.Debugf("Resolving block dependencies")
	if err := w.ResolveBlockDependencies(ctx); err != nil {
		return nil, err
	}

	// Update workflow and step addresses
	log.Debugf("Resolving workflows")
	if err := w.ResolveWorkflows(ctx); err != nil {
		return nil, err
	}

	// Bootstrap env vars
	w.bootstrapEnvVars(ctx).Flush()

	// Bootstrap container mounts
	if err := w.bootstrapMounts(); err != nil {
		return nil, err
	}

	// Template action scripts
	w.templateActionScripts().Flush()

	// Mark workspace as loaded
	w.loaded = true

	// if sync.Options.Enabled && sync.Options.Auto {
	// 	sync.Sync().Flush()
	// 	// Commented out, takes too much time, a commit is enough
	// 	//sync.Commit("Workspace loaded").Flush()
	// }

	// log.Debugf("Loading sync module")
	// if err := w.LoadSync(ctx); err != nil {
	// 	return nil, err
	// }

	return w, nil
	//os.Exit(1)
}

// func (w *Workspace) load() *Workspace {
// 	log.Fatal("DEPRECATED")
// 	// Return the workspace if it has been already loaded
// 	if w.loaded {
// 		return w
// 	}
// 	ctx := context.Background()

// 	log.WithFields(log.Fields{
// 		"path": w.LocalPath,
// 	}).Debugf("Loading workspace")

// 	// Load Workspace config (e.g. workspace.poly)
// 	//w.LoadWorkspaceConfig().Flush()
// 	err := w.LoadConfigFromFile(ctx, w.LocalPath, true)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// Set workspace.Path depending on --local
// 	if local {
// 		w.Path = w.LocalPath
// 	} else {
// 		w.Path = w.ContainerPath
// 	}

// 	// Save revision data
// 	w.revision = &WorkspaceRevision{}
// 	w.revision.Date = time.Now().Format(time.RFC3339)
// 	w.revision.Command = w.FormatCommand(globalCmd)
// 	w.revision.Transaction = uuid.New()
// 	w.revision.Version = w.Version

// 	w.revision.UserEmail, _ = GitGetUserEmail(ctx)
// 	w.revision.UserName, _ = GitGetUserName(ctx)

// 	// Cleanup workspace runtime dir
// 	w.Cleanup().Flush()

// 	// Bootstrap the workspace index
// 	w.BootstrapIndex().Flush()

// 	// Find all blocks in the workspace
// 	w.DiscoverInstalledBlocks().Flush()

// 	// Pull dependencies
// 	w.PullDependencies().Flush()

// 	// Load all discovered blocks in the workspace
// 	w.ImportInstalledBlocks(ctx).Flush()

// 	// Resolve block dependencies
// 	if err := w.ResolveBlockDependencies(ctx, w.LocalPath, w.ContainerPath); err != nil {
// 		log.Fatal(err)
// 	}

// 	// Update workflow and step addresses
// 	w.resolveWorkflows().Flush()

// 	// Bootstrap env vars
// 	w.bootstrapEnvVars(ctx).Flush()

// 	// Bootstrap container mounts
// 	w.bootstrapMounts()

// 	// Template action scripts
// 	w.templateActionScripts().Flush()

// 	log.WithFields(log.Fields{
// 		"workspace": w.Name,
// 		"blocks":    len(workspace.Blocks),
// 		"workflows": len(workspace.Workflows),
// 	}).Debugf("Workspace ready")

// 	// Mark workspace as loaded
// 	w.loaded = true

// 	// if sync.Options.Enabled && sync.Options.Auto {
// 	// 	sync.Sync().Flush()
// 	// 	// Commented out, takes too much time, a commit is enough
// 	// 	//sync.Commit("Workspace loaded").Flush()
// 	// }

// 	// if err := w.LoadSync(ctx); err != nil {
// 	// 	log.Fatal(err)
// 	// }

// 	return w
// }

// func (w *Workspace) SyncLoadRepo() *Workspace {
// 	path := w.LocalPath
// 	var err error
// 	// Check if it's a git repo already
// 	log.WithFields(log.Fields{
// 		"path": path,
// 	}).Debugf("Loading local repository")

// 	if GitIsRepo(path) {
// 		// It's a git repo
// 		// 1. Get repo's remote
// 		// 2. Compare with configured remote
// 		// 2.1 No remote configured? Update configured remote with repo's remote
// 		// 2.2 No repo remot? Update with configured remote
// 		// 2.3 Unequal? Update repo remote with configured remote

// 		// Check remote
// 		remoteUrl, err := GitGetRemoteUrl(path, GitDefaultRemote)
// 		if err != nil {
// 			w.err = err
// 			return w
// 		}
// 		if remoteUrl == "" {
// 			log.WithFields(log.Fields{
// 				"path": path,
// 			}).Debugf("Local repository has no remote url configured")

// 			// Check if workspace has a remote url configured
// 			if w.SyncOptions.Remote.Url != "" {
// 				// Create the remote from the workspace config
// 				err := GitCreateRemote(path, GitDefaultRemote, w.SyncOptions.Remote.Url)
// 				if err != nil {
// 					w.err = err
// 					return w
// 				}
// 			} else {
// 				// Exit with error - workspace.SyncOptions.Remote.Url is not configured
// 				w.err = fmt.Errorf("workspace has no remote configured")
// 				return w
// 			}
// 		} else {
// 			// Remote is already configured
// 			// Check if workspace has a remote url configured
// 			if w.SyncOptions.Remote.Url != "" {
// 				// Check if its url matches the configured remote url

// 				if remoteUrl != w.SyncOptions.Remote.Url {
// 					// Urls don't match
// 					// Update the repository with the configured remote
// 					log.WithFields(log.Fields{
// 						"path":      path,
// 						"workspace": w.Name,
// 					}).Debugf("Local repository remote url doesn't match workspace remote url. Fixing.")

// 					err := GitUpdateRemoteUrl(path, GitDefaultRemote, w.SyncOptions.Remote.Url)
// 					if err != nil {
// 						w.err = err
// 						return w
// 					}
// 				}
// 			} else {
// 				// Update the workspace remote with the local remote
// 				log.WithFields(log.Fields{
// 					"path": path,
// 				}).Debugf("Workspace has no remote url configured. Updating with local repository remote url")
// 				log.WithFields(log.Fields{
// 					"path": path,
// 				}).Warnf("Updating workspace remote url with local repository remote url")
// 				w.updateConfig("sync.remote.url", remoteUrl).Flush()
// 			}
// 		}
// 		log.WithFields(log.Fields{
// 			"path":      path,
// 			"workspace": w.Name,
// 			"remote":    w.SyncOptions.Remote.Name,
// 			"branch":    w.SyncOptions.Remote.Branch.Name,
// 		}).Debugf("Tracking remote branch")
// 		_, err = GitSetUpstreamTracking(path, w.SyncOptions.Remote.Name, w.SyncOptions.Remote.Branch.Name)
// 		if err != nil {
// 			w.err = err
// 			return w
// 		}
// 	} else {
// 		// Not a git repo
// 		// Check if a remote url is configured

// 		if w.SyncOptions.Remote.Url != "" {
// 			// We have a remote url configured
// 			// Create a repository with the given url
// 			log.WithFields(log.Fields{
// 				"path": path,
// 				"url":  w.SyncOptions.Remote.Url,
// 			}).Debugf("Creating new repository with remote url from workspace config")

// 			err = GitCreateRepo(path, w.SyncOptions.Remote.Name, w.SyncOptions.Remote.Branch.Name, w.SyncOptions.Remote.Url)
// 			if err != nil {
// 				w.err = err
// 				return w
// 			}
// 		} else {
// 			// No remote url configured
// 			log.WithFields(log.Fields{
// 				"path": path,
// 			}).Warnf("Workspace has no remote url configured.")
// 			w.err = errors.New("cannot sync this repository. No remote configured in workspace or repository")
// 			return w
// 		}
// 	}

// 	log.WithFields(log.Fields{
// 		"path": path,
// 	}).Debugf("Local repository loaded")

// 	return w
// }

// func (w *Workspace) LoadSyncRepo(ctx context.Context) error {
// 	path := w.LocalPath
// 	var err error
// 	log := polycrate.GetContextLogger(ctx)
// 	log = log.WithField("path", path)

// 	if GitIsRepo(path) {
// 		// It's a git repo
// 		// 1. Get repo's remote
// 		// 2. Compare with configured remote
// 		// 2.1 No remote configured? Update configured remote with repo's remote
// 		// 2.2 No repo remot? Update with configured remote
// 		// 2.3 Unequal? Update repo remote with configured remote

// 		// Check remote
// 		remoteUrl, err := GitGetRemoteUrl(path, GitDefaultRemote)
// 		if err != nil {
// 			return err
// 		}
// 		if remoteUrl == "" {
// 			log.Tracef("Local repository has no remote url configured")

// 			// Check if workspace has a remote url configured
// 			if w.SyncOptions.Remote.Url != "" {
// 				// Create the remote from the workspace config
// 				err := GitCreateRemote(path, GitDefaultRemote, w.SyncOptions.Remote.Url)
// 				if err != nil {
// 					return err
// 				}
// 			} else {
// 				// Exit with error - workspace.SyncOptions.Remote.Url is not configured
// 				err = fmt.Errorf("workspace has no remote configured")
// 				return err
// 			}
// 		} else {
// 			// Remote is already configured
// 			// Check if workspace has a remote url configured
// 			if w.SyncOptions.Remote.Url != "" {
// 				// Check if its url matches the configured remote url

// 				if remoteUrl != w.SyncOptions.Remote.Url {
// 					// Urls don't match
// 					// Update the repository with the configured remote
// 					log.Tracef("Local repository remote url doesn't match workspace remote url. Fixing.")

// 					err := GitUpdateRemoteUrl(path, GitDefaultRemote, w.SyncOptions.Remote.Url)
// 					if err != nil {
// 						return err
// 					}
// 				}
// 			} else {
// 				// Update the workspace remote with the local remote
// 				log.Tracef("Workspace has no remote url configured. Updating with local repository remote url")
// 				log.Warnf("Updating workspace remote url with local repository remote url")
// 				w.updateConfig("sync.remote.url", remoteUrl).Flush()
// 			}
// 		}

// 		log.WithField("remote", w.SyncOptions.Remote.Name).WithField("branch", w.SyncOptions.Remote.Branch.Name).Tracef("Tracking remote branch")

// 		_, err = GitSetUpstreamTracking(path, w.SyncOptions.Remote.Name, w.SyncOptions.Remote.Branch.Name)
// 		if err != nil {
// 			return err
// 		}
// 	} else {
// 		// Not a git repo
// 		// Check if a remote url is configured

// 		if w.SyncOptions.Remote.Url != "" {
// 			// We have a remote url configured
// 			// Create a repository with the given url
// 			log.Tracef("Creating new repository with remote url from workspace config")

// 			err = GitCreateRepo(path, w.SyncOptions.Remote.Name, w.SyncOptions.Remote.Branch.Name, w.SyncOptions.Remote.Url)
// 			if err != nil {
// 				return err
// 			}
// 		} else {
// 			// No remote url configured
// 			err = errors.New("cannot sync this repository. No remote configured in workspace or repository")
// 			return err
// 		}
// 	}
// 	return err
// }

// func (w *Workspace) LoadSync(ctx context.Context) error {
// 	if w.SyncOptions.Enabled {

// 		err := w.LoadSyncRepo(ctx)
// 		if err != nil {
// 			return err
// 		}
// 		w.syncLoaded = true
// 	} else {
// 		w.syncLoaded = false
// 	}
// 	return nil
// }

func (w *Workspace) Cleanup() *Workspace {

	// log.WithFields(log.Fields{
	// 	"workspace": w.Name,
	// }).Debugf("Cleaning orphaned containers")
	// workspace.PruneContainer().Flush()

	// Create runtime dir
	log.WithFields(log.Fields{
		"workspace": w.Name,
		"path":      w.runtimeDir,
	}).Debugf("Creating runtime directory")

	err := os.MkdirAll(w.runtimeDir, os.ModePerm)
	if err != nil {
		w.err = err
		return w
	}

	// Purge all contents of runtime dir
	dir, err := ioutil.ReadDir(w.runtimeDir)
	if err != nil {
		w.err = err
		return w
	}
	for _, d := range dir {
		log.WithFields(log.Fields{
			"workspace": w.Name,
			"path":      w.runtimeDir,
			"file":      d.Name(),
		}).Debugf("Cleaning runtime directory")
		err := os.RemoveAll(filepath.Join([]string{w.runtimeDir, d.Name()}...))
		if err != nil {
			w.err = err
			return w
		}
	}
	return w
}

func (c *Workspace) Uninstall() error {
	// e.g. $HOME/.polycrate/workspaces/workspace-1/artifacts/blocks/block-1
	if _, err := os.Stat(c.LocalPath); os.IsNotExist(err) {
		log.WithFields(log.Fields{
			"workspace": c.Name,
			"path":      c.LocalPath,
		}).Debugf("Workspace directory does not exist")
	} else {
		log.WithFields(log.Fields{
			"workspace": c.Name,
			"path":      c.LocalPath,
		}).Debugf("Removing workspace directory")
		err := os.RemoveAll(c.LocalPath)
		if err != nil {
			return err
		}
	}
	log.WithFields(log.Fields{
		"workspace": c.Name,
		"path":      c.LocalPath,
	}).Debugf("Successfully uninstalled workspace")
	return nil

}

// Resolves the 'from:' stanza of all blocks
func (w *Workspace) ResolveBlockDependencies(ctx context.Context) error {
	log := polycrate.GetContextLogger(ctx)

	missing := len(w.Blocks)

	// Iterate over all Blocks in the Workspace
	// Until nothing is "missing" anymore
	for missing > 0 {
		log.Debugf("Unresolved blocks: %d", missing)
		for i := 0; i < len(w.Blocks); i++ {
			loadedBlock := w.Blocks[i]

			log.Tracef("Resolving block %s - resolved? %s", loadedBlock.Name, loadedBlock.resolved)

			if !loadedBlock.resolved {
				err := w.ResolveBlock(ctx, loadedBlock, w.LocalPath, w.ContainerPath)

				if err != nil {
					switch {
					case errors.Is(err, ErrDependencyNotResolved):
						loadedBlock.resolved = false
						loadedBlock.err = nil
						continue
					default:
						return err
					}
				}
			}
			missing--
		}
	}
	return nil
}

func (w *Workspace) resolveWorkflows() *Workspace {
	log.WithFields(log.Fields{
		"workspace": w.Name,
	}).Debugf("Resolving workflows")

	for i := 0; i < len(w.Workflows); i++ {
		loadedWorkflow := w.Workflows[i]

		loadedWorkflow.address = loadedWorkflow.Name
		// Register the workflow to the Index
		w.registerWorkflow(loadedWorkflow)

		// Loop over the steps
		for _, step := range loadedWorkflow.Steps {
			loadedStep := loadedWorkflow.getStepByName(step.Name)

			// Set step address
			loadedStep.address = strings.Join([]string{loadedWorkflow.Name, loadedStep.Name}, ".")
			loadedStep.Workflow = loadedWorkflow.Name

			log.WithFields(log.Fields{
				"workspace": w.Name,
				"workflow":  loadedWorkflow.Name,
				"step":      loadedStep.Name,
			}).Tracef("Validating step")
			if err := loadedStep.validate(); err != nil {
				w.err = err
				return w
			}

			log.WithFields(log.Fields{
				"workspace": w.Name,
				"workflow":  loadedWorkflow.Name,
				"step":      loadedStep.Name,
			}).Tracef("Registering step")
			w.registerStep(loadedStep)
		}

		log.WithFields(log.Fields{
			"workspace": w.Name,
			"workflow":  loadedWorkflow.Name,
		}).Debugf("Validating workflow")
		if err := loadedWorkflow.validate(); err != nil {
			w.err = err
			return w
		}

	}
	return w
}
func (w *Workspace) ResolveWorkflows(ctx context.Context) error {
	for i := 0; i < len(w.Workflows); i++ {
		loadedWorkflow := w.Workflows[i]

		loadedWorkflow.address = loadedWorkflow.Name
		// Register the workflow to the Index
		w.registerWorkflow(loadedWorkflow)

		// Loop over the steps
		for _, step := range loadedWorkflow.Steps {
			loadedStep := loadedWorkflow.getStepByName(step.Name)

			// Set step address
			loadedStep.address = strings.Join([]string{loadedWorkflow.Name, loadedStep.Name}, ".")
			loadedStep.Workflow = loadedWorkflow.Name

			if err := loadedStep.validate(); err != nil {
				return err
			}
			w.registerStep(loadedStep)
		}

		if err := loadedWorkflow.validate(); err != nil {
			return err
		}

	}
	return nil
}

func (c *Workspace) Validate(ctx context.Context) ([]string, error) {
	err := validate.Struct(c)
	errors := []string{}

	if err != nil {
		// this check is only needed when your code could produce
		// an invalid value for validation such as interface with nil
		// value most including myself do not usually have code like this.
		if _, ok := err.(*validator.InvalidValidationError); ok {
			log.Error("Encountered problematic validation error")
			log.Error(err)
			return errors, nil
		}

		for _, err := range err.(validator.ValidationErrors) {
			var tag string = err.Tag()
			var namespace string = strings.ToLower(err.Namespace())
			var errorString string
			if tag == "metadata_name" {
				tag = "malformed"
				errorString = strings.Join([]string{"Validation error:", "option", namespace, "is", fmt.Sprintf("malformed: '%s'", err.Value()), fmt.Sprintf("(regex: `%s`)", ValidateMetaDataNameRegex)}, " ")
			} else {
				errorString = strings.Join([]string{"Validation error:", "option", strings.ToLower(err.Namespace()), "is", tag}, " ")
			}
			errors = append(errors, errorString)
		}

		// from here you can create your own error messages in whatever language you wish
		return errors, fmt.Errorf("error validating Workspace")
	}
	return errors, nil
}

func (c *Workspace) validate() error {
	log.WithFields(log.Fields{
		"workspace": c.Name,
	}).Debugf("Validating workspace")
	err := validate.Struct(c)

	if err != nil {
		log.WithFields(log.Fields{
			"workspace": c.Name,
		}).Errorf("Error validating workspace")

		// this check is only needed when your code could produce
		// an invalid value for validation such as interface with nil
		// value most including myself do not usually have code like this.
		if _, ok := err.(*validator.InvalidValidationError); ok {
			log.Error(err)
			return nil
		}

		for _, err := range err.(validator.ValidationErrors) {
			log.WithFields(log.Fields{
				"workspace": c.Name,
				"option":    strings.ToLower(err.Namespace()),
				"error":     err.Tag(),
			}).Errorf("Validation error")
		}

		// from here you can create your own error messages in whatever language you wish
		return goErrors.New("error validating Workspace")
	}
	return nil
}

func (c *Workspace) ListBlocks() *Workspace {
	for _, block := range c.Blocks {
		str := block.Name

		if block.From != "" {
			str = str + " (from: " + block.From + ")"
		}
		fmt.Println(str)
	}
	return c
}

func (c *Workspace) ListWorkflows() error {
	for workflow := range c.index.Workflows {
		fmt.Println(workflow)
	}
	return nil
}

func (c *Workspace) ListActions() *Workspace {
	for action := range c.index.Actions {
		fmt.Println(action)
	}
	return c
}

func (w *Workspace) BootstrapIndex() *Workspace {
	w.index = &WorkspaceIndex{}
	w.index.Actions = make(map[string]*Action)
	w.index.Blocks = make(map[string]*Block)
	w.index.Workflows = make(map[string]*Workflow)
	w.index.Steps = make(map[string]*Step)
	return w
}

func (c *Workspace) bootstrapMounts() error {
	// Prepare mounts map
	c.mounts = map[string]string{}

	// Mount the Workspace
	c.registerMount(c.LocalPath, c.ContainerPath)

	// Mount the Docker Socket if possible
	dockerSocket := "/var/run/docker.sock"
	if _, err := os.Stat(dockerSocket); !os.IsNotExist(err) {
		c.registerMount(dockerSocket, dockerSocket)
	}

	for _, extraMount := range c.ExtraMounts {
		// Split by :
		p := strings.Split(extraMount, ":")

		if len(p) == 2 {
			c.registerMount(p[0], p[1])
		} else {
			return goErrors.New("Illegal value for mount found: " + extraMount)
		}
	}
	return nil
}

// Resolve the templates of all actions
// If we don't do this, Ansible reads in Go's template tags
// which results in a jinja error
func (c *Workspace) templateActionScripts() *Workspace {
	// Loop over all blocks
	for _, block := range c.index.Blocks {
		// Generate intermediate snapshot
		snapshot := WorkspaceSnapshot{
			Workspace: c,
			Block:     block,
		}

		for index := range block.Actions {
			action := &block.Actions[index]
			// Loop over script of action
			var newScript = []string{}
			//printObject(snapshot)
			if len(action.Script) > 0 {
				// Loop over script slice and substitute the template
				for _, scriptLine := range action.Script {
					//log.Debug("Templating at snapshot generation, line: " + scriptLine)
					t, err := template.New("script-line").Parse(scriptLine)
					if err != nil {
						c.err = err
						return c
					}

					// Execute the template and save the substituted content to a var
					var substitutedScriptLine bytes.Buffer
					err = t.Execute(&substitutedScriptLine, snapshot)
					if err != nil {
						c.err = err
						return c
					}
					//log.Debug("Substituted the following content: " + substitutedScriptLine.String())

					//scriptLineStr := fmt.Sprintf("%v", scriptLine)
					//scriptStrings = append(scriptStrings, scriptLineStr)
					//log.Debugf("scriptLine as String: %s", scriptLineStr)
					newScript = append(newScript, substitutedScriptLine.String())
				}
			}

			// Update the action script with the substituted one
			action.Script = newScript
		}
	}
	return c
}

func (w *Workspace) DumpEnv() []string {
	envVars := []string{}
	for envVar := range w.env {
		value := w.env[envVar]
		s := strings.Join([]string{envVar, value}, "=")
		envVars = append(envVars, s)
	}
	return envVars
}

func (w *Workspace) DumpMounts() []string {
	mounts := []string{}
	for mount := range w.mounts {
		value := w.mounts[mount]
		s := strings.Join([]string{mount, value}, "=")
		mounts = append(mounts, s)
	}
	return mounts
}

func (w *Workspace) bootstrapEnvVars(ctx context.Context) *Workspace {
	// Prepare env map
	w.env = map[string]string{}

	var _force string
	if force {
		_force = "1"
	} else {
		_force = "0"
	}
	var _in_container string
	if local {
		_in_container = "0"
	} else {
		_in_container = "1"
	}
	w.registerEnvVar("ANSIBLE_DISPLAY_SKIPPED_HOSTS", "False")
	w.registerEnvVar("ANSIBLE_DISPLAY_OK_HOSTS", "True")
	w.registerEnvVar("ANSIBLE_HOST_KEY_CHECKING", "False")
	w.registerEnvVar("ANSIBLE_ACTION_WARNINGS", "False")
	w.registerEnvVar("ANSIBLE_COMMAND_WARNINGS", "False")
	w.registerEnvVar("ANSIBLE_LOCALHOST_WARNING", "False")
	w.registerEnvVar("ANSIBLE_DEPRECATION_WARNINGS", "False")
	w.registerEnvVar("ANSIBLE_ROLES_PATH", "/root/.ansible/roles:/usr/share/ansible/roles:/etc/ansible/roles")
	w.registerEnvVar("ANSIBLE_COLLECTIONS_PATH", "/root/.ansible/collections:/usr/share/ansible/collections:/etc/ansible/collections")
	w.registerEnvVar("ANSIBLE_VERBOSITY", strconv.Itoa(polycrate.Config.Loglevel))
	w.registerEnvVar("ANSIBLE_SSH_PRIVATE_KEY_FILE", filepath.Join(w.ContainerPath, w.Config.SshPrivateKey))
	w.registerEnvVar("ANSIBLE_PRIVATE_KEY_FILE", filepath.Join(w.ContainerPath, w.Config.SshPrivateKey))
	//c.registerEnvVar("ANSIBLE_VARS_ENABLED", "polycrate_vars")
	w.registerEnvVar("ANSIBLE_RUN_VARS_PLUGINS", "start")
	w.registerEnvVar("ANSIBLE_VARS_PLUGINS", "/root/.ansible/plugins/vars:/usr/share/ansible/plugins/vars")
	w.registerEnvVar("DEFAULT_VARS_PLUGIN_PATH", "/root/.ansible/plugins/vars:/usr/share/ansible/plugins/vars")
	//c.registerEnvVar("ANSIBLE_CALLBACKS_ENABLED", "timer,profile_tasks,profile_roles")
	w.registerEnvVar("POLYCRATE_CLI_VERSION", version)
	w.registerEnvVar("POLYCRATE_IMAGE_REFERENCE", workspace.Config.Image.Reference)
	w.registerEnvVar("POLYCRATE_IMAGE_VERSION", workspace.Config.Image.Version)
	w.registerEnvVar("POLYCRATE_FORCE", _force)
	w.registerEnvVar("POLYCRATE_VERSION", version)
	w.registerEnvVar("IN_CI", "true")
	w.registerEnvVar("IN_CONTAINER", _in_container)
	w.registerEnvVar("TERM", "xterm-256color")

	if local {
		// Not in container
		w.registerEnvVar("POLYCRATE_WORKSPACE", workspace.LocalPath)
	} else {
		// In container
		w.registerEnvVar("POLYCRATE_WORKSPACE", workspace.ContainerPath)
	}

	for _, envVar := range w.ExtraEnv {
		// Split by =
		p := strings.Split(envVar, "=")

		if len(p) == 2 {
			w.registerEnvVar(p[0], p[1])
		} else {
			w.err = goErrors.New("Illegal value for env var found: " + envVar)
			return w
		}
	}

	return w
}

func (c *Workspace) GetSnapshot(ctx context.Context) WorkspaceSnapshot {
	log := polycrate.GetContextLogger(ctx)
	log.Debug("Generating snapshot")

	snapshot := WorkspaceSnapshot{
		Workspace: c,
		Action:    c.currentAction,
		Block:     c.currentBlock,
		Workflow:  c.currentWorkflow,
		Step:      c.currentStep,
		Env:       c.env,
		Mounts:    c.mounts,
	}

	// c.RegisterSnapshotEnv(snapshot).Flush()
	return snapshot
}

func (w *Workspace) SaveSnapshot(ctx context.Context) (string, error) {
	log := polycrate.GetContextLogger(ctx)
	snapshot := w.GetSnapshot(ctx)

	//c.RegisterSnapshotEnv(snapshot).Flush()

	// Marshal the snapshot to yaml
	data, err := yaml.Marshal(snapshot)
	if err != nil {
		return "", err
	}

	if data != nil {
		txid := polycrate.GetContextTXID(ctx)
		snapshotSlug := slugify([]string{txid.String(), "workspace", "snapshot"})
		snapshotFilename := strings.Join([]string{snapshotSlug, "yml"}, ".")

		f, err := polycrate.getTempFile(ctx, snapshotFilename)
		if err != nil {
			return "", err
		}

		// Create a viper config object
		snapshotConfig := viper.NewWithOptions(viper.KeyDelimiter("::"))
		snapshotConfig.SetConfigType("yaml")
		snapshotConfig.ReadConfig(bytes.NewBuffer(data))
		err = snapshotConfig.WriteConfigAs(f.Name())
		if err != nil {
			return "", err
		}

		log = log.WithField("path", f.Name())
		log.Debugf("Saved snapshot")

		// Closing file descriptor
		// Getting fatal errors on windows WSL2 when accessing
		// the mounted script file from inside the container if
		// the file descriptor is still open
		// Works flawlessly with open file descriptor on M1 Mac though
		// It's probably safer to close the fd anyways
		f.Close()

		w.registerEnvVar("POLYCRATE_WORKSPACE_SNAPSHOT_YAML", f.Name())
		w.registerMount(f.Name(), f.Name())

		return f.Name(), nil
	} else {
		return "", fmt.Errorf("cannot save snapshot")
	}

}

func (w *Workspace) registerEnvVar(key string, value string) {
	w.env[key] = value
}

func (c *Workspace) registerMount(host string, container string) {
	log.Tracef("Registering container mount: %s -> %s", host, container)
	c.mounts[host] = container
}

func (c *Workspace) registerCurrentBlock(block *Block) {

	c.registerEnvVar("POLYCRATE_BLOCK", block.Name)

	if block.Workdir.exists {
		c.registerEnvVar("POLYCRATE_BLOCK_WORKDIR", block.Workdir.ContainerPath)
	}
	c.currentBlock = block
}
func (c *Workspace) registerCurrentAction(action *Action) {

	c.registerEnvVar("POLYCRATE_ACTION", action.Name)
	c.currentAction = action
}
func (c *Workspace) registerCurrentWorkflow(workflow *Workflow) {

	c.registerEnvVar("POLYCRATE_WORKFLOW", workflow.Name)
	c.currentWorkflow = workflow
}
func (c *Workspace) registerCurrentStep(step *Step) {

	c.registerEnvVar("POLYCRATE_STEP", step.Name)
	c.currentStep = step
}

func (c *Workspace) GetActionFromIndex(name string) *Action {
	if action, ok := c.index.Actions[name]; ok {
		return action
	}
	return nil
}

// func (c *Workspace) getActionByPath(actionPath string) *Action {
// 	// Validate actionPath
// 	s := strings.Split(actionPath, ".")
// 	blockName := s[0]
// 	actionName := s[1]

// 	block := workspace.getBlockByName(blockName)

// 	if block != nil {
// 		action := block.getActionByName(actionName)

// 		if action != nil {
// 			return action
// 		}
// 	}
// 	return nil
// }

func (w *Workspace) GetBlockWithContext(ctx context.Context, name string) (context.Context, *Block, error) {
	block, err := w.GetBlock(name)
	if err != nil {
		return ctx, nil, err
	}

	blockKey := ContextKey("block")
	ctx = context.WithValue(ctx, blockKey, block)

	log := polycrate.GetContextLogger(ctx)
	log = log.WithField("block", block.Name)
	ctx = polycrate.SetContextLogger(ctx, log)

	return ctx, block, nil
}

func (w *Workspace) GetWorkflowWithContext(ctx context.Context, name string) (context.Context, *Workflow, error) {
	workflow, err := w.GetWorkflow(name)
	if err != nil {
		return ctx, nil, err
	}

	workflowKey := ContextKey("workflow")
	ctx = context.WithValue(ctx, workflowKey, workflow)

	log := polycrate.GetContextLogger(ctx)
	log = log.WithField("workflow", workflow.Name)
	ctx = polycrate.SetContextLogger(ctx, log)

	return ctx, workflow, nil
}

func (c *Workspace) GetBlockFromIndex(name string) *Block {
	if block, ok := c.index.Blocks[name]; ok {
		return block
	}
	return nil
}

func (c *Workspace) getBlockByName(blockName string) *Block {
	for i := 0; i < len(c.Blocks); i++ {
		block := c.Blocks[i]
		if block.Name == blockName {
			return block
		}
	}
	return nil
}
func (c *Workspace) GetBlock(name string) (*Block, error) {
	for i := 0; i < len(c.Blocks); i++ {
		block := c.Blocks[i]
		if block.Name == name {
			return block, nil
		}
	}
	return nil, fmt.Errorf("block not found: %s", name)
}
func (c *Workspace) GetWorkflow(name string) (*Workflow, error) {
	for i := 0; i < len(c.Workflows); i++ {
		workflow := c.Workflows[i]
		if workflow.Name == name {
			return workflow, nil
		}
	}
	return nil, fmt.Errorf("workflow not found: %s", name)
}

func (c *Workspace) GetWorkflowFromIndex(name string) *Workflow {
	if workflow, ok := c.index.Workflows[name]; ok {
		return workflow
	}
	return nil
}

// func (c *Workspace) Create(ctx context.Context, path string) error {
// 	workspacePath := filepath.Join(path, c.Name)

// 	// Check if a workspace with this name already exists
// 	if localWorkspaceIndex[c.Name] != "" {
// 		// We found a workspace with that name in the index

// 		c.err = fmt.Errorf("workspace already exists: %s", localWorkspaceIndex[c.Name])
// 		return c
// 	}

// 	// Check if the directory for this workspace already exists in polycrateWorkspaceDir
// 	if _, err := os.Stat(workspacePath); !os.IsNotExist(err) {
// 		// Directory already exists
// 		c.err = fmt.Errorf("workspace directory already exists: %s", workspacePath)
// 		return c
// 	} else {
// 		err := CreateDir(workspacePath)
// 		if err != nil {
// 			c.err = err
// 			return c
// 		}

// 		// Set the new directory as workspace.Path
// 		c.LocalPath = workspacePath
// 	}
// 	// Save a new workspace config to workspace.poly in the given directory
// 	var n Workspace

// 	// Update workspace vars with minimum defaults
// 	n.Name = c.Name

// 	if c.SyncOptions.Enabled {
// 		n.SyncOptions.Enabled = true
// 	}

// 	if c.SyncOptions.Auto {
// 		n.SyncOptions.Auto = true
// 	}

// 	if c.SyncOptions.Remote.Url != "" {
// 		n.SyncOptions.Remote.Url = c.SyncOptions.Remote.Url
// 		n.SyncOptions.Remote.Branch.Name = c.SyncOptions.Remote.Branch.Name
// 	}

// 	n.saveWorkspace(c.LocalPath).Flush()

// 	return c
// }

func (w *Workspace) UpdateSyncStatus(ctx context.Context) error {
	log := polycrate.GetContextLogger(ctx)

	if polycrate.Config.Sync.Enabled {
		if GitHasRemote(ctx, w.LocalPath) {
			log.Tracef("Getting remote repository status")

			// https://stackoverflow.com/posts/68187853/revisions
			// Get remote reference
			_, err := GitFetch(ctx, w.LocalPath)
			if err != nil {
				return err
			}

			// LocalRefReference := this.SyncOptions.Local.Branch.Name
			// RemoteRefReference := fmt.Sprintf("%s/%s", this.SyncOptions.Remote.Name, this.SyncOptions.Remote.Branch.Name)

			//log.Tracef("Getting last local commit")
			// LocalRefBranchName := this.SyncOptions.Local.Branch.Name
			// LocalRefCommit, err := GitGetHeadCommit(this.LocalPath, LocalRefReference)
			// if err != nil {
			// 	this.err = err
			// 	return this
			// }

			//log.Tracef("Getting last remote commit")
			// RemoteRefBranchName := this.SyncOptions.Remote.Branch.Name
			// RemoteRefCommit, err := GitGetHeadCommit(this.LocalPath, RemoteRefReference)
			// if err != nil {
			// 	this.err = err
			// 	return this
			// }

			log.Tracef("Checking if behind remote")
			behindBy, err := GitBehindBy(ctx, w.LocalPath)
			if err != nil {
				return err
			}

			log.Tracef("Checking if ahead of remote")
			aheadBy, err := GitAheadBy(ctx, w.LocalPath)
			if err != nil {
				return err
			}

			// ahead > 0, behind == 0
			if aheadBy != 0 && behindBy == 0 {
				w.syncStatus = "ahead"
			}

			// ahead == 0, behind > 0
			if behindBy != 0 && aheadBy == 0 {
				w.syncStatus = "behind"
			}

			// ahead == 0, behind == 0
			if behindBy == 0 && aheadBy == 0 {
				w.syncStatus = "synced"
			}

			// ahead > 0, behind > 0
			if behindBy != 0 && aheadBy != 0 {
				w.syncStatus = "diverged"
			}

			log.Tracef("Checking for uncommited changes")

			// Has uncommited changes?
			if GitHasChanges(ctx, w.LocalPath) {
				w.syncStatus = "changed"
			}

			log = log.WithField("status", w.syncStatus).WithField("ahead", aheadBy).WithField("behind", behindBy)
			log.Debugf("Reporting sync status")
		} else {
			err := fmt.Errorf("no remote configured")
			return err
		}

	} else {
		log = log.WithField("status", "disabled")
		log.Debugf("Reporting sync status")
	}
	return nil
}

// func (this *Workspace) SyncCommit(message string) *Workspace {
// 	if this.syncLoaded {
// 		hash, err := GitCommitAll(this.LocalPath, message)
// 		if err != nil {
// 			this.err = err
// 			return this
// 		}

// 		log.WithFields(log.Fields{
// 			"workspace": this.Name,
// 			"message":   message,
// 			"hash":      hash,
// 			"module":    "sync",
// 		}).Debugf("Added commit")
// 	} else {
// 		log.WithFields(log.Fields{
// 			"workspace": this.Name,
// 			"message":   message,
// 			"module":    "sync",
// 		}).Debugf("Not committing. Sync module not loaded")
// 	}
// 	return this
// }
func (w *Workspace) Commit(ctx context.Context, message string) error {
	log := polycrate.GetContextLogger(ctx)

	hash, err := GitCommitAll(ctx, w.LocalPath, message)
	if err != nil {
		return err
	}

	log = log.WithField("message", message)
	log = log.WithField("hash", hash)
	log.Tracef("Added commit")

	return nil
}

// func (this *Workspace) SyncPull() *Workspace {
// 	_, err := GitPull(this.LocalPath, this.SyncOptions.Remote.Name, this.SyncOptions.Remote.Branch.Name)
// 	if err != nil {
// 		this.err = err
// 		return this
// 	}
// 	return this
// }

func (w *Workspace) Pull(ctx context.Context) error {

	_, err := GitPull(ctx, w.LocalPath, w.SyncOptions.Remote.Name, w.SyncOptions.Remote.Branch.Name)
	if err != nil {
		return err
	}

	return nil
}

// func (this *Workspace) SyncPush() *Workspace {
// 	_, err := GitPush(this.LocalPath, this.SyncOptions.Remote.Name, this.SyncOptions.Remote.Branch.Name)
// 	if err != nil {
// 		this.err = err
// 		return this
// 	}
// 	return this
// }

func (w *Workspace) Push(ctx context.Context) error {

	_, err := GitPush(ctx, w.LocalPath, w.SyncOptions.Remote.Name, w.SyncOptions.Remote.Branch.Name)
	if err != nil {
		return err
	}

	return nil
}

func (w *Workspace) RunSync(ctx context.Context) error {
	log := polycrate.GetContextLogger(ctx)

	// if !w.syncLoaded {
	// 	err := fmt.Errorf("sync not loaded")
	// 	return err
	// }

	if w.SyncOptions.Enabled {
		if err := w.UpdateSyncStatus(ctx); err != nil {
			return err
		}

		switch status := w.syncStatus; status {
		case "changed":
			log.Debugf("Sync status: changes found - committing")

			if err := w.Commit(ctx, "Sync auto-commit"); err != nil {
				return err
			}

			if err := w.RunSync(ctx); err != nil {
				return err
			}
		case "synced":
			log.Debugf("Sync status: up-to-date")
		case "diverged":
			// log.WithFields(log.Fields{
			// 	"workspace": workspace.Name,
			// 	"module":    "sync",
			// }).Fatalf("Sync error - run `polycrate sync status` for more information")
			err := fmt.Errorf("sync error - run `polycrate sync status` for more information")
			return err
		case "ahead":
			log.Debugf("Sync status: ahead of remote - pushing")

			if err := w.Push(ctx); err != nil {
				return err
			}

			if err := w.RunSync(ctx); err != nil {
				return err
			}
		case "behind":
			log.Debugf("Sync status: behind remote - pulling")

			if err := w.Pull(ctx); err != nil {
				return err
			}

			if err := w.RunSync(ctx); err != nil {
				return err
			}
		}
	} else {
		err := fmt.Errorf("sync disabled")
		return err
	}

	return nil
}

func (w *Workspace) Sync(ctx context.Context) error {
	if w.isGitRepo {
		if !w.synced {
			if w.syncNeeded {
				if polycrate.Config.Sync.Enabled {
					// if !w.syncLoaded {
					// 	if err := w.LoadSync(ctx); err != nil {
					// 		return err
					// 	}
					// }

					// sync workspace
					if err := w.RunSync(ctx); err != nil {
						return err
					}

					// mark workspace as synced
					w.synced = true

					// unmark needed sync
					w.syncNeeded = false

				} else {
					return fmt.Errorf("sync disabled")
				}
			}
		}
	}

	return nil
}

func (w *Workspace) IsBlockInstalled(fullTag string, registryUrl string, blockName string, blockVersion string) bool {
	// Options for fullTag:
	// - cargo.ayedo.cloud/block/name:0.0.1
	// - cargo.ayedo.cloud/block/name --> latest
	// - block/name:0.0.1
	// - block/name --> latest
	var fullBlockName string
	var installedBlockFullName string

	if registryUrl != "" {
		fullBlockName = strings.Join([]string{registryUrl, blockName}, "/")
	} else {
		// Prepend default registry URL
		fullBlockName = strings.Join([]string{polycrate.Config.Registry.Url, blockName}, "/")
	}

	for _, installedBlock := range w.installedBlocks {
		_, installedBlockRegistryUrl, installedBlockName, _ := mapDockerTag(installedBlock.Name)

		if installedBlockRegistryUrl != "" {
			installedBlockFullName = strings.Join([]string{installedBlockRegistryUrl, installedBlockName}, "/")
		} else {
			// Prepend default registry URL
			installedBlockFullName = strings.Join([]string{polycrate.Config.Registry.Url, installedBlockName}, "/")
		}

		if installedBlockFullName == fullBlockName {
			if installedBlock.Version == blockVersion {
				return true
			}
		}
	}

	return false
}

func (w *Workspace) UpdateBlocks(ctx context.Context, args []string) error {
	log := polycrate.GetContextLogger(ctx)

	log.Infof("%d blocks to update", len(args))

	eg := new(errgroup.Group)
	for _, arg := range args {
		arg := arg // https://go.dev/doc/faq#closures_and_goroutines
		eg.Go(func() error {
			fullTag, registryUrl, blockName, blockVersion := mapDockerTag(arg)

			log = log.WithField("version", blockVersion)
			log = log.WithField("block", blockName)
			ctx = polycrate.SetContextLogger(ctx, log)

			//log.Debugf("Updating dependency")

			// Download blocks from registry
			err := w.PullBlock(ctx, fullTag, registryUrl, blockName, blockVersion)
			if err != nil {
				return err
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}

// - Accepts 1 arg: the block name/slug as it is in the registry
// - Accepts 1 flag: version; defaults to latest
// - Checks if a block with that name exists already AND has a block dir
// - If the block exists, the command fails with a warning and shows a hint to the update command
// - If no block exists, looks up the name of the block via Wordpress API at polycrate.io
// - If a block is found, gets the list of releases
// - Marks the youngest release as "latest"
// - Downloads the release bundle
// - If download succeeds, creates a block dir for the block
// - unpacks the release bundle to the block dir
// func (c *Workspace) InstallBlocks(args []string) error {
// 	for _, arg := range args {
// 		blockName, blockVersion, err := registry.resolveArg(arg)
// 		if err != nil {
// 			return err
// 		}

func (w *Workspace) PullBlock(ctx context.Context, fullTag string, registryUrl string, blockName string, blockVersion string) error {
	log := polycrate.GetContextLogger(ctx)

	log = log.WithField("block", blockName)
	log = log.WithField("version", blockVersion)
	log = log.WithField("registry", registryUrl)
	ctx = polycrate.SetContextLogger(ctx, log)

	if w.IsBlockInstalled(fullTag, registryUrl, blockName, blockVersion) {
		log.Infof("Block is already installed")
		return nil
	}

	targetDir := filepath.Join(w.LocalPath, w.Config.BlocksRoot, registryUrl, blockName)

	//log.Debugf("Pulling block %s:%s", blockName, blockVersion)
	log = log.WithField("path", targetDir)
	ctx = polycrate.SetContextLogger(ctx, log)

	log.Infof("Pulling block")

	err := UnwrapOCIImage(ctx, targetDir, registryUrl, blockName, blockVersion)
	if err != nil {
		return err
	}

	// Load Block
	//var block Block
	//block.Workdir.LocalPath = targetDir
	block, err := w.LoadBlock(ctx, targetDir)
	if err != nil {
		return err
	}

	// Register installed block
	w.installedBlocks = append(w.installedBlocks, block)

	return nil
}

func (c *Workspace) PushBlock(ctx context.Context, blockName string) error {
	log := polycrate.GetContextLogger(ctx)
	// Get the block
	// Check that it has a version
	// Check if release with new version exists in registry
	// Get the latest version from the registry
	// Check if new version > current version
	//

	block := c.GetBlockFromIndex(blockName)
	if block != nil {
		// The block exists
		// Check it has a version
		if block.Version == "" {
			return fmt.Errorf("block has no version")
		}
	} else {
		return fmt.Errorf("block not found in workspace: %s", blockName)
	}
	log = log.WithField("block", block.Name)
	log = log.WithField("version", block.Version)
	log = log.WithField("path", block.Workdir.LocalPath)

	log.Infof("Pushing block")

	_, registryUrl, blockName, _ := mapDockerTag(block.Name)
	tagVersion := block.Version

	// if --dev flag has been used, assume we're just developing that block
	// append "dev" and the first 98 chars of the transaction ID to the tag (still sem-ver compatible)
	if dev {
		// Append "-dev" to tag
		txid := polycrate.GetContextTXID(ctx)
		_tagVersion := strings.Join([]string{tagVersion, "dev"}, "-")
		tagVersion = strings.Join([]string{_tagVersion, txid.String()[:8]}, ".")

		if block.Labels == nil {
			block.Labels = map[string]string{}
		}
		block.Labels["polycrate.flags.dev"] = "true"
	}
	err := WrapOCIImage(ctx, block.Workdir.LocalPath, registryUrl, blockName, tagVersion, block.Labels)
	if err != nil {
		return err
	}
	return nil
}

func (c *Workspace) UninstallBlocks(ctx context.Context, args []string) error {
	for _, blockName := range args {
		block := c.GetBlockFromIndex(blockName)
		if block != nil {
			err := block.Uninstall(ctx, pruneBlock)
			if err != nil {
				return err
			}
		} else {
			err := fmt.Errorf("Block not found: %s", blockName)
			return err
		}
	}
	return nil
}

// func (c *Workspace) getWorkflowByName(workflowName string) *Workflow {

// 	for i := 0; i < len(c.Workflows); i++ {
// 		workflow := &c.Workflows[i]
// 		if workflow.Name == workflowName {
// 			return workflow
// 		}
// 	}
// 	return nil
// }

func (c *Workspace) Print() {
	//fmt.Printf("%#v\n", c)
	printObject(c)
}

func (w *Workspace) ImportInstalledBlocks(ctx context.Context) *Workspace {
	log := polycrate.GetContextLogger(ctx)
	log.Debugf("Importing installed blocks")

	for _, installedBlock := range w.installedBlocks {
		log = log.WithField("installed_block", installedBlock.Name)
		ctx = polycrate.SetContextLogger(ctx, log)

		// Check if Block exists
		existingBlock := w.getBlockByName(installedBlock.Name)

		if existingBlock != nil {
			log = log.WithField("existing_block", installedBlock.Name)
			ctx = polycrate.SetContextLogger(ctx, log)

			// Block exists
			log.Tracef("Found existing block. Merging.")

			err := existingBlock.MergeIn(installedBlock)
			if err != nil {
				w.err = err
				return w
			}
		} else {
			w.Blocks = append(w.Blocks, installedBlock)
		}
	}
	return w

}
func (w *Workspace) LoadInstalledBlocks(ctx context.Context) error {
	log := polycrate.GetContextLogger(ctx)
	for _, installedBlock := range w.installedBlocks {
		log = log.WithField("block", installedBlock.Name)
		ctx = polycrate.SetContextLogger(ctx, log)

		// Check if Block exists
		existingBlock, err := w.GetBlock(installedBlock.Name)

		if err != nil {
			// Block is not in the index yet
			w.Blocks = append(w.Blocks, installedBlock)
		} else {

			// Block exists
			log.Tracef("Found existing block with the same name. Merging.")

			// Make sure "resolved" is false
			// If not, the block will stay "resolved" during a workflow
			// and base blocks will not be merged in again
			log.Tracef("Marking block as unresolved")
			existingBlock.resolved = false

			err := existingBlock.MergeIn(installedBlock)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *Workspace) LoadBlock(ctx context.Context, path string) (*Block, error) {
	block := new(Block)
	blockConfigFilePath := filepath.Join(path, w.Config.BlocksConfig)
	block.Workdir.LocalPath = path

	//blockConfigObject := viper.New()
	// https: //github.com/spf13/viper#unmarshaling
	// allows to use dots in keys
	blockConfigObject := viper.NewWithOptions(viper.KeyDelimiter("::"))
	blockConfigObject.SetConfigType("yaml")
	blockConfigObject.SetConfigFile(blockConfigFilePath)

	// log.WithFields(log.Fields{
	// 	"workspace": w.Name,
	// 	"path":      b.Workdir.LocalPath,
	// }).Debugf("Loading installed block")

	if err := blockConfigObject.MergeInConfig(); err != nil {
		return nil, err
	}

	if err := blockConfigObject.UnmarshalExact(&block); err != nil {
		return nil, err
	}

	// Load schema
	schemaFilePath := filepath.Join(block.Workdir.LocalPath, "schema.json")
	if _, err := os.Stat(schemaFilePath); !os.IsNotExist(err) {
		block.schema = schemaFilePath
	}

	if err := block.validate(); err != nil {
		return nil, err
	}

	// Set Block vars
	relativeBlockPath, err := filepath.Rel(filepath.Join(w.LocalPath, w.Config.BlocksRoot), block.Workdir.LocalPath)
	if err != nil {
		return nil, err
	}
	block.Workdir.ContainerPath = filepath.Join(w.ContainerPath, w.Config.BlocksRoot, relativeBlockPath)

	if local {
		block.Workdir.Path = block.Workdir.LocalPath
	} else {
		block.Workdir.Path = block.Workdir.ContainerPath
	}

	return block, nil
}

// func (c *Workspace) LoadInstalledBlocks() *Workspace {
// 	log.WithFields(log.Fields{
// 		"workspace": c.Name,
// 	}).Debugf("Loading installed blocks")

// 	for _, blockPath := range blockPaths {
// 		var loadedBlock Block
// 		loadedBlock.Workdir.LocalPath = filepath.Join(blockPath, c.Config.BlocksConfig)
// 		err := c.LoadBlock(&loadedBlock)
// 		if err != nil {
// 			c.err = err
// 			return c
// 		}

// 		// blockConfigFilePath := filepath.Join(blockPath, c.Config.BlocksConfig)

// 		// blockConfigObject := viper.New()
// 		// blockConfigObject.SetConfigType("yaml")
// 		// blockConfigObject.SetConfigFile(blockConfigFilePath)

// 		// log.WithFields(log.Fields{
// 		// 	"workspace": c.Name,
// 		// 	"path":      blockPath,
// 		// }).Debugf("Loading installed block")
// 		// if err := blockConfigObject.MergeInConfig(); err != nil {
// 		// 	c.err = err
// 		// 	return c
// 		// }

// 		// if err := blockConfigObject.UnmarshalExact(&loadedBlock); err != nil {
// 		// 	c.err = err
// 		// 	return c
// 		// }
// 		// if err := loadedBlock.validate(); err != nil {
// 		// 	c.err = err
// 		// 	return c
// 		// }
// 		// log.WithFields(log.Fields{
// 		// 	"block":     loadedBlock.Name,
// 		// 	"workspace": c.Name,
// 		// }).Debugf("Loaded block")

// 		// // Set Block vars
// 		// loadedBlock.Workdir.LocalPath = blockPath
// 		// loadedBlock.Workdir.ContainerPath = filepath.Join(c.ContainerPath, c.Config.BlocksRoot, loadedBlock.Name)

// 		// if local {
// 		// 	loadedBlock.Workdir.Path = loadedBlock.Workdir.LocalPath
// 		// } else {
// 		// 	loadedBlock.Workdir.Path = loadedBlock.Workdir.ContainerPath
// 		// }

// 		// // Add block to installedBlocks
// 		// c.installedBlocks = append(installedBlocks, loadedBlock)
// 	}
// 	return c

// }

func (w *Workspace) DiscoverInstalledBlocks() *Workspace {
	blocksDir := filepath.Join(w.LocalPath, w.Config.BlocksRoot)

	if _, err := os.Stat(blocksDir); !os.IsNotExist(err) {
		log.WithFields(log.Fields{
			"workspace": w.Name,
			"path":      blocksDir,
		}).Debugf("Discovering blocks")

		// This function adds all valid Blocks to the list of
		err := filepath.WalkDir(blocksDir, w.WalkBlocksDir)
		if err != nil {
			w.err = err
		}
	} else {
		log.WithFields(log.Fields{
			"workspace": w.Name,
			"path":      blocksDir,
		}).Debugf("Skipping block discovery. Blocks directory not found")
	}

	return w
}
func (w *Workspace) FindInstalledBlocks(ctx context.Context, path string) error {
	blocksDir := filepath.Join(w.LocalPath, w.Config.BlocksRoot)

	// reset installedBlocks
	// when running a workflow, without reseting the count of installed blocks just continuously increases
	w.installedBlocks = []*Block{}

	if _, err := os.Stat(blocksDir); !os.IsNotExist(err) {
		// This function adds all valid Blocks to the list of
		err := filepath.WalkDir(blocksDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !d.IsDir() {
				fileinfo, _ := d.Info()

				if fileinfo.Name() == w.Config.BlocksConfig {
					blockConfigFileDir := filepath.Dir(path)

					//var block Block
					//block.Workdir.LocalPath = blockConfigFileDir

					block, err := w.LoadBlock(ctx, blockConfigFileDir)
					if err != nil {
						return err
					}

					// Add block to installedBlocks
					w.installedBlocks = append(w.installedBlocks, block)
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	} else {
		log := polycrate.GetContextLogger(ctx)
		log.WithField("path", blocksDir).Debugf("Skipping block discovery. Blocks directory not found")
	}
	log.Debugf("Installed blocks: %d", len(w.installedBlocks))

	return nil
}

func (w *Workspace) WalkBlocksDir(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}
	log.Warn("WalkBlocksDir is deprecated")

	if !d.IsDir() {
		fileinfo, _ := d.Info()

		if fileinfo.Name() == w.Config.BlocksConfig {
			blockConfigFileDir := filepath.Dir(path)

			// var block Block
			// block.Workdir.LocalPath = blockConfigFileDir

			ctx := context.Background()
			block, err := w.LoadBlock(ctx, blockConfigFileDir)
			if err != nil {
				return err
			}

			// Add block to installedBlocks
			w.installedBlocks = append(w.installedBlocks, block)
		}
	}
	return nil

}

func (c *Workspace) PullDependencies() *Workspace {
	// Loop through dependencies
	// Resolve each dep into block/version
	// Check if block already exists
	// If yes, check if version matches
	// If version matches, continue
	// If version doesn't match, update
	// If no, install

	log.WithFields(log.Fields{
		"workspace": c.Name,
	}).Debugf("Resolving workspace dependencies")

	// err := c.UpdateBlocks(c.Dependencies)
	// if err != nil {
	// 	c.err = err
	// 	return c
	// }
	return c
}

func (c *Workspace) getTempFile(ctx context.Context, filename string) (*os.File, error) {
	fp := filepath.Join(workspace.runtimeDir, filename)
	f, err := os.Create(fp)
	if err != nil {
		return nil, err
	}
	return f, nil
}
