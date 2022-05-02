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
	"errors"
	goErrors "errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/InVisionApp/conjungo"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var actionsCmd = &cobra.Command{
	Use: "action",
	Aliases: []string{
		"actions",
	},
	Short: "Control Polycrate actions",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		workspace.load()
		if workspace.Flush() != nil {
			log.Fatal(workspace.Flush)
		}

		workspace.listActions()
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
	//Metadata            Metadata               `mapstructure:"metadata,squash" json:"metadata" validate:"required"`
	Name                string                 `mapstructure:"name" json:"name" validate:"required"`
	Description         string                 `mapstructure:"description" json:"description"`
	Labels              map[string]string      `mapstructure:"labels" json:"labels"`
	Alias               []string               `mapstructure:"alias" json:"alias"`
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

func (c *Action) RunContainer() error {
	containerImage := strings.Join([]string{workspace.Config.Image.Reference, workspace.Config.Image.Version}, ":")
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	if err != nil {
		panic(err)
	}

	// Check if a Dockerfile is configured in the Workspace
	if workspace.Config.Dockerfile != "" {
		if build {

			// Create the filepath
			dockerfilePath := filepath.Join(workspace.path, workspace.Config.Dockerfile)

			// Check if the file exists
			if _, err := os.Stat(dockerfilePath); !os.IsNotExist(err) {
				// We need to build and tag this
				log.Debugf("Found %s in Workspace", workspace.Config.Dockerfile)

				tag := workspace.Name + ":" + version
				log.Debugf("Building image '%s', --build=%t", tag, build)

				tags := []string{tag}
				containerImage, err = buildContainerImage(workspace.Config.Dockerfile, tags)
				if err != nil {
					return err
				}
			} else {
				return errors.New("Could not find Dockerfile at " + workspace.Config.Dockerfile)
			}
		} else {
			if pull {
				log.Debugf("Pulling image %s: --pull=%t, --build=%t", containerImage, pull, build)
				err := workspace.pullContainerImage(containerImage)

				if err != nil {
					log.Fatal(err)
				}
			} else {
				log.Debugf("Not pulling/building image %s: --pull=%t, --build=%t", containerImage, pull, build)
			}
		}
	} else {
		if pull {
			log.Debugf("Pulling image %s: --pull=%t, --build=%t", containerImage, pull, build)
			err := workspace.pullContainerImage(containerImage)

			if err != nil {
				log.Fatal(err)
			}
		} else {
			log.Debugf("Not pulling/building image %s: --pull=%t, --build=%t", containerImage, pull, build)
		}
	}

	entrypoint := []string{"bash", "-c"}
	runCommand := []string{}

	if c.executionScriptPath != "" {
		log.Debugf("Running Script from Action at %s", c.executionScriptPath)
		runCommand = append(runCommand, c.executionScriptPath)
	} else {
		return goErrors.New("no execution script path given. Nothing to do")
	}
	log.Debugf("Running entrypoint %s", entrypoint)
	log.Debugf("Running command %s", runCommand)

	cc := &container.Config{
		Image: containerImage,
		//Cmd:        runCommand,
		Entrypoint:   entrypoint,
		Cmd:          runCommand,
		Tty:          interactive,
		AttachStderr: true,
		AttachStdin:  interactive,
		AttachStdout: true,
		StdinOnce:    interactive,
		OpenStdin:    interactive,
		Env:          workspace.DumpEnv(),
		WorkingDir:   workspace.currentBlock.workdir.containerPath,
	}

	// Setup mounts
	containerMounts := []mount.Mount{}
	for containerMount := range workspace.mounts {
		m := mount.Mount{
			Type:   mount.TypeBind,
			Source: containerMount,
			Target: workspace.mounts[containerMount],
		}
		containerMounts = append(containerMounts, m)
	}

	hc := &container.HostConfig{
		Mounts: containerMounts,
	}

	err = runContainer(cli, cc, hc)
	return err
}

func (c *Action) Run() error {
	log.Infof("Running Action '%s' of Block '%s'", c.Name, workspace.currentBlock.Name)

	// 1. Determine the correct workdir
	if workspace.currentBlock.workdir.localPath == "" {
		workdir = workspace.path
		workdirContainer = workspace.containerPath
	} else {
		workdir = workspace.currentBlock.workdir.localPath
		workdirContainer = workspace.currentBlock.workdir.containerPath
	}
	log.Debugf("Workdir Local: %s", workdir)
	log.Debugf("Workdir Container: %s", workdirContainer)

	// 3. Determine inventory path
	workspace.registerEnvVar("ANSIBLE_INVENTORY", workspace.currentBlock.getInventoryPath())

	// 4. Determine kubeconfig path
	workspace.registerEnvVar("KUBECONFIG", workspace.currentBlock.getKubeconfigPath())

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

	if c.Interactive {
		// Set interactive=true globally
		interactive = true
	}

	if !local {
		err := c.RunContainer()
		return err
	}

	return nil
}

func (c *Action) saveExecutionScript() error {
	script := c.GetExecutionScript()

	if script != nil {
		f, err := ioutil.TempFile("/tmp", "polycrate."+workspace.Name+".script.*.sh")
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
	} else {
		return fmt.Errorf("'script' section of Action is empty")
	}

}

func (c *Action) GetExecutionScript() []string {
	// Prepare script
	scriptSlice := []string{
		"#!/bin/bash",
		"set -euo pipefail",
	}

	if len(c.Script) > 0 {
		script := append(scriptSlice, c.Script...)
		return script
	}
	return nil
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
	f, err := ioutil.TempFile("/tmp", "cloudstack."+workspace.Name+".run.*.sh")
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
