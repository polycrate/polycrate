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
		workspace.load()
		if workspace.Flush() != nil {
			log.Fatal(workspace.Flush)
		}
	},
}

func init() {
	rootCmd.AddCommand(workflowsCmd)
}

type Step struct {
	//Metadata    Metadata          `mapstructure:"metadata" json:"metadata" validate:"required"`
	Name        string            `mapstructure:"name" json:"name" validate:"required,metadata_name"`
	Description string            `mapstructure:"description" json:"description"`
	Labels      map[string]string `mapstructure:"labels" json:"labels"`
	Alias       []string          `mapstructure:"alias" json:"alias"`
	Block       string            `mapstructure:"block" json:"block" validate:"required"`
	Action      string            `mapstructure:"action" json:"action" validate:"required"`
	Workflow    Workflow          `mapstructure:"workflow,omitempty" json:"workflow,omitempty"`
	address     string
	//err         error
}

type Workflow struct {
	//Metadata    Metadata          `mapstructure:"metadata" json:"metadata" validate:"required"`
	Name        string            `mapstructure:"name" json:"name" validate:"required,metadata_name"`
	Description string            `mapstructure:"description" json:"description"`
	Labels      map[string]string `mapstructure:"labels" json:"labels"`
	Alias       []string          `mapstructure:"alias" json:"alias"`
	Steps       []Step            `mapstructure:"steps,omitempty" json:"steps,omitempty"`
	address     string
	//err         error
}

func (c *Workflow) Inspect() {
	printObject(c)
}

func (c *Workflow) RunStep(address string) error {
	//step := workspace.LookupStep(address)
	return nil
}

func (c *Workflow) Run() error {
	log.Infof("Running Workflow '%s'", c.Name)

	// Check if any steps are configured
	// Return an error if not
	if len(c.Steps) == 0 {
		return goErrors.New("no steps defined for workflow " + c.Name)
	}

	for index, step := range c.Steps {
		log.Debugf("Running step %d (%s) of workflow %s", index, step.Name, c.Name)

		err := step.Run()
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Step) Run() error {
	//printObject(c)

	log.Infof("Running Step '%s' of workflow ", c.Name, c.Workflow.Name)

	// Check if an a block and an action have been configured
	if c.Block == "" {
		return goErrors.New("no block configured for step " + c.Name + " of workflow " + c.Workflow.Name)
	}
	if c.Action == "" {
		return goErrors.New("no action configured for step " + c.Name + " of workflow " + c.Workflow.Name)
	}

	// Run the configured action
	log.Debugf("Running action '%s' of block '%s'")

	workspace.registerCurrentStep(c)

	err := workspace.RunAction(strings.Join([]string{c.Block, c.Action}, "."))
	if err != nil {
		return err
	}

	return nil
}

func (c *Workflow) Validate() error {
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

func (c *Step) Validate() error {
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
