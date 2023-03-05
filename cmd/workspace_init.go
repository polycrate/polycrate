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
	"fmt"
	"os"
	"path/filepath"

	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var withName string
var withSshKeys bool
var withConfig bool

// installCmd represents the install command
var workspaceInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a workspace",
	Long:  ``,
	Args:  cobra.ExactArgs(0), // https://github.com/spf13/cobra/blob/master/user_guide.md
	Run: func(cmd *cobra.Command, args []string) {
		_w := cmd.Flags().Lookup("workspace").Value.String()

		ctx := context.Background()
		ctx, _, cancel, err := polycrate.NewTransaction(ctx, cmd)
		defer polycrate.StopTransaction(ctx, cancel)
		if err != nil {
			log.Fatal(err)
		}

		log := polycrate.GetContextLogger(ctx)
		log.Info("Initializing workspace")

		// Check if config file exists already
		workspaceConfigFilePath := filepath.Join(_w, defaultWorkspace.Config.WorkspaceConfig)
		if _, err := os.Stat(workspaceConfigFilePath); os.IsNotExist(err) {
			log = log.WithField("config", workspace.Config.WorkspaceConfig)
			// Check if a name has been given via flag
			if withName == "" {
				// Ask for a name via prompt
				validate := func(input string) error {
					valid := ValidateMetaDataName(input)
					if !valid {
						return fmt.Errorf("invalid workspace name: %s", input)
					}
					return nil
				}

				prompt := promptui.Prompt{
					Label:    "Workspace name",
					Validate: validate,
				}

				result, err := prompt.Run()

				if err != nil {
					log.Fatalf("Failed to save workspace name: %s", err)
				}
				withName = result
			}

			if !withConfig {

				// Ask if sync with git repo is wanted
				configConsentPrompt := promptui.Prompt{
					Label:     "Do you want to create a config file for this workspace?",
					IsConfirm: true,
				}

				configConsentPromptResult, _ := configConsentPrompt.Run()

				// if err != nil {
				// 	log.Fatalf("Failed to save git repository: %s", err)
				// }
				if configConsentPromptResult == "y" {
					withConfig = true
				}
			}
		} else {
			log.Debug("Config already exists")
			withConfig = false
		}

		if !withSshKeys {

			// Ask if sync with git repo is wanted
			sshKeysConsentPrompt := promptui.Prompt{
				Label:     "Do you want to create SSH keys for this workspace?",
				IsConfirm: true,
			}

			sshKeysConsentPromptResult, _ := sshKeysConsentPrompt.Run()

			// if err != nil {
			// 	log.Fatalf("Failed to save git repository: %s", err)
			// }
			if sshKeysConsentPromptResult == "y" {
				withSshKeys = true
			}
		}

		_, err = polycrate.InitWorkspace(ctx, _w, withName, withSshKeys, withConfig)
		if err != nil {
			log.Fatal(err)
		}

		// // Check if a git repo has been given via flag
		// if withRemoteUrl == "" {
		// 	if !withSync {
		// 		// Ask if sync with git repo is wanted
		// 		gitRepoConsentPrompt := promptui.Prompt{
		// 			Label:     "Do you want to sync this workspace with a git repository",
		// 			IsConfirm: true,
		// 		}

		// 		gitRepoConsentPromptResult, _ := gitRepoConsentPrompt.Run()

		// 		// if err != nil {
		// 		// 	log.Fatalf("Failed to save git repository: %s", err)
		// 		// }
		// 		if gitRepoConsentPromptResult == "y" {
		// 			withSync = true
		// 		}
		// 	}

		// 	if withSync {
		// 		// var group PolycrateProviderGroup
		// 		// if config.Sync.CreateRepo {
		// 		// 	if _workspaceGroup == "" {

		// 		// 		groups, err := config.Gitlab.GetGroups()
		// 		// 		if err != nil {
		// 		// 			log.Fatal(err)
		// 		// 		}

		// 		// 		templates := &promptui.SelectTemplates{
		// 		// 			Label:    "{{ .name }}",
		// 		// 			Active:   "{{ .name | blue }}",
		// 		// 			Inactive: "{{ .name }}",
		// 		// 			Selected: "{{ .name | green }}",
		// 		// 		}

		// 		// 		gitlabGroupsPrompt := promptui.Select{
		// 		// 			Label:     "Choose a group to create the project in",
		// 		// 			Items:     groups,
		// 		// 			Templates: templates,
		// 		// 		}

		// 		// 		// Returns the resulting struct
		// 		// 		index, _, err := gitlabGroupsPrompt.Run()

		// 		// 		if err != nil {
		// 		// 			log.Fatal(err)
		// 		// 		}

		// 		// 		group = groups[index]
		// 		// 	} else {
		// 		// 		var err error
		// 		// 		group, err = config.Gitlab.GetGroup(_workspaceGroup)
		// 		// 		if err != nil {
		// 		// 			log.Fatal(err)
		// 		// 		}
		// 		// 	}

		// 		// 	// Create the project
		// 		// 	project, err := config.Gitlab.CreateProject(group, _workspaceName)
		// 		// 	if err != nil {
		// 		// 		log.Fatal(err)
		// 		// 	}
		// 		// 	printObject(project)

		// 		// }

		// 		workspace.SyncOptions.Enabled = true

		// 		validate := func(input string) error {
		// 			if len(input) <= 0 {
		// 				return errors.New("invalid git remote url")
		// 			}
		// 			return nil
		// 		}

		// 		prompt := promptui.Prompt{
		// 			Label:    "Git remote url",
		// 			Validate: validate,
		// 		}

		// 		result, err := prompt.Run()

		// 		if err != nil {
		// 			log.Fatalf("Failed to save git repository: %s", err)
		// 		}
		// 		withRemoteUrl = result
		// 		workspace.SyncOptions.Remote.Url = result
		// 		log.Debugf("Setting sync url: %s", withRemoteUrl)

		// 		if !withAutoSync {
		// 			// Ask if sync with git repo is wanted
		// 			autoSyncConsentPrompt := promptui.Prompt{
		// 				Label:     "Do you want to enable auto sync?",
		// 				IsConfirm: true,
		// 			}

		// 			autoSyncConsentPromptResult, _ := autoSyncConsentPrompt.Run()

		// 			// if err != nil {
		// 			// 	log.Fatalf("Failed to save git repository: %s", err)
		// 			// }
		// 			if autoSyncConsentPromptResult == "y" {
		// 				withAutoSync = true
		// 			}
		// 		}

		// 		if withAutoSync {
		// 			workspace.SyncOptions.Auto = true
		// 		}
		// 	} else {
		// 		workspace.SyncOptions.Enabled = false
		// 		workspace.SyncOptions.Auto = false
		// 	}

		// } else {
		// 	workspace.SyncOptions.Remote.Url = withRemoteUrl
		// }

		// Check if a git repo has been given via flag

	},
}

func init() {
	workspaceInitCmd.Flags().StringVar(&withName, "with-name", "", "The name of the workspace")
	workspaceInitCmd.Flags().BoolVar(&withSshKeys, "with-ssh-keys", true, "Create SSH keys")
	workspaceInitCmd.Flags().BoolVar(&withConfig, "with-config", true, "Create config file")

	workspaceCmd.AddCommand(workspaceInitCmd)
}
