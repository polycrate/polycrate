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
	"bufio"
	goErrors "errors"
	"io/ioutil"
	"os"

	"github.com/InVisionApp/conjungo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var actionsCmd = &cobra.Command{
	Use:   "actions",
	Short: "Control Polycrate actions",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		loadWorkspace()
		// List pipelines
	},
}

func init() {
	rootCmd.AddCommand(actionsCmd)
}

type ActionAnsibleConfig struct {
	Inventory string `mapstructure:"inventory,omitempty" json:"inventory,omitempty"`
	Hosts     string `mapstructure:"hosts,omitempty" json:"hosts,omitempty"`
}

type ActionKubernetesConfig struct {
	Kubeconfig string `mapstructure:"kubeconfig,omitempty" json:"kubeconfig,omitempty"`
	Context    string `mapstructure:"context,omitempty" json:"context,omitempty"`
}
type Action struct {
	Metadata            Metadata               `mapstructure:"metadata" json:"metadata" validate:"required"`
	Interactive         bool                   `mapstructure:"interactive,omitempty" json:"interactive,omitempty"`
	Script              []string               `mapstructure:"script,omitempty" json:"script,omitempty" validate:"required_if=Action"`
	Ansible             ActionAnsibleConfig    `mapstructure:"ansible,omitempty" json:"ansible,omitempty"`
	Kubernetes          ActionKubernetesConfig `mapstructure:"kubernetes,omitempty" json:"kubernetes,omitempty"`
	executionScriptPath string
	address             string
	block               *Block
}

type ActionRun struct {
	Env          []string
	Mounts       []string
	ContainerCmd []string
	Local        bool
	ID           string
	Action       *Action
	//Container    ContainerConfig
	err error
}

func (c *Action) MergeIn(action *Action) error {
	opts := conjungo.NewOptions()
	opts.Overwrite = false // do not overwrite existing values in workspaceConfig
	if err := conjungo.Merge(c, action, opts); err != nil {
		return err
	}
	return nil
}

func (c *Action) Run() error {
	log.Infof("Running Action '%s' of Block '%s'", c.Metadata.Name, workspace.currentBlock.Metadata.Name)

	var ar ActionRun = ActionRun{}
	// What to do if you want to run an action
	// 1. Detect the correct workdir
	// 2. Determine if the Run is local or in the container
	// 3. Determine if you need to mount an inventory
	// 4. Determine if you need to mount a Kubeconfig
	// 5. Set Env Vars
	//   ANSIBLE_INVENTORY=/etc/ansible/hosts
	//   POLYCRATE_WORKSPACE_ROOT=/workspace
	//   KUBECONFIG=...
	//   KUBECONFIG=...
	// 6. Prepare Mounts + Env
	// 7.

	// 0. Set Action
	ar.Action = c

	// 1. Determine the correct workdir
	if workspace.currentBlock.workdir.localPath == "" {
		workdir = workspaceDir
		workdirContainer = workspaceContainerDir
	} else {
		workdir = workspace.currentBlock.workdir.localPath
		workdirContainer = workspace.currentBlock.workdir.containerPath
	}
	log.Debugf("Workdir Local: %s", workdir)
	log.Debugf("Workdir Container: %s", workdirContainer)

	// 2. Determine if the Run is local or in the container
	ar.Local = local

	// 3. Bootstrap envVars
	workspace.bootstrapEnvVars()

	// 3. Determine inventory path
	workspace.registerEnvVar("ANSIBLE_INVENTORY", workspace.currentBlock.getInventoryPath())

	// 4. Determine kubeconfig path
	workspace.registerEnvVar("KUBECONFIG", workspace.currentBlock.getInventoryPath())

	// Bootstrap mounts
	workspace.bootstrapMounts()

	// Save execution script
	err := c.saveExecutionScript()
	if err != nil {
		return err
	}

	// register environment variables
	workspace.registerEnvVar("POLYCRATE_RUNTIME_SCRIPT_PATH", c.executionScriptPath)

	// register mounts
	workspace.registerMount(c.executionScriptPath, c.executionScriptPath)

	// Wrapup
	ar.Env = envVars
	ar.Mounts = mounts
	ar.ContainerCmd = []string{"bash", "-c", c.executionScriptPath}

	if action.Interactive {
		// Set interactive=true globally
		interactive = true
	}

	// 5. Assemble container config
	// cc := &dockerclient.ContainerConfig{
	// 	// ...
	// 	Volumes: map[string]struct{}{
	// 		"foo": struct{}{},
	// 		"bar": struct{}{},
	// 		// ...
	// 	},
	// }

	doContainerStuff()

	return nil
}

func (c *Action) saveExecutionScript() error {
	// Prepare script
	scriptSlice := []string{
		"#!/bin/bash",
		"set -euo pipefail",
	}
	script := append(scriptSlice, c.Script...)

	f, err := ioutil.TempFile("/tmp", "polycrate."+workspace.Metadata.Name+".script.*.sh")
	if err != nil {
		return err
	}
	datawriter := bufio.NewWriter(f)

	for _, data := range script {
		_, _ = datawriter.WriteString(data + "\n")
	}

	datawriter.Flush()
	log.Debug("Saved temporary execution script to " + f.Name())

	// Make executable
	err = os.Chmod(f.Name(), 0755)
	if err != nil {
		return err
	}

	// Closing file descriptor
	// Getting fatal errors on windows WSL2 when accessing
	// the mounted script file from inside the container if
	// the file descriptor is still open
	// Works flawlessly with open file descriptor on M1 Mac though
	// It's probably safer to close the fd anyways
	f.Close()

	// Set executionScriptPath
	c.executionScriptPath = f.Name()
	return nil
}

func (c *Action) GetExecutionScript() []string {
	// Prepare script
	scriptSlice := []string{
		"#!/bin/bash",
		"set -euo pipefail",
	}

	script := append(scriptSlice, c.Script...)
	return script
}

func (c *Action) Validate() error {
	if c.Script == nil {
		return goErrors.New("No script found for Action")
	}
	return nil
}

func (c *Action) ValidateScript() error {
	if c.Script == nil {
		return goErrors.New("No script found for command")
	}
	return nil
}

// func (c *Action) Trigger() (int, error) {
// 	err, runtimeScriptPath := c.SaveExecutionScript()
// 	if err != nil {
// 		return 1, err
// 	}

// 	runCommand := []string{"bash", "-c", runtimeScriptPath}

// 	// register environment variables
// 	registerEnvVar("POLYCRATE_RUNTIME_SCRIPT_PATH", runtimeScriptPath)

// 	// register mounts
// 	registerMount(runtimeScriptPath, runtimeScriptPath)

// 	if action.Interactive {
// 		// Set interactive=true globally
// 		interactive = true
// 	}

// 	// Execute the container
// 	var exitCode int
// 	if local {
// 		exitCode, err = RunContainer(
// 			workspace.Config.Image.Reference,
// 			workspace.Config.Image.Version,
// 			runCommand,
// 		)
// 	} else {
// 		exitCode, err = RunCommand("docker", runCommand...)
// 	}
// 	if err != nil {
// 		log.Error("Plugin ", "asd", " failed with exit code ", exitCode, ": ", err.Error())
// 	} else {
// 		log.Info("Plugin ", "asd", " succeeded with exit code ", exitCode, ": OK")
// 	}

// 	return 0, nil
// }

func (c *Action) SaveExecutionScript() (error, string) {
	f, err := ioutil.TempFile("/tmp", "cloudstack."+workspace.Metadata.Name+".run.*.sh")
	if err != nil {
		return err, ""
	}
	datawriter := bufio.NewWriter(f)

	for _, data := range c.GetExecutionScript() {
		_, _ = datawriter.WriteString(data + "\n")
	}
	datawriter.Flush()
	log.Debug("Saved script to " + f.Name())

	err = os.Chmod(f.Name(), 0755)
	if err != nil {
		return err, ""
	}

	// Closing file descriptor
	// Getting fatal errors on windows WSL2 when accessing
	// the mounted script file from inside the container if
	// the file descriptor is still open
	// Works flawlessly with open file descriptor on M1 Mac though
	// It's probably safer to close the fd anyways
	f.Close()
	return nil, f.Name()
}

func (c *Action) Inspect() {
	printObject(c)
}
