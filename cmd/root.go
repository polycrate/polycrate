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

	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

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
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		sync.History.cmd = cmd

	},
	Version: version,
}

func Execute() {
	CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&logLevel, "loglevel", "0", "loglevel")
	rootCmd.PersistentFlags().BoolVarP(&pull, "pull", "p", true, "Pull the workspace image before running the container. Defaults to true.")
	rootCmd.PersistentFlags().BoolVarP(&local, "local", "l", false, "Run actions locally (without the polycrate container). Defaults to false.")
	rootCmd.PersistentFlags().BoolVarP(&force, "force", "f", false, "Force whatever you want to do. Like sudo with more willpower. Defaults to false.")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output-format", "o", "yaml", "Output format (currently no-op).")
	rootCmd.PersistentFlags().BoolVarP(&interactive, "interactive", "i", false, "Make the container interactive and accept input from stdin. Like '-it' for Docker.")
	rootCmd.PersistentFlags().BoolVarP(&build, "build", "b", true, "When this is true, a custom image will be built from the workspace Dockerfile. This image will then be used to run the action. Defaults to true.")
	rootCmd.PersistentFlags().BoolVarP(&snapshot, "snapshot", "", false, "Only dump the workspace snapshot, do not run anything.")
	rootCmd.PersistentFlags().StringVar(&editor, "editor", DefaultEditor, "Editor to use to open the workspace")

	rootCmd.PersistentFlags().StringVar(&config.Gitlab.Url, "gitlab-url", GitLabDefaultUrl, "Default GitLab API endpoint")
	rootCmd.PersistentFlags().StringVar(&config.Gitlab.Transport, "gitlab-transport", GitLabDefaultTransport, "Default GitLab repository action transport (ssh|http)")
	rootCmd.PersistentFlags().StringVar(&config.Sync.DefaultBranch, "git-default-branch", GitDefaultBranch, "Default git branch")
	rootCmd.PersistentFlags().StringVar(&workspace.SyncOptions.Local.Branch.Name, "sync-local-branch", GitDefaultBranch, "Default git branch")
	rootCmd.PersistentFlags().StringVar(&workspace.SyncOptions.Remote.Branch.Name, "sync-remote-branch", GitDefaultBranch, "Default git branch")
	rootCmd.PersistentFlags().StringVar(&workspace.SyncOptions.Remote.Name, "sync-remote-name", GitDefaultRemote, "Default git remote")
	rootCmd.PersistentFlags().BoolVar(&workspace.SyncOptions.Enabled, "sync-enabled", false, "Sync enabled")
	rootCmd.PersistentFlags().BoolVar(&workspace.SyncOptions.Auto, "sync-auto", false, "Sync automatically")

	//rootCmd.PersistentFlags().StringSliceVarP(&workspace.overrides, "set", "s", []string{}, "Workspace ovrrides")
	rootCmd.PersistentFlags().StringVarP(&workspace.LocalPath, "workspace", "w", cwd, "The path to the workspace. Defaults to $PWD")
	rootCmd.PersistentFlags().StringVar(&workspace.ContainerPath, "container-root", WorkspaceContainerRoot, "Workspace container root directory.")
	rootCmd.PersistentFlags().StringVar(&workspace.Config.WorkspaceConfig, "workspace-config", WorkspaceConfigFile, "The name of the config file that holds the workspace config.")
	rootCmd.PersistentFlags().StringVar(&workspace.Config.Image.Reference, "image-ref", WorkspaceConfigImageRef, "Workspace image reference. Defaults to the official Polycrate image")
	rootCmd.PersistentFlags().StringVar(&workspace.Config.Image.Version, "image-version", version, "Workspace image version. Defaults to the version of Polycrate")
	rootCmd.PersistentFlags().StringVar(&workspace.Config.BlocksRoot, "blocks-root", WorkspaceConfigBlocksRoot, "Blocks root directory. Must be located inside the workspace.")
	rootCmd.PersistentFlags().StringVar(&workspace.Config.BlocksConfig, "blocks-config", BlocksConfigFile, "The name of the config file that holds a block's config.")
	rootCmd.PersistentFlags().StringVar(&workspace.Config.ArtifactsRoot, "artifacts-root", WorkspaceConfigArtifactsRoot, "Artifacts root directory. Must be located inside the workspace.")
	rootCmd.PersistentFlags().StringVar(&workspace.Config.WorkflowsRoot, "workflows-root", WorkspaceConfigWorkflowsRoot, "Workflows root directory. Must be located inside the workspace.")
	rootCmd.PersistentFlags().StringVar(&workspace.Config.Dockerfile, "dockerfile", WorkspaceConfigDockerfile, "The workspace Dockerfile. Can be used to permanently modify the workspace container and adjust it to your needs (e.g. install a .NET runtime, etc). polycrate builds the workspace container automatically when a Dockerfile is detected in the workspace.")
	rootCmd.PersistentFlags().StringVar(&workspace.Config.SshPrivateKey, "ssh-private-key", WorkspaceConfigSshPrivateKey, "Workspace ssh private key. This key can be used to connect to remote hosts via ssh. Must be located inside the workspace.")
	rootCmd.PersistentFlags().StringVar(&workspace.Config.SshPublicKey, "ssh-public-key", WorkspaceConfigSshPublicKey, "Workspace ssh public key. Add this key to your remote hosts' authorized_keys file Must be located inside the workspace..")
	rootCmd.PersistentFlags().StringVar(&workspace.Config.RemoteRoot, "remote-root", WorkspaceConfigRemoteRoot, "Remote root. This can be used as a common directory on remote hosts (e.g. a common directory to save Docker stacks and volumes to).")
	rootCmd.PersistentFlags().StringSliceVarP(&workspace.ExtraEnv, "env", "e", []string{}, "Additional environment variables for the workspace in the format 'KEY=value'")
	rootCmd.PersistentFlags().StringSliceVarP(&workspace.ExtraMounts, "mount", "m", []string{}, "Additional mounts for the workspace container in the format '/host:/container'. This will be ignored when used with --local")

	rootCmd.PersistentFlags().StringVar(&config.Registry.Url, "registry-url", RegistryUrl, "The URL of the OCI registry")
	rootCmd.PersistentFlags().StringVar(&config.Registry.BlockNamespace, "registry-block-namespace", RegistryBlockNamespace, "The Block namespace in the OCI registry")
	rootCmd.PersistentFlags().StringVar(&config.Registry.ApiBase, "registry-api-base", RegistryApiBase, "The API base path of the Polycrate registry")
	rootCmd.PersistentFlags().StringVar(&config.Registry.BaseImage, "registry-base-image", RegistryBaseImage, "The base image to package blocks in OCI format")
	rootCmd.PersistentFlags().StringVar(&config.Registry.Username, "registry-username", "", "The username used to authenticate with the Polycrate registry")
	rootCmd.PersistentFlags().StringVar(&config.Registry.Password, "registry-password", "", "The password used to authenticate with the Polycrate registry")
}

func initConfig() {
	// Load Polycrate config
	var polycrateConfig = viper.New()

	// Match CLI Flags with Config options
	// CLI Flags have precedence
	polycrateConfig.BindPFlag("gitlab.url", rootCmd.Flags().Lookup("gitlab-url"))
	polycrateConfig.BindPFlag("registry.url", rootCmd.Flags().Lookup("registry-url"))
	polycrateConfig.BindPFlag("registry.api_base", rootCmd.Flags().Lookup("registry-api-base"))
	polycrateConfig.BindPFlag("registry.username", rootCmd.Flags().Lookup("registry-username"))
	polycrateConfig.BindPFlag("registry.password", rootCmd.Flags().Lookup("registry-password"))
	// workspaceConfig.BindPFlag("config.image.reference", rootCmd.Flags().Lookup("image-ref"))
	// workspaceConfig.BindPFlag("config.blocksroot", rootCmd.Flags().Lookup("blocks-root"))
	// workspaceConfig.BindPFlag("config.blocksconfig", rootCmd.Flags().Lookup("blocks-config"))
	// workspaceConfig.BindPFlag("config.workflowsroot", rootCmd.Flags().Lookup("workflows-root"))
	// workspaceConfig.BindPFlag("config.workspaceconfig", rootCmd.Flags().Lookup("workspace-config"))
	// workspaceConfig.BindPFlag("config.artifactsroot", rootCmd.Flags().Lookup("artifacts-root"))
	// workspaceConfig.BindPFlag("config.containerroot", rootCmd.Flags().Lookup("container-root"))
	// workspaceConfig.BindPFlag("config.remoteroot", rootCmd.Flags().Lookup("remote-root"))
	// workspaceConfig.BindPFlag("config.sshprivatekey", rootCmd.Flags().Lookup("ssh-private-key"))
	// workspaceConfig.BindPFlag("config.sshpublickey", rootCmd.Flags().Lookup("ssh-public-key"))
	// workspaceConfig.BindPFlag("config.dockerfile", rootCmd.Flags().Lookup("dockerfile"))

	polycrateConfig.SetEnvPrefix(EnvPrefix)
	polycrateConfig.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	polycrateConfig.AutomaticEnv()

	polycrateConfig.SetConfigType("yaml")
	polycrateConfig.SetConfigFile(polycrateConfigFilePath)

	if _, err := os.Stat(polycrateConfigFilePath); os.IsNotExist(err) {
		// Seems config wasn't found
		// Let's initialize it
		CreateDir(polycrateHome)
		CreateFile(polycrateConfigFilePath)
	}
	err := polycrateConfig.ReadInConfig()
	if err != nil {
		log.Fatalf(err.Error())
	}

	if err = polycrateConfig.Unmarshal(&config); err != nil {
		log.Fatal(err)
	}

	if err = config.validate(); err != nil {
		log.Fatal(err)
	}

	// Goroutine to capture signals (SIGINT, etc)
	// Exits with exit code 1 when ctrl-c is captured
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	go func() {
		<-signals
		//signal.Stop(signals)
		fmt.Println()
		// Deal with running containers
		cleanupWorkspace()

		log.Fatalf("ctrl-c received")
	}()

	var logrusLogLevel string
	switch logLevel {
	case "0":
		logrusLogLevel = "Info"
	case "1":
		logrusLogLevel = "Debug"
	case "2":
		logrusLogLevel = "Trace"
	default:
		logrusLogLevel = "Info"
	}

	logrusLevel, err = log.ParseLevel(logrusLogLevel)
	if err != nil {
		logrusLevel = log.InfoLevel
	}

	// Set global log level
	log.SetLevel(logrusLevel)

	// Set a different image if we're in development
	if version == "development" {
		workspace.Config.Image.Version = "latest"
		log.Debug("Setting image version to latest (development mode)")
	}

	// Register the custom validators to the global validator variable
	validate.RegisterValidation("metadata_name", validateMetadataName)

	// Discover local workspaces
	err = discoverWorkspaces()
	if err != nil {
		log.Fatal(err)
	}
}
