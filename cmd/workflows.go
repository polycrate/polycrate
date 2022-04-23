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
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var pipelinesCmd = &cobra.Command{
	Use:   "pipelines",
	Short: "Control Cloudstack pipelines",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		loadWorkspace()
		// List pipelines
	},
}

func init() {
	rootCmd.AddCommand(pipelinesCmd)
}

type Step struct {
	Display string   `mapstructure:"display" json:"display" validate:"required"`
	Block   string   `mapstructure:"block" json:"block" validate:"required"`
	Action  string   `mapstructure:"action" json:"action" validate:"required"`
	Trigger string   `mapstructure:"trigger" json:"trigger"`
	Scope   string   `mapstructure:"scope" json:"scope"`
	Labels  []string `mapstructure:"labels,omitempty" json:"labels,omitempty"`
}

type Workflow struct {
	Metadata Metadata `mapstructure:"metadata" json:"metadata" validate:"required"`
	Steps    []Step   `mapstructure:"steps,omitempty" json:"steps,omitempty"`
}
