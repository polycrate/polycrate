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
	"errors"
	"fmt"
	"log"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var activateFlag bool

type promptContent struct {
	errorMsg string
	label    string
}

// releaseCmd represents the release command
var CreateStackCmd = &cobra.Command{
	Use:     "stack [name]",
	Short:   "Create a new stack",
	Long:    `This command creates a new stack. The first argument to the command must be the stack's name.`,
	Aliases: []string{"s"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		setupStack(cmd, args[0])
	},
}

func init() {
	createCmd.AddCommand(CreateStackCmd)

	CreateStackCmd.Flags().BoolVar(&activateFlag, "activate", false, "Activate the stack after creation")
}

func setupStack(cmd *cobra.Command, stackName string) error {
	// Assert stack doesn't exist
	if CheckStackExists(stackName) {
		CheckErr(errors.New("a stack with this name exists"))
	}

	// Set stack name
	viper.Set("stack.name", stackName)

	stackDir := GetStackDir(stackName)

	// Set stack dir
	//viper.Set("stack.dir", stackDir)

	// Ask for Kubernetes cluster
	kubernetesPromptContent := promptContent{
		"Please select Yes or No",
		"Do you have a Kubernetes cluster?",
	}

	hasKubernetes := promptYesNo(kubernetesPromptContent)
	fmt.Println(hasKubernetes)

	var kubeconfig string

	if hasKubernetes == "Yes" {
		// Enable AKP
		viper.Set("akp.enabled", true)

		// Ask for kubeconfig
		kubeconfigPromptContent := promptContent{
			"What kubeconfig should be used?",
			"Please provide the path to a kubeconfig:",
		}
		kubeconfig = promptGetInput(kubeconfigPromptContent)

		// Validate kubeconfig
		if !CheckKubeconfigExists(kubeconfig) {
			CheckErr(errors.New("No kubeconfig found at " + kubeconfig))
		}
		fmt.Println("Set kubeconfig to ", kubeconfig)

		viper.Set("kubeconfig", kubeconfig)
	}

	categoryPromptContent := promptContent{
		"Please provide a category.",
		fmt.Sprintf("What category does %s belong to?", kubeconfig),
	}
	category := promptGetSelect(categoryPromptContent)
	fmt.Println(category)

	//CreateSSHKeyCmd.Run(cmd, []string{})
	fmt.Println("Stack created at ", stackDir)

	stackDir, err := CreateStackDir(stackName)
	CheckErr(err)

	return nil
}

func yesNo() bool {
	prompt := promptui.Select{
		Label: "Select[Yes/No]",
		Items: []string{"Yes", "No"},
	}
	_, result, err := prompt.Run()
	if err != nil {
		log.Fatalf("Prompt failed %v\n", err)
	}
	return result == "Yes"
}
