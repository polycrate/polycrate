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
	"fmt"
	"os"
	"os/signal"

	//"strconv"

	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "polycrate",
	Short: "Polycrate ist ein Framework zum Entwickeln von Plattformen",
	Long: `Polycrate
	
Polycrate is a framework for platform development.
	
Learn more at https://docs.polycrate.io
	`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
	Version: version,
	// PersistentPreRun: func(cmd *cobra.Command, args []string) {
	// 	if err := loadStatefile(); err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	state.StartHistoryItem(cmd, "in progress")
	// },
	// PersistentPostRun: func(cmd *cobra.Command, args []string) {
	// 	currentHistoryItem.UpdateStatus(strconv.Itoa(pluginCallExitCode))

	// 	if err := state.WriteHistory(); err != nil {
	// 		log.Fatal(err)
	// 	}
	// },
}

func Execute() {
	CheckErr(rootCmd.Execute())
}

func init() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	go func() {
		<-signals
		signal.Stop(signals)
		fmt.Println()
		fmt.Println("CTRL-C command received. Exiting...")

		os.Exit(0)
	}()

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&logLevel, "loglevel", "0", "loglevel")

	rootCmd.PersistentFlags().BoolVarP(&pull, "pull", "p", true, "Pull the workspace image before running the container. Defaults to true.")

	rootCmd.PersistentFlags().BoolVarP(&local, "local", "l", false, "Run actions locally (without the polycrate container). Defaults to false.")

	rootCmd.PersistentFlags().BoolVarP(&force, "force", "f", false, "Force whatever you want to do. Like sudo with more willpower. Defaults to false.")

	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output-format", "o", "yaml", "Output format (currently no-op).")

	rootCmd.PersistentFlags().BoolVarP(&interactive, "interactive", "i", false, "Make the container interactive and accept input from stdin. Like '-it' for Docker.")

	rootCmd.PersistentFlags().BoolVarP(&build, "build", "b", true, "When this is true, a custom image will be built from the workspace Dockerfile. This image will then be used to run the action. Defaults to true.")

	rootCmd.PersistentFlags().BoolVarP(&snapshot, "snapshot", "", false, "Only dump the workspace snapshot, do not run anything.")

	// Workspace

	//rootCmd.PersistentFlags().StringSliceVarP(&workspace.overrides, "set", "s", []string{}, "Workspace ovrrides")
	rootCmd.PersistentFlags().StringVarP(&workspace.Path, "workspace", "w", cwd, "The path to the workspace. Defaults to $PWD")

	rootCmd.PersistentFlags().StringVar(&workspace.ContainerPath, "container-root", WorkspaceContainerRoot, "Workspace container root directory.")

	rootCmd.PersistentFlags().StringVar(&workspace.Config.WorkspaceConfig, "workspace-config", WorkspaceConfigFile, "The config file that holds the workspace config.")

	rootCmd.PersistentFlags().StringVar(&workspace.Config.Image.Reference, "image-ref", WorkspaceConfigImageRef, "Workspace image reference. Defaults to the official polycrate image")

	rootCmd.PersistentFlags().StringVar(&workspace.Config.Image.Version, "image-version", version, "Workspace image version. Defaults to the version of polycrate")

	rootCmd.PersistentFlags().StringVar(&workspace.Config.BlocksRoot, "blocks-root", WorkspaceConfigBlocksRoot, "Blocks root directory")

	rootCmd.PersistentFlags().StringVar(&workspace.Config.BlocksConfig, "blocks-config", BlocksConfigFile, "The config file that holds the block config.")

	rootCmd.PersistentFlags().StringVar(&workspace.Config.ArtifactsRoot, "artifacts-root", WorkspaceConfigArtifactsRoot, "Artifacts root directory")

	rootCmd.PersistentFlags().StringVar(&workspace.Config.WorkflowsRoot, "workflows-root", WorkspaceConfigWorkflowsRoot, "Workflows root directory")

	rootCmd.PersistentFlags().StringVar(&workspace.Config.Dockerfile, "dockerfile", WorkspaceConfigDockerfile, "The workspace Dockerfile. Can be used to permanently modify the workspace container and adjust it to your needs (e.g. install a .NET runtime, etc). polycrate builds the workspace container automatically when a Dockerfile is detected in the workspace.")

	rootCmd.PersistentFlags().StringVar(&workspace.Config.SshPrivateKey, "ssh-private-key", WorkspaceConfigSshPrivateKey, "Workspace ssh private key. This key can be used to connect to remote hosts via ssh.")

	rootCmd.PersistentFlags().StringVar(&workspace.Config.SshPublicKey, "ssh-public-key", WorkspaceConfigSshPublicKey, "Workspace ssh public key. Add this key to your remote hosts' authorized_keys file.")

	rootCmd.PersistentFlags().StringVar(&workspace.Config.RemoteRoot, "remote-root", WorkspaceConfigRemoteRoot, "Remote root. This can be used as a common directory on remote hosts (e.g. a common directory to save Docker stacks and volumes to).")

	rootCmd.PersistentFlags().StringSliceVarP(&workspace.ExtraEnv, "env", "e", []string{}, "Additional environment variables for the workspace in the format 'KEY=value'")

	rootCmd.PersistentFlags().StringSliceVarP(&workspace.ExtraMounts, "mount", "m", []string{}, "Additional mounts for the workspace container in the format '/host:/container'. This will be ignored when used with --local")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	var logrusLogLevel string
	switch logLevel {
	case "0":
		logrusLogLevel = "Warn"
	case "1":
		logrusLogLevel = "Info"
	case "2":
		logrusLogLevel = "Debug"
	default:
		logrusLogLevel = "Warn"
	}
	var err error
	logrusLevel, err = log.ParseLevel(logrusLogLevel)
	if err != nil {
		logrusLevel = log.InfoLevel
	}

	// set global log level
	log.SetLevel(logrusLevel)

	//validate.RegisterValidation("metadata_name", validateMetadataName)

	//log := log.WithFields(logrus.Fields{"workspace": workspace.Metadata.Name})

	if version == "development" {
		workspace.Config.Image.Version = "latest"
		log.Debug("Setting image version to latest (development mode)")
	}

	validate.RegisterValidation("metadata_name", validateMetadataName)

}
