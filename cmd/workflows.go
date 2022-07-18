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
	"strings"

	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var workflowsCmd = &cobra.Command{
	Use:   "workflow",
	Short: "Control Polycrate Workflows",
	Long:  ``,
	Aliases: []string{
		"workflows",
	},
	Run: func(cmd *cobra.Command, args []string) {
		workspace.load().Flush()
		workspace.ListWorkflows().Flush()
	},
}

func init() {
	rootCmd.AddCommand(workflowsCmd)
}

type Step struct {
	//Metadata    Metadata          `mapstructure:"metadata" json:"metadata" validate:"required"`
	Name        string            `yaml:"name,omitempty" mapstructure:"name,omitempty" json:"name,omitempty" validate:"required,metadata_name"`
	Description string            `yaml:"description,omitempty" mapstructure:"description,omitempty" json:"description,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty" mapstructure:"labels,omitempty" json:"labels,omitempty"`
	Alias       []string          `yaml:"alias,omitempty" mapstructure:"alias,omitempty" json:"alias,omitempty"`
	Block       string            `yaml:"block,omitempty" mapstructure:"block,omitempty" json:"block,omitempty" validate:"required"`
	Action      string            `yaml:"action,omitempty" mapstructure:"action,omitempty" json:"action,omitempty" validate:"required"`
	Workflow    string            `yaml:"workflow,omitempty" mapstructure:"workflow,omitempty" json:"workflow,omitempty"`
	address     string
	//err         error
}

type Workflow struct {
	//Metadata    Metadata          `mapstructure:"metadata" json:"metadata" validate:"required"`
	Name        string            `yaml:"name,omitempty" mapstructure:"name,omitempty" json:"name,omitempty" validate:"required,metadata_name"`
	Description string            `yaml:"description,omitempty" mapstructure:"description,omitempty" json:"description,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty" mapstructure:"labels,omitempty" json:"labels,omitempty"`
	Alias       []string          `yaml:"alias,omitempty" mapstructure:"alias,omitempty" json:"alias,omitempty"`
	Steps       []Step            `yaml:"steps,omitempty" mapstructure:"steps,omitempty" json:"steps,omitempty"`
	address     string
	//err         error
}

func (c *Workflow) Inspect() {
	printObject(c)
}

func (c *Workflow) run() error {
	log.Infof("Running Workflow '%s'", c.Name)

	// Check if any steps are configured
	// Return an error if not
	if len(c.Steps) == 0 {
		return goErrors.New("no steps defined for workflow " + c.Name)
	}

	for index, step := range c.Steps {
		log.Debugf("Running step %d (%s) of workflow %s", index, step.Name, c.Name)

		err := step.run()
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Step) run() error {
	// Get workflow from step
	workflow := workspace.GetWorkflowFromIndex(c.Workflow)

	log.Infof("Running step '%s' of workflow '%s'", c.Name, workflow.Name)

	// Reloading Workspace to discover new files
	workspace.load().Flush()

	// Check if an a block and an action have been configured
	if c.Block == "" {
		return goErrors.New("no block configured for step " + c.Name + " of workflow " + workflow.Name)
	}
	if c.Action == "" {
		return goErrors.New("no action configured for step " + c.Name + " of workflow " + workflow.Name)
	}

	// Run the configured action
	log.Debugf("Running action '%s' of block '%s'")

	workspace.registerCurrentStep(c)

	actionAddress := strings.Join([]string{c.Block, c.Action}, ".")
	workspace.RunAction(actionAddress).Flush()

	return nil
}

func (c *Workflow) validate() error {
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

func (c *Step) validate() error {
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

func (c *Workflow) getStepByName(stepName string) *Step {

	//for _, block := range c.Blocks {
	for i := 0; i < len(c.Steps); i++ {
		step := &c.Steps[i]
		if step.Name == stepName {
			return step
		}
	}
	return nil
}
