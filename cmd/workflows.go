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
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var workflowsCmd = &cobra.Command{
	Use:   "workflows",
	Short: "Control Polycrate Workflows",
	Long:  ``,
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
	Metadata Metadata `mapstructure:"metadata" json:"metadata" validate:"required"`
	Block    string   `mapstructure:"block" json:"block" validate:"required"`
	Action   string   `mapstructure:"action" json:"action" validate:"required"`
	workflow *Workflow
	address  string
	err      error
}

type Workflow struct {
	Metadata Metadata `mapstructure:"metadata" json:"metadata" validate:"required"`
	Steps    []Step   `mapstructure:"steps,omitempty" json:"steps,omitempty"`
	address  string
	err      error
}

func (c *Workflow) Inspect() {
	printObject(c)
}
