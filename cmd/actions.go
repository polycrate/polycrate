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
	"path/filepath"
	"strings"

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

type Action struct {
	Metadata    Metadata `mapstructure:"metadata" json:"metadata" validate:"required"`
	Interactive bool     `mapstructure:"interactive,omitempty" json:"interactive,omitempty"`
	Scope       string   `mapstructure:"scope,omitempty" json:"scope,omitempty"`
	Script      []string `mapstructure:"script,omitempty" json:"script,omitempty" validate:"required_if=Action"`
}

func resolveActionPath(actionPath string) error {
	log.Debug("Resolving ActionPath " + actionPath)
	parts := strings.Split(actionPath, ".")

	if len(parts) == 2 {
		blockName = parts[0]
		actionName = parts[1]
		log.Debug("Determined Block '" + blockName + "' and Action '" + actionName + "'")
	} else {
		return goErrors.New("Illegal path (more than 2 parts): " + actionPath)
	}
	return nil
}

func trigger(actionPath string) (int, error) {
	// Resolve the action path
	// This fills variables `blockName` and `actionName` if the path is correct
	// Returns an error if the path is blyat
	err := resolveActionPath(actionPath)
	if err != nil {
		return 1, err
	}

	// Set blockDir
	blockDir = filepath.Join(blocksDir, blockName)
	blockContainerDir = filepath.Join(blocksContainerDir, blockName)

	// Load plugin and command config
	//block = workspace.Blocks[blockName]
	//action := block.Actions[actionName]

	// VALIDATE BLOCK
	//err = block.Validate()
	if err != nil {
		return 1, err
	}

	// VALIDATE ACTION
	err = action.Validate()
	if err != nil {
		return 1, err
	}

	// Set workdir
	if !local {
		workdirContainer = filepath.Join(blockContainerDir)
	} else {
		workdir = filepath.Join(blockDir)
	}
	log.Debug("Changing workdir to " + workdir)

	// Lookup block directory

	return 0, nil
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

func (c *Action) Trigger() (int, error) {
	err, runtimeScriptPath := c.SaveExecutionScript()
	if err != nil {
		return 1, err
	}

	runCommand := []string{"bash", "-c", runtimeScriptPath}

	// register environment variables
	registerEnvVar("POLYCRATE_RUNTIME_SCRIPT_PATH", runtimeScriptPath)

	// register mounts
	registerMount(runtimeScriptPath, runtimeScriptPath)

	if action.Interactive {
		// Set interactive=true globally
		interactive = true
	}

	// Execute the container
	var exitCode int
	if local {
		exitCode, err = RunContainer(
			workspace.Config.Image.Reference,
			workspace.Config.Image.Version,
			runCommand,
		)
	} else {
		exitCode, err = RunCommand("docker", runCommand...)
	}
	if err != nil {
		log.Error("Plugin ", "asd", " failed with exit code ", exitCode, ": ", err.Error())
	} else {
		log.Info("Plugin ", "asd", " succeeded with exit code ", exitCode, ": OK")
	}

	return 0, nil
}

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
