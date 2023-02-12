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
	"context"
	goErrors "errors"
	"fmt"
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
		_w := cmd.Flags().Lookup("workspace").Value.String()

		ctx := context.Background()
		ctx, cancel, err := polycrate.NewTransaction(ctx, cmd)
		defer polycrate.StopTransaction(ctx, cancel)
		if err != nil {
			log.Fatal(err)
		}

		log := polycrate.GetContextLogger(ctx)

		_, workspace, err := polycrate.GetWorkspaceWithContext(ctx, _w, true)
		if err != nil {
			log.Fatal(err)
		}

		err = workspace.ListWorkflows()
		if err != nil {
			log.Fatal(err)
		}
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
	Block       string            `yaml:"block" mapstructure:"block" json:"block" validate:"required"`
	Action      string            `yaml:"action" mapstructure:"action" json:"action" validate:"required"`
	Prompt      Prompt            `yaml:"prompt,omitempty" mapstructure:"prompt,omitempty" json:"prompt,omitempty"`
	Workflow    string            `yaml:"workflow,omitempty" mapstructure:"workflow,omitempty" json:"workflow,omitempty"`
	address     string
	//err         error
}

type Workflow struct {
	//Metadata    Metadata          `mapstructure:"metadata" json:"metadata" validate:"required"`
	Name         string            `yaml:"name,omitempty" mapstructure:"name,omitempty" json:"name,omitempty" validate:"required,metadata_name"`
	Description  string            `yaml:"description,omitempty" mapstructure:"description,omitempty" json:"description,omitempty"`
	Labels       map[string]string `yaml:"labels,omitempty" mapstructure:"labels,omitempty" json:"labels,omitempty"`
	Alias        []string          `yaml:"alias,omitempty" mapstructure:"alias,omitempty" json:"alias,omitempty"`
	Steps        []Step            `yaml:"steps,omitempty" mapstructure:"steps,omitempty" json:"steps,omitempty"`
	Prompt       Prompt            `yaml:"prompt,omitempty" mapstructure:"prompt,omitempty" json:"prompt,omitempty"`
	AllowFailure bool              `yaml:"allow_failure,omitempty" mapstructure:"allow_failure,omitempty" json:"allow_failure,omitempty"`
	address      string
	//err         error
}

func (c *Workflow) Inspect() {
	printObject(c)
}

func (w *Workflow) RunWithContext(ctx context.Context, stepName string) (context.Context, error) {
	log := polycrate.GetContextLogger(ctx)

	log.Infof("Running Workflow")

	// Check if a prompt is configured and execute it
	if w.Prompt.Message != "" {
		result := w.Prompt.Validate(ctx)
		if !result {
			return ctx, fmt.Errorf("not running workflow. user confirmation declined")
		}
	}

	// Check if any steps are configured
	// Return an error if not
	if len(w.Steps) == 0 {
		return ctx, goErrors.New("no steps defined for workflow " + w.Name)
	}

	// If a step name has been given, only run this step
	if stepName != "" {
		ctx, step, err := w.GetStepWithContext(ctx, stepName)
		if err != nil {
			return ctx, err
		}

		return ctx, step.Run(ctx)
	}

	for _, step := range w.Steps {
		ctx, step, err := w.GetStepWithContext(ctx, step.Name)
		if err != nil {
			return ctx, err
		}

		err = step.Run(ctx)
		if err != nil {
			// Check AllowFailure, move on if it's OK
			if !w.AllowFailure {
				return ctx, err
			}
			log.Warnf("Step exited with an error: '%s'; continuing workflow execution because `allow_failure` is true", err)
		}

		// reloading workspace to account for new artifacts
		//workspace.Reload(ctx)
	}

	return ctx, nil
}
func (w *Workflow) Run(ctx context.Context) error {
	log := polycrate.GetContextLogger(ctx)

	// Check if a prompt is configured and execute it
	if w.Prompt.Message != "" {
		result := w.Prompt.Validate(ctx)
		if !result {
			return fmt.Errorf("not running workflow. user confirmation declined")
		}
	}

	workspace, err := polycrate.GetContextWorkspace(ctx)
	if err != nil {
		return err
	}

	log.Infof("Running Workflow")

	// Check if any steps are configured
	// Return an error if not
	if len(w.Steps) == 0 {
		return goErrors.New("no steps defined for workflow " + w.Name)
	}

	for _, step := range w.Steps {
		log = log.WithField("step", step.Name)
		ctx = polycrate.SetContextLogger(ctx, log)

		err := step.Run(ctx)
		if err != nil {
			return err
		}

		// reloading workspace to account for new artifacts
		workspace.Reload(ctx, true)
	}

	return nil
}

func (s *Step) Run(ctx context.Context) error {
	log := polycrate.GetContextLogger(ctx)

	workspace, err := polycrate.GetContextWorkspace(ctx)
	if err != nil {
		return err
	}

	// // Get workflow from step
	// workflow := workspace.GetWorkflowFromIndex(s.Workflow)

	log.Infof("Running step")

	// Reloading Workspace to discover new files
	//workspace.load().Flush()

	// Check if an a block and an action have been configured
	if s.Block == "" {
		return goErrors.New("no block configured")
	}
	if s.Action == "" {
		return goErrors.New("no action configured")
	}

	workspace.registerCurrentStep(s)

	// Check for prompt
	var runStep = true
	if s.Prompt.Message != "" {
		runStep = s.Prompt.Validate(ctx)
	}

	if runStep {
		_, err = workspace.RunActionWithContext(ctx, s.Block, s.Action)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("not running step. user confirmation declined")
	}

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
func (w *Workflow) GetStepWithContext(ctx context.Context, name string) (context.Context, *Step, error) {
	step, err := w.GetStep(name)
	if err != nil {
		return ctx, nil, err
	}

	stepKey := ContextKey("step")
	ctx = context.WithValue(ctx, stepKey, step)

	log := polycrate.GetContextLogger(ctx)
	log = log.WithField("step", step.Name)
	ctx = polycrate.SetContextLogger(ctx, log)

	return ctx, step, nil
}

func (c *Workflow) GetStep(name string) (*Step, error) {

	//for _, block := range c.Blocks {
	for i := 0; i < len(c.Steps); i++ {
		step := &c.Steps[i]
		if step.Name == name {
			return step, nil
		}
	}
	return nil, fmt.Errorf("step not found: %s", name)
}
func (c *Workflow) GetStepByIndex(index int) *Step {

	//for _, block := range c.Blocks {
	for i := 0; i < len(c.Steps); i++ {
		step := &c.Steps[i]
		if i == index {
			return step
		}
	}
	return nil
}
