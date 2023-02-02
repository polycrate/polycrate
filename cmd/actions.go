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
	"bytes"
	"context"
	goErrors "errors"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/go-playground/validator/v10"
	"github.com/imdario/mergo"
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
		ctx, cancelFunc := context.WithCancel(context.Background())
		ctx, err := polycrate.StartTransaction(ctx, cancelFunc)
		if err != nil {
			polycrate.ContextExit(ctx, cancelFunc, err)
		}

		log := polycrate.GetContextLogger(ctx)

		workspace, err := polycrate.LoadWorkspace(ctx, cmd.Flags().Lookup("workspace").Value.String(), true)
		if err != nil {
			polycrate.ContextExit(ctx, cancelFunc, err)
		}

		log = log.WithField("workspace", workspace.Name)
		ctx = polycrate.SetContextLogger(ctx, log)

		workspace.ListActions().Flush()

		polycrate.ContextExit(ctx, cancelFunc, err)
	},
}

func init() {
	rootCmd.AddCommand(actionsCmd)
}

// type ActionAnsibleConfig struct {
// 	Inventory string `yaml:"inventory,omitempty" mapstructure:"inventory,omitempty" json:"inventory,omitempty"`
// 	Hosts     string `yaml:"hosts,omitempty" mapstructure:"hosts,omitempty" json:"hosts,omitempty"`
// }

// type ActionKubernetesConfig struct {
// 	Kubeconfig string `yaml:"kubeconfig,omitempty" mapstructure:"kubeconfig,omitempty" json:"kubeconfig,omitempty"`
// 	Context    string `yaml:"context,omitempty" mapstructure:"context,omitempty" json:"context,omitempty"`
// }
type Action struct {
	//Metadata            Metadata               `mapstructure:"metadata,squash" json:"metadata" validate:"required"`
	Name        string            `yaml:"name,omitempty" mapstructure:"name,omitempty" json:"name,omitempty" validate:"required,metadata_name"`
	Description string            `yaml:"description,omitempty" mapstructure:"description,omitempty" json:"description,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty" mapstructure:"labels,omitempty" json:"labels,omitempty"`
	Alias       []string          `yaml:"alias,omitempty" mapstructure:"alias,omitempty" json:"alias,omitempty"`
	Interactive bool              `yaml:"interactive,omitempty" mapstructure:"interactive,omitempty" json:"interactive,omitempty"`
	Script      []string          `yaml:"script,omitempty" mapstructure:"script,omitempty" json:"script,omitempty" validate:"required_without=Playbook,excluded_with=Playbook"`
	Playbook    string            `yaml:"playbook,omitempty" mapstructure:"playbook,omitempty" json:"playbook,omitempty" validate:"required_without=Script,excluded_with=Script"`
	//Ansible             ActionAnsibleConfig    `yaml:"ansible,omitempty" mapstructure:"ansible,omitempty" json:"ansible,omitempty"`
	//Kubernetes          ActionKubernetesConfig `yaml:"kubernetes,omitempty" mapstructure:"kubernetes,omitempty" json:"kubernetes,omitempty"`
	executionScriptPath string
	address             string
	Block               string                 `yaml:"block,omitempty" mapstructure:"block,omitempty" json:"block,omitempty"`
	Config              map[string]interface{} `yaml:"config,omitempty" mapstructure:"config,omitempty" json:"config,omitempty"`
}

func (c *Action) MergeIn(action Action) error {

	if err := mergo.Merge(c, action); err != nil {
		log.Fatal(err)
	}
	return nil
}

// func (c *Action) RunContainer() error {
// 	containerImage := strings.Join([]string{workspace.Config.Image.Reference, workspace.Config.Image.Version}, ":")
// 	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

// 	if err != nil {
// 		return err
// 	}
// 	log.WithFields(log.Fields{
// 		"workspace": workspace.Name,
// 		"action":    c.Name,
// 		"block":     c.Block,
// 		"build":     build,
// 		"pull":      pull,
// 	}).Debugf("Running container")

// 	// Check if a Dockerfile is configured in the Workspace
// 	if workspace.Config.Dockerfile != "" {
// 		// Create the filepath
// 		dockerfilePath := filepath.Join(workspace.LocalPath, workspace.Config.Dockerfile)

// 		// Check if the file exists
// 		if _, err := os.Stat(dockerfilePath); !os.IsNotExist(err) {
// 			if build {
// 				// We need to build and tag this
// 				log.WithFields(log.Fields{
// 					"workspace": workspace.Name,
// 					"action":    c.Name,
// 					"block":     c.Block,
// 					"path":      dockerfilePath,
// 					"build":     build,
// 					"pull":      pull,
// 				}).Debugf("Dockerfile detected")

// 				tag := workspace.Name + ":" + version
// 				log.WithFields(log.Fields{
// 					"workspace": workspace.Name,
// 					"action":    c.Name,
// 					"block":     c.Block,
// 					"path":      dockerfilePath,
// 					"image":     tag,
// 					"build":     build,
// 					"pull":      pull,
// 				}).Warnf("Building image")

// 				tags := []string{tag}
// 				containerImage, err = buildContainerImage(workspace.Config.Dockerfile, tags)
// 				if err != nil {
// 					return err
// 				}
// 			} else {
// 				if pull {
// 					log.WithFields(log.Fields{
// 						"workspace": workspace.Name,
// 						"action":    c.Name,
// 						"block":     c.Block,
// 						"image":     containerImage,
// 						"build":     build,
// 						"pull":      pull,
// 					}).Debugf("Pulling image")
// 					err := pullContainerImage(containerImage)

// 					if err != nil {
// 						return err
// 					}
// 				} else {
// 					log.WithFields(log.Fields{
// 						"workspace": workspace.Name,
// 						"action":    c.Name,
// 						"block":     c.Block,
// 						"image":     containerImage,
// 						"build":     build,
// 						"pull":      pull,
// 					}).Debugf("Not pulling/building image")
// 				}
// 			}
// 		} else {
// 			if pull {
// 				log.WithFields(log.Fields{
// 					"workspace": workspace.Name,
// 					"action":    c.Name,
// 					"block":     c.Block,
// 					"image":     containerImage,
// 					"build":     build,
// 					"pull":      pull,
// 				}).Debugf("Pulling image")
// 				err := pullContainerImage(containerImage)

// 				if err != nil {
// 					return err
// 				}
// 			} else {
// 				log.WithFields(log.Fields{
// 					"workspace": workspace.Name,
// 					"action":    c.Name,
// 					"block":     c.Block,
// 					"image":     containerImage,
// 					"build":     build,
// 					"pull":      pull,
// 				}).Debugf("Not pulling/building image")
// 			}
// 		}
// 	} else {
// 		if pull {
// 			log.WithFields(log.Fields{
// 				"workspace": workspace.Name,
// 				"action":    c.Name,
// 				"block":     c.Block,
// 				"image":     containerImage,
// 				"build":     build,
// 				"pull":      pull,
// 			}).Debugf("Pulling image")
// 			err := pullContainerImage(containerImage)

// 			if err != nil {
// 				return err
// 			}
// 		} else {
// 			log.WithFields(log.Fields{
// 				"workspace": workspace.Name,
// 				"action":    c.Name,
// 				"block":     c.Block,
// 				"image":     containerImage,
// 				"build":     build,
// 				"pull":      pull,
// 			}).Debugf("Not pulling/building image")
// 		}
// 	}

// 	entrypoint := []string{"bash", "-c"}
// 	runCommand := []string{}

// 	if c.executionScriptPath != "" {
// 		log.WithFields(log.Fields{
// 			"workspace": workspace.Name,
// 			"action":    c.Name,
// 			"block":     c.Block,
// 			"path":      c.executionScriptPath,
// 		}).Debugf("Running script")

// 		runCommand = append(runCommand, c.executionScriptPath)
// 	} else {
// 		return goErrors.New("no execution script path given. Nothing to do")
// 	}
// 	log.WithFields(log.Fields{
// 		"workspace": workspace.Name,
// 		"action":    c.Name,
// 		"block":     c.Block,
// 	}).Debugf("Running entrypoint %s", entrypoint)

// 	log.WithFields(log.Fields{
// 		"workspace": workspace.Name,
// 		"action":    c.Name,
// 		"block":     c.Block,
// 	}).Debugf("Running command %s", runCommand)

// 	containerName := slugify([]string{workspace.Name, workspace.currentBlock.Name, c.Name})

// 	cc := &container.Config{
// 		Image: containerImage,
// 		//Cmd:        runCommand,
// 		Entrypoint: entrypoint,
// 		Cmd:        runCommand,
// 		Tty:        true,
// 		Labels: map[string]string{
// 			"polycrate.workspace": workspace.Name,
// 			"polycrate.name":      containerName,
// 		},
// 		AttachStderr: true,
// 		AttachStdin:  interactive,
// 		AttachStdout: true,
// 		StdinOnce:    interactive,
// 		OpenStdin:    interactive,
// 		Env:          workspace.DumpEnv(),
// 		WorkingDir:   workspace.currentBlock.Workdir.ContainerPath,
// 	}

// 	// Setup mounts
// 	containerMounts := []mount.Mount{}
// 	for containerMount := range workspace.mounts {
// 		m := mount.Mount{
// 			Type:   mount.TypeBind,
// 			Source: containerMount,
// 			Target: workspace.mounts[containerMount],
// 		}
// 		containerMounts = append(containerMounts, m)
// 	}

// 	hc := &container.HostConfig{
// 		Mounts: containerMounts,
// 	}

// 	err = runContainer(cli, cc, hc, containerName)
// 	// containerName := slugify([]string{workspace.Name, workspace.currentBlock.Name, c.Name})
// 	// workspace.RunContainer(containerName, workspace.ContainerPath, runCommand).Flush()
// 	return err
// }

func (a *Action) RunWithContext(ctx context.Context) (context.Context, error) {
	log := polycrate.GetContextLogger(ctx)

	log.Infof("Running action")

	workspace, err := polycrate.GetContextWorkspace(ctx)
	if err != nil {
		return ctx, err
	}

	block, err := polycrate.GetContextBlock(ctx)
	if err != nil {
		return ctx, err
	}

	log.Debugf("Running action")

	// 3. Determine inventory path
	inventoryPath := block.getInventoryPath(workspace)
	workspace.registerEnvVar("ANSIBLE_INVENTORY", inventoryPath)
	log.Tracef("Updating inventory: %s", inventoryPath)

	// 4. Determine kubeconfig path
	kubeconfigPath := block.getKubeconfigPath()
	workspace.registerEnvVar("KUBECONFIG", kubeconfigPath)
	log.Tracef("Updating kubeconfig: %s", kubeconfigPath)

	// register environment variables
	workspace.registerEnvVar("POLYCRATE_RUNTIME_SCRIPT_PATH", a.executionScriptPath)

	// Wrapup
	if a.Interactive {
		// Set interactive=true globally
		interactive = true
	}

	if snapshot {
		workspace.Snapshot(ctx)
	} else {
		// Save snapshot before running the action
		if snapshotContainerPath, err := workspace.SaveSnapshot(ctx); err != nil {
			return ctx, err
		} else {
			// Save execution script
			var err error
			if len(a.Script) > 0 {
				err = a.SaveExecutionScript(ctx)
				// We use the vars plugin in "script" mode
				workspace.registerEnvVar("ANSIBLE_VARS_ENABLED", "polycrate_vars")
			} else if a.Playbook != "" {
				err = a.saveAnsibleScript(ctx, snapshotContainerPath)
			} else {
				err = fmt.Errorf("neither 'script' nor 'playbook' have been defined")
			}

			if err != nil {
				return ctx, err
			}

			// register mounts
			workspace.registerMount(a.executionScriptPath, a.executionScriptPath)

			if !local {
				txid := polycrate.GetContextTXID(ctx)
				containerName := txid.String()

				runCommand := []string{}
				if a.executionScriptPath != "" {
					log.Debugf("Running script: %s", a.executionScriptPath)

					runCommand = append(runCommand, a.executionScriptPath)
				} else {
					return ctx, goErrors.New("no execution script path given. Nothing to do")
				}

				ctx, err = workspace.RunContainerWithContext(ctx, containerName, block.Workdir.Path, runCommand)
				if err != nil {
					return ctx, err
				}
			} else {
				err := fmt.Errorf("'local' mode not yet implemented")
				return ctx, err
			}
		}
	}
	return ctx, nil
}

func (c *Action) SaveExecutionScript(ctx context.Context) error {
	txid := polycrate.GetContextTXID(ctx)
	workspace, err := polycrate.GetContextWorkspace(ctx)
	if err != nil {
		return err
	}
	log := polycrate.GetContextLogger(ctx)

	script := c.GetExecutionScript(ctx)
	snapshot := workspace.GetSnapshot(ctx)

	scriptSlug := slugify([]string{txid.String(), "execution", "script"})
	scriptFilename := strings.Join([]string{scriptSlug, "sh"}, ".")

	if script != nil {
		f, err := polycrate.getTempFile(ctx, scriptFilename)
		if err != nil {
			return err
		}

		datawriter := bufio.NewWriter(f)

		for _, data := range script {
			// Load the script line into a template object and parse it
			// return an error if the parsing fails
			t, err := template.New("script-line").Parse(data)
			if err != nil {
				return err
			}

			// Execute the template and save the substituted content to a var
			var substitutedScriptLine bytes.Buffer
			err = t.Execute(&substitutedScriptLine, snapshot)
			if err != nil {
				return err
			}
			// Append the substituted script line to the script
			_, _ = datawriter.WriteString(substitutedScriptLine.String() + "\n")
		}

		datawriter.Flush()
		log = log.WithField("path", f.Name())
		log.Debug("Saved temporary execution script")

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
		return fmt.Errorf("'script' section of action is empty")
	}

}

func (a *Action) saveAnsibleScript(ctx context.Context, snapshotContainerPath string) error {
	log := polycrate.GetContextLogger(ctx)

	// Prepare script
	scriptSlice := []string{
		"#!/bin/bash",
		"set -euo pipefail",
	}

	scriptString := fmt.Sprintf("ansible-playbook -e '@%s' %s", snapshotContainerPath, a.Playbook)
	script := append(scriptSlice, scriptString)
	snapshot := workspace.GetSnapshot(ctx)

	txid := polycrate.GetContextTXID(ctx)
	scriptSlug := slugify([]string{txid.String(), "execution", "script"})
	scriptFilename := strings.Join([]string{scriptSlug, "sh"}, ".")

	if script != nil {
		f, err := polycrate.getTempFile(ctx, scriptFilename)
		if err != nil {
			return err
		}

		datawriter := bufio.NewWriter(f)

		for _, data := range script {
			// Load the script line into a template object and parse it
			// return an error if the parsing fails
			t, err := template.New("script-line").Parse(data)
			if err != nil {
				return err
			}

			// Execute the template and save the substituted content to a var
			var substitutedScriptLine bytes.Buffer
			err = t.Execute(&substitutedScriptLine, snapshot)
			if err != nil {
				return err
			}
			// Append the substituted script line to the script
			_, _ = datawriter.WriteString(substitutedScriptLine.String() + "\n")
		}

		datawriter.Flush()
		log = log.WithField("path", f.Name())
		log.Debug("Saved temporary execution script")

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
		a.executionScriptPath = f.Name()
		return nil
	} else {
		return fmt.Errorf("ansible: 'script' section of action is empty")
	}

}

func (c *Action) GetExecutionScript(ctx context.Context) []string {
	log := polycrate.GetContextLogger(ctx)

	// Prepare script
	scriptSlice := []string{
		"#!/bin/bash",
		"set -euo pipefail",
	}

	if len(c.Script) > 0 {
		// Loop over script slice, convert interface to string
		scriptStrings := []string{}
		for _, scriptLine := range c.Script {
			scriptLineStr := fmt.Sprintf("%v", scriptLine)
			scriptStrings = append(scriptStrings, scriptLineStr)
			log.Debugf("scriptLine as String: %s", scriptLineStr)
		}
		script := append(scriptSlice, scriptStrings...)
		return script
	}
	return nil
}

func (c *Action) GetAnsibleScript(varsPath string, playbook string) []string {
	// Prepare script
	scriptSlice := []string{
		"#!/bin/bash",
		"set -euo pipefail",
	}

	scriptString := fmt.Sprintf("ansible-playbook -e '@%s' %s", varsPath, playbook)
	script := append(scriptSlice, scriptString)

	return script
}

func (a *Action) validate() error {
	err := validate.Struct(a)
	if err != nil {

		// this check is only needed when your code could produce
		// an invalid value for validation such as interface with nil
		// value most including myself do not usually have code like this.
		if _, ok := err.(*validator.InvalidValidationError); ok {
			log.Error(err)
			return nil
		}

		for _, err := range err.(validator.ValidationErrors) {
			if err.Tag() == "excluded_with" {
				log.WithFields(log.Fields{
					"workspace": workspace.Name,
					"action":    a.Name,
					"block":     a.Block,
					"namespace": strings.ToLower(err.Namespace()),
				}).Errorf("Configuration option '%s' is mutually exclusive with '%s'", strings.ToLower(err.Field()), strings.ToLower(err.Param()))
				//log.Errorf("Configuration option '%s' is mutually exclusive with '%s'", strings.ToLower(err.Namespace()), strings.ToLower(err.Param()))
				err.Field()
			} else {
				log.Error("Configuration option `" + strings.ToLower(err.Namespace()) + "` failed to validate: " + err.Tag() + "=" + strings.ToLower(err.Param()))
			}
		}

		// from here you can create your own error messages in whatever language you wish
		return fmt.Errorf("error validating Action '%s' of Block '%s'", a.Name, a.Block)
	}
	// if a.Script == nil {
	// 	return goErrors.New("no script found for Action")
	// }
	return nil
}

func (c *Action) ValidateScript() error {
	if c.Script == nil {
		return goErrors.New("no script found for Action")
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

// func (c *Action) SaveExecutionScript() (string, error) {
// 	f, err := ioutil.TempFile("/tmp", "cloudstack."+workspace.Name+".run.*.sh")
// 	if err != nil {
// 		return err, ""
// 	}
// 	datawriter := bufio.NewWriter(f)

// 	for _, data := range c.GetExecutionScript() {
// 		_, _ = datawriter.WriteString(data + "\n")
// 	}
// 	datawriter.Flush()
// 	log.Debug("Saved script to " + f.Name())

// 	err = os.Chmod(f.Name(), 0755)
// 	if err != nil {
// 		return err, ""
// 	}

// 	// Closing file descriptor
// 	// Getting fatal errors on windows WSL2 when accessing
// 	// the mounted script file from inside the container if
// 	// the file descriptor is still open
// 	// Works flawlessly with open file descriptor on M1 Mac though
// 	// It's probably safer to close the fd anyways
// 	f.Close()
// 	return nil, f.Name()
// }

func (c *Action) Inspect() {
	printObject(c)
}
