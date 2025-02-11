/*
Copyright © 2021 Fabian Peter <fp@ayedo.de>

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

	//"strconv"

	"context"

	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

// rootCmd represents the base command when called without any subcommands

func Execute() {
	//_rootCmd := newRootCmd(os.Args[1:])
	//ctx := context.Background()

	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}

}

var rootCmd = &cobra.Command{
	Use:   "polycrate",
	Short: "Polycrate ist ein Framework zum Entwickeln von Plattformen",
	Long: `Polycrate
		
	Polycrate is a framework for platform development.
		
	Learn more at https://docs.polycrate.io
		`,

	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		globalCmd = cmd

	},
	Version: version,
}

func init() {
	cobra.OnInitialize(initConfig)

	fs := rootCmd.PersistentFlags()

	// Action related
	fs.BoolVarP(&local, "local", "l", false, "Run actions locally (without the polycrate container). Defaults to false.")
	fs.BoolVarP(&interactive, "interactive", "i", false, "Make the container interactive and accept input from stdin. Like '-it' for Docker.")
	fs.BoolVarP(&pull, "pull", "p", true, "Pull the workspace image before running the container. Defaults to true.")
	fs.BoolVar(&blocksAutoPull, "blocks-auto-pull", false, "Automatically pull blocks that are missing from the workspace. Defaults to false.")
	fs.BoolVarP(&build, "build", "b", true, "When this is true, a custom image will be built from the workspace Dockerfile. This image will then be used to run the action. Defaults to true.")

	// CLI related
	fs.StringVarP(&outputFormat, "output-format", "o", "yaml", "Output format (currently no-op).")
	//fs.BoolVarP(&force, "force", "f", false, "Force whatever you want to do. Like sudo with more willpower. Defaults to false.")
	fs.BoolVarP(&snapshot, "snapshot", "", false, "Only dump the workspace snapshot, do not run anything.")
	fs.BoolVar(&dev, "dev", false, "Enable development mode for working with blocks")
	fs.StringVar(&editor, "editor", DefaultEditor, "Editor to use to open the workspace")

	// Polycrate main config
	fs.StringVar(&polycrateConfigDir, "config-dir", polycrateConfigDir, "Path to the config directory")
	fs.StringVar(&polycrateConfigFilePath, "config-file", polycrateConfigFilePath, "Path to the config file")
	fs.StringVar(&polycrateWorkspaceDir, "workspace-dir", polycrateWorkspaceDir, "Path to the workspaces")
	fs.StringVar(&polycrateRuntimeDir, "runtime-dir", polycrateRuntimeDir, "Path to the runtime directory")

	// Workspace path/name
	fs.StringVarP(&defaultWorkspace.LocalPath, "workspace", "w", cwd, "The path to the workspace. Defaults to $PWD")

	// Common config
	fs.IntVar(&polycrate.Config.Loglevel, "loglevel", 1, "How verbose Polycrate should be")
	fs.BoolVar(&polycrate.Config.CheckUpdates, "check-updates", false, "Check for Polycrate updates on program start")
	fs.BoolVar(&polycrate.Config.Experimental.MergeV2, "merge-v2", true, "Feature flag: use merge v2")
	fs.StringVar(&polycrate.Config.Logformat, "logformat", "default", "Log format (JSON/YAML)")
	fs.StringVar(&polycrate.Config.Kubeconfig, "kubeconfig", KubeconfigPath, "Path to a global kubeconfig")
	fs.StringVar(&polycrate.Config.Registry.Url, "registry-url", RegistryUrl, "The URL of the OCI registry")
	fs.StringVar(&polycrate.Config.Registry.BaseImage, "registry-base-image", RegistryBaseImage, "The base image to package blocks in OCI format")

	// Workspace related config
	fs.StringVar(&defaultWorkspace.SyncOptions.Local.Branch.Name, "sync-local-branch", GitDefaultBranch, "Default git branch")
	fs.StringVar(&defaultWorkspace.SyncOptions.Remote.Branch.Name, "sync-remote-branch", GitDefaultBranch, "Default git branch")
	fs.StringVar(&defaultWorkspace.SyncOptions.Remote.Name, "sync-remote-name", GitDefaultRemote, "Default git remote")
	fs.BoolVar(&defaultWorkspace.SyncOptions.Enabled, "sync-enabled", false, "Sync enabled")
	fs.BoolVar(&defaultWorkspace.SyncOptions.Auto, "sync-auto", false, "Sync automatically")
	fs.StringVar(&defaultWorkspace.Events.Handler, "event-handler", WorkspaceEventHandler, "Default event handler")
	fs.StringVar(&defaultWorkspace.Events.Endpoint, "event-endpoint", "", "Default event endpoint")
	fs.BoolVar(&defaultWorkspace.Events.Commit, "event-commit", false, "Auto-commit each event")
	fs.StringSliceVarP(&defaultWorkspace.ExtraEnv, "env", "e", []string{}, "Additional environment variables for the workspace in the format 'KEY=value'")
	fs.StringSliceVarP(&defaultWorkspace.ExtraMounts, "mount", "m", []string{}, "Additional mounts for the workspace container in the format '/host:/container'. This will be ignored when used with --local")
	fs.StringVar(&defaultWorkspace.Config.WorkspaceConfig, "workspace-config", WorkspaceConfigFile, "The name of the config file that holds the workspace config.")
	fs.StringVar(&defaultWorkspace.Config.Image.Reference, "image-ref", WorkspaceConfigImageRef, "Workspace image reference. Defaults to the official Polycrate image")
	fs.StringVar(&defaultWorkspace.Config.Image.Version, "image-version", version, "Workspace image version. Defaults to the version of Polycrate")
	fs.StringVar(&defaultWorkspace.Config.LogsRoot, "logs-root", WorkspaceConfigLogsRoot, "Logs root directory. Must be located inside the workspace.")
	fs.StringVar(&defaultWorkspace.Config.BlocksRoot, "blocks-root", WorkspaceConfigBlocksRoot, "Blocks root directory. Must be located inside the workspace.")
	fs.StringVar(&defaultWorkspace.Config.BlocksConfig, "blocks-config", BlocksConfigFile, "The name of the config file that holds a block's config.")
	fs.StringVar(&defaultWorkspace.Config.ArtifactsRoot, "artifacts-root", WorkspaceConfigArtifactsRoot, "Artifacts root directory. Must be located inside the workspace.")
	fs.StringVar(&defaultWorkspace.Config.WorkflowsRoot, "workflows-root", WorkspaceConfigWorkflowsRoot, "Workflows root directory. Must be located inside the workspace.")
	fs.StringVar(&defaultWorkspace.Config.Dockerfile, "dockerfile", WorkspaceConfigDockerfile, "The workspace Dockerfile. Can be used to permanently modify the workspace container and adjust it to your needs (e.g. install a .NET runtime, etc). polycrate builds the workspace container automatically when a Dockerfile is detected in the workspace.")
	fs.StringVar(&defaultWorkspace.Config.SshPrivateKey, "ssh-private-key", WorkspaceConfigSshPrivateKey, "Workspace ssh private key. This key can be used to connect to remote hosts via ssh. Must be located inside the workspace.")
	fs.StringVar(&defaultWorkspace.Config.SshPublicKey, "ssh-public-key", WorkspaceConfigSshPublicKey, "Workspace ssh public key. Add this key to your remote hosts' authorized_keys file Must be located inside the workspace..")
	fs.StringVar(&defaultWorkspace.Config.RemoteRoot, "remote-root", WorkspaceConfigRemoteRoot, "Remote root. This can be used as a common directory on remote hosts (e.g. a common directory to save Docker stacks and volumes to).")
	fs.StringVar(&defaultWorkspace.Config.ContainerRoot, "container-root", WorkspaceContainerRoot, "Workspace container root directory.")
}

func initConfig() {
	ctx := context.Background()

	if err := polycrate.Load(ctx); err != nil {
		log.Fatal(err)
	}

	if polycrate.Config.CheckUpdates {
		stableVersion, err := polycrate.GetStableVersion(ctx)
		if err != nil {
			log.Warn(err)
		} else {
			if stableVersion != version {
				log.Infof("There's a new version of Polycrate available: %s. Run `polycrate update` to update", stableVersion)
			}
		}

	}

	if version == "latest" {
		workspace.Config.Image.Version = "latest-amd64"
		log.Debug("Setting image version to latest-amd64 (development mode)")
	}

}

// func init() {
// 	rootCmd.AddCommand(kubectl.NewDefaultKubectlCommand())
// }
