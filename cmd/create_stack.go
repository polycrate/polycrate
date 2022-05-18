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

var activateFlag bool

type promptContent struct {
	errorMsg string
	label    string
}

// releaseCmd represents the release command
var CreateStackCmd = &cobra.Command{
	Hidden:  true,
	Use:     "stack [name]",
	Short:   "Create a new stack",
	Long:    `This command creates a new stack. The first argument to the command must be the stack's name.`,
	Aliases: []string{"s"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		//setupStack(cmd, args[0])
		log.Warn("Comming soon! Check https://polycrate.io for more")
	},
}

func init() {
	createCmd.AddCommand(CreateStackCmd)

	CreateStackCmd.Flags().BoolVar(&activateFlag, "activate", false, "Activate the stack after creation")
}
