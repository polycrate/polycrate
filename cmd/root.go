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
		//sync.History.cmd = cmd
		globalCmd = cmd

	},
	// PreRun: func(cmd *cobra.Command, args []string) {
	// 	ctx, cancelFunc := context.WithCancel(context.Background())
	// 	ctx, err := polycrate.StartTransaction(ctx, cancelFunc)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	cmd.SetContext(ctx)
	// },
	// PostRun: func(cmd *cobra.Command, args []string) {

	// },
	// PersistentPostRun: func(cmd *cobra.Command, args []string) {
	// 	workspace.SaveRevision().Flush()
	// 	workspace.Sync().Flush()

	// },
	Version: version,
}

func init() {
	cobra.OnInitialize(initConfig)

	fs := rootCmd.PersistentFlags()

	// Action related
	fs.BoolVarP(&local, "local", "l", false, "Run actions locally (without the polycrate container). Defaults to false.")
	fs.BoolVarP(&interactive, "interactive", "i", false, "Make the container interactive and accept input from stdin. Like '-it' for Docker.")
	fs.BoolVarP(&pull, "pull", "p", true, "Pull the workspace image before running the container. Defaults to true.")
	fs.BoolVarP(&build, "build", "b", true, "When this is true, a custom image will be built from the workspace Dockerfile. This image will then be used to run the action. Defaults to true.")

	// CLI related
	fs.StringVarP(&outputFormat, "output-format", "o", "yaml", "Output format (currently no-op).")
	fs.BoolVarP(&force, "force", "f", false, "Force whatever you want to do. Like sudo with more willpower. Defaults to false.")
	fs.BoolVarP(&snapshot, "snapshot", "", false, "Only dump the workspace snapshot, do not run anything.")
	fs.StringVar(&editor, "editor", DefaultEditor, "Editor to use to open the workspace")

	// Polycrate main config
	fs.StringVar(&polycrateConfigDir, "config-dir", polycrateConfigDir, "Path to the config directory")
	fs.StringVar(&polycrateConfigFilePath, "config-file", polycrateConfigFilePath, "Path to the config file")
	fs.StringVar(&polycrateWorkspaceDir, "workspace-dir", polycrateWorkspaceDir, "Path to the workspaces")
	fs.StringVar(&polycrateRuntimeDir, "runtime-dir", polycrateRuntimeDir, "Path to the runtime directory")

	// Workspace path/name
	fs.StringVarP(&defaultWorkspace.LocalPath, "workspace", "w", cwd, "The path to the workspace. Defaults to $PWD")

	// Common config
	fs.IntVar(&polycrate.Config.Loglevel, "loglevel", 1, "loglevel")
	fs.StringVar(&polycrate.Config.Logformat, "logformat", "default", "loglevel")
	fs.StringVar(&polycrate.Config.Registry.Url, "registry-url", RegistryUrl, "The URL of the OCI registry")
	fs.StringVar(&polycrate.Config.Registry.BaseImage, "registry-base-image", RegistryBaseImage, "The base image to package blocks in OCI format")

	// Workspace related config
	fs.StringVar(&defaultWorkspace.SyncOptions.Local.Branch.Name, "sync-local-branch", GitDefaultBranch, "Default git branch")
	fs.StringVar(&defaultWorkspace.SyncOptions.Remote.Branch.Name, "sync-remote-branch", GitDefaultBranch, "Default git branch")
	fs.StringVar(&defaultWorkspace.SyncOptions.Remote.Name, "sync-remote-name", GitDefaultRemote, "Default git remote")
	fs.BoolVar(&defaultWorkspace.SyncOptions.Enabled, "sync-enabled", false, "Sync enabled")
	fs.BoolVar(&defaultWorkspace.SyncOptions.Auto, "sync-auto", false, "Sync automatically")
	fs.StringSliceVarP(&defaultWorkspace.ExtraEnv, "env", "e", []string{}, "Additional environment variables for the workspace in the format 'KEY=value'")
	fs.StringSliceVarP(&defaultWorkspace.ExtraMounts, "mount", "m", []string{}, "Additional mounts for the workspace container in the format '/host:/container'. This will be ignored when used with --local")
	fs.StringVar(&defaultWorkspace.Config.WorkspaceConfig, "workspace-config", WorkspaceConfigFile, "The name of the config file that holds the workspace config.")
	fs.StringVar(&defaultWorkspace.Config.Image.Reference, "image-ref", WorkspaceConfigImageRef, "Workspace image reference. Defaults to the official Polycrate image")
	fs.StringVar(&defaultWorkspace.Config.Image.Version, "image-version", version, "Workspace image version. Defaults to the version of Polycrate")
	fs.StringVar(&defaultWorkspace.Config.BlocksRoot, "blocks-root", WorkspaceConfigBlocksRoot, "Blocks root directory. Must be located inside the workspace.")
	fs.StringVar(&defaultWorkspace.Config.BlocksConfig, "blocks-config", BlocksConfigFile, "The name of the config file that holds a block's config.")
	fs.StringVar(&defaultWorkspace.Config.ArtifactsRoot, "artifacts-root", WorkspaceConfigArtifactsRoot, "Artifacts root directory. Must be located inside the workspace.")
	fs.StringVar(&defaultWorkspace.Config.WorkflowsRoot, "workflows-root", WorkspaceConfigWorkflowsRoot, "Workflows root directory. Must be located inside the workspace.")
	fs.StringVar(&defaultWorkspace.Config.Dockerfile, "dockerfile", WorkspaceConfigDockerfile, "The workspace Dockerfile. Can be used to permanently modify the workspace container and adjust it to your needs (e.g. install a .NET runtime, etc). polycrate builds the workspace container automatically when a Dockerfile is detected in the workspace.")
	fs.StringVar(&defaultWorkspace.Config.SshPrivateKey, "ssh-private-key", WorkspaceConfigSshPrivateKey, "Workspace ssh private key. This key can be used to connect to remote hosts via ssh. Must be located inside the workspace.")
	fs.StringVar(&defaultWorkspace.Config.SshPublicKey, "ssh-public-key", WorkspaceConfigSshPublicKey, "Workspace ssh public key. Add this key to your remote hosts' authorized_keys file Must be located inside the workspace..")
	fs.StringVar(&defaultWorkspace.Config.RemoteRoot, "remote-root", WorkspaceConfigRemoteRoot, "Remote root. This can be used as a common directory on remote hosts (e.g. a common directory to save Docker stacks and volumes to).")
	fs.StringVar(&defaultWorkspace.Config.ContainerRoot, "container-root", WorkspaceContainerRoot, "Workspace container root directory.")

	// rootCmd.PersistentFlags().BoolVarP(&pull, "pull", "p", true, "Pull the workspace image before running the container. Defaults to true.")
	// rootCmd.PersistentFlags().BoolVarP(&local, "local", "l", false, "Run actions locally (without the polycrate container). Defaults to false.")
	// rootCmd.PersistentFlags().BoolVarP(&interactive, "interactive", "i", false, "Make the container interactive and accept input from stdin. Like '-it' for Docker.")

	// rootCmd.PersistentFlags().BoolVarP(&force, "force", "f", false, "Force whatever you want to do. Like sudo with more willpower. Defaults to false.")
	// rootCmd.PersistentFlags().IntVar(&logLevel, "loglevel", 1, "loglevel")
	// rootCmd.PersistentFlags().StringVarP(&outputFormat, "output-format", "o", "yaml", "Output format (currently no-op).")
	// rootCmd.PersistentFlags().BoolVarP(&build, "build", "b", true, "When this is true, a custom image will be built from the workspace Dockerfile. This image will then be used to run the action. Defaults to true.")
	// rootCmd.PersistentFlags().BoolVarP(&snapshot, "snapshot", "", false, "Only dump the workspace snapshot, do not run anything.")
	// rootCmd.PersistentFlags().StringVar(&editor, "editor", DefaultEditor, "Editor to use to open the workspace")

	// rootCmd.PersistentFlags().StringVar(&polycrate.Config.Gitlab.Url, "gitlab-url", GitLabDefaultUrl, "Default GitLab API endpoint")
	// rootCmd.PersistentFlags().StringVar(&polycrate.Config.Gitlab.Transport, "gitlab-transport", GitLabDefaultTransport, "Default GitLab repository action transport (ssh|http)")
	// rootCmd.PersistentFlags().StringVar(&polycrate.Config.Sync.DefaultBranch, "git-default-branch", GitDefaultBranch, "Default git branch")

	//rootCmd.PersistentFlags().StringSliceVarP(&workspace.overrides, "set", "s", []string{}, "Workspace ovrrides")

	// rootCmd.PersistentFlags().StringVar(&workspace.SyncOptions.Local.Branch.Name, "sync-local-branch", GitDefaultBranch, "Default git branch")
	// rootCmd.PersistentFlags().StringVar(&workspace.SyncOptions.Remote.Branch.Name, "sync-remote-branch", GitDefaultBranch, "Default git branch")
	// rootCmd.PersistentFlags().StringVar(&workspace.SyncOptions.Remote.Name, "sync-remote-name", GitDefaultRemote, "Default git remote")
	// rootCmd.PersistentFlags().BoolVar(&workspace.SyncOptions.Enabled, "sync-enabled", false, "Sync enabled")
	// rootCmd.PersistentFlags().BoolVar(&workspace.SyncOptions.Auto, "sync-auto", false, "Sync automatically")
	// rootCmd.PersistentFlags().StringVarP(&workspace.LocalPath, "workspace", "w", cwd, "The path to the workspace. Defaults to $PWD")
	// rootCmd.PersistentFlags().StringVar(&workspace.ContainerPath, "container-root", WorkspaceContainerRoot, "Workspace container root directory.")
	// rootCmd.PersistentFlags().StringVar(&workspace.Config.WorkspaceConfig, "workspace-config", WorkspaceConfigFile, "The name of the config file that holds the workspace config.")
	// rootCmd.PersistentFlags().StringVar(&workspace.Config.Image.Reference, "image-ref", WorkspaceConfigImageRef, "Workspace image reference. Defaults to the official Polycrate image")
	// rootCmd.PersistentFlags().StringVar(&workspace.Config.Image.Version, "image-version", version, "Workspace image version. Defaults to the version of Polycrate")
	// rootCmd.PersistentFlags().StringVar(&workspace.Config.BlocksRoot, "blocks-root", WorkspaceConfigBlocksRoot, "Blocks root directory. Must be located inside the workspace.")
	// rootCmd.PersistentFlags().StringVar(&workspace.Config.BlocksConfig, "blocks-config", BlocksConfigFile, "The name of the config file that holds a block's config.")
	// rootCmd.PersistentFlags().StringVar(&workspace.Config.ArtifactsRoot, "artifacts-root", WorkspaceConfigArtifactsRoot, "Artifacts root directory. Must be located inside the workspace.")
	// rootCmd.PersistentFlags().StringVar(&workspace.Config.WorkflowsRoot, "workflows-root", WorkspaceConfigWorkflowsRoot, "Workflows root directory. Must be located inside the workspace.")
	// rootCmd.PersistentFlags().StringVar(&workspace.Config.Dockerfile, "dockerfile", WorkspaceConfigDockerfile, "The workspace Dockerfile. Can be used to permanently modify the workspace container and adjust it to your needs (e.g. install a .NET runtime, etc). polycrate builds the workspace container automatically when a Dockerfile is detected in the workspace.")
	// rootCmd.PersistentFlags().StringVar(&workspace.Config.SshPrivateKey, "ssh-private-key", WorkspaceConfigSshPrivateKey, "Workspace ssh private key. This key can be used to connect to remote hosts via ssh. Must be located inside the workspace.")
	// rootCmd.PersistentFlags().StringVar(&workspace.Config.SshPublicKey, "ssh-public-key", WorkspaceConfigSshPublicKey, "Workspace ssh public key. Add this key to your remote hosts' authorized_keys file Must be located inside the workspace..")
	// rootCmd.PersistentFlags().StringVar(&workspace.Config.RemoteRoot, "remote-root", WorkspaceConfigRemoteRoot, "Remote root. This can be used as a common directory on remote hosts (e.g. a common directory to save Docker stacks and volumes to).")
	// rootCmd.PersistentFlags().StringVar(&polycrate.Config.Registry.Url, "registry-url", RegistryUrl, "The URL of the OCI registry")
	// rootCmd.PersistentFlags().StringVar(&polycrate.Config.Registry.BaseImage, "registry-base-image", RegistryBaseImage, "The base image to package blocks in OCI format")
	// rootCmd.PersistentFlags().StringSliceVarP(&workspace.ExtraEnv, "env", "e", []string{}, "Additional environment variables for the workspace in the format 'KEY=value'")
	// rootCmd.PersistentFlags().StringSliceVarP(&workspace.ExtraMounts, "mount", "m", []string{}, "Additional mounts for the workspace container in the format '/host:/container'. This will be ignored when used with --local")
	//rootCmd.PersistentFlags().StringVar(&config.Registry.BlockNamespace, "registry-block-namespace", RegistryBlockNamespace, "The Block namespace in the OCI registry")
	//rootCmd.PersistentFlags().StringVar(&polycrate.Config.Registry.ApiBase, "registry-api-base", RegistryApiBase, "The API base path of the Polycrate registry")
	// rootCmd.PersistentFlags().StringVar(&polycrate.Config.Registry.Username, "registry-username", "", "The username used to authenticate with the Polycrate registry")
	// rootCmd.PersistentFlags().StringVar(&polycrate.Config.Registry.Password, "registry-password", "", "The password used to authenticate with the Polycrate registry")

	// rootCmd.PersistentFlags().MarkDeprecated("registry-api-base", "Registry has moved to OCI compatible registry; API Base is deprecated.")
	// rootCmd.PersistentFlags().MarkDeprecated("registry-username", "Registry has moved to OCI compatible registry; Username is deprecated.")
	// rootCmd.PersistentFlags().MarkDeprecated("registry-password", "Registry has moved to OCI compatible registry; Password is deprecated.")
}

func initConfig() {
	ctx := context.Background()

	if err := polycrate.Load(ctx); err != nil {
		log.Fatal(err)
	}

	if version == "latest" {
		workspace.Config.Image.Version = "latest-amd64"
		log.Debug("Setting image version to latest-amd64 (development mode)")
	}
	// // Load Polycrate config
	// var polycrateConfig = viper.New()

	// // Match CLI Flags with Config options
	// // CLI Flags have precedence
	// polycrateConfig.BindPFlag("gitlab.url", rootCmd.Flags().Lookup("gitlab-url"))
	// polycrateConfig.BindPFlag("registry.url", rootCmd.Flags().Lookup("registry-url"))
	// //polycrateConfig.BindPFlag("registry.api_base", rootCmd.Flags().Lookup("registry-api-base"))
	// polycrateConfig.BindPFlag("registry.username", rootCmd.Flags().Lookup("registry-username"))
	// polycrateConfig.BindPFlag("registry.password", rootCmd.Flags().Lookup("registry-password"))
	// // workspaceConfig.BindPFlag("config.image.reference", rootCmd.Flags().Lookup("image-ref"))
	// // workspaceConfig.BindPFlag("config.blocksroot", rootCmd.Flags().Lookup("blocks-root"))
	// // workspaceConfig.BindPFlag("config.blocksconfig", rootCmd.Flags().Lookup("blocks-config"))
	// // workspaceConfig.BindPFlag("config.workflowsroot", rootCmd.Flags().Lookup("workflows-root"))
	// // workspaceConfig.BindPFlag("config.workspaceconfig", rootCmd.Flags().Lookup("workspace-config"))
	// // workspaceConfig.BindPFlag("config.artifactsroot", rootCmd.Flags().Lookup("artifacts-root"))
	// // workspaceConfig.BindPFlag("config.containerroot", rootCmd.Flags().Lookup("container-root"))
	// // workspaceConfig.BindPFlag("config.remoteroot", rootCmd.Flags().Lookup("remote-root"))
	// // workspaceConfig.BindPFlag("config.sshprivatekey", rootCmd.Flags().Lookup("ssh-private-key"))
	// // workspaceConfig.BindPFlag("config.sshpublickey", rootCmd.Flags().Lookup("ssh-public-key"))
	// // workspaceConfig.BindPFlag("config.dockerfile", rootCmd.Flags().Lookup("dockerfile"))

	// polycrateConfig.SetEnvPrefix(EnvPrefix)
	// polycrateConfig.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	// polycrateConfig.AutomaticEnv()

	// polycrateConfig.SetConfigType("yaml")
	// polycrateConfig.SetConfigFile(polycrateConfigFilePath)

	// if _, err := os.Stat(polycrateConfigFilePath); os.IsNotExist(err) {
	// 	// Seems config wasn't found
	// 	// Let's initialize it
	// 	CreateDir(polycrateHome)
	// 	CreateFile(polycrateConfigFilePath)
	// }
	// err = polycrateConfig.ReadInConfig()
	// if err != nil {
	// 	log.Fatalf(err.Error())
	// }

	// if err = polycrateConfig.Unmarshal(&polycrate.Config); err != nil {
	// 	log.Fatal(err)
	// }

	// if err = polycrate.Config.validate(); err != nil {
	// 	log.Fatal(err)
	// }

	// Goroutine to capture signals (SIGINT, etc)
	// Exits with exit code 1 when ctrl-c is captured
	//signals := make(chan os.Signal, 1)
	//done := make(chan bool, 1)

	// signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	// go func() {
	// 	s := <-signals

	// 	signalHandler(s)

	// }()

	//<-done
	//fmt.Println("exiting after SIGINT")

	// var logrusLogLevel string
	// switch polycrate.Config.Loglevel {
	// case 1:
	// 	logrusLogLevel = "Info"
	// case 2:
	// 	logrusLogLevel = "Debug"
	// case 3:
	// 	logrusLogLevel = "Trace"
	// default:
	// 	logrusLogLevel = "Info"
	// }

	// logrusLevel, err = log.ParseLevel(logrusLogLevel)
	// if err != nil {
	// 	logrusLevel = log.InfoLevel
	// }

	// // Set global log level
	// log.SetLevel(logrusLevel)

	// Set a different image if we're in development

	// Register the custom validators to the global validator variable
	// validate.RegisterValidation("metadata_name", validateMetadataName)
	// validate.RegisterValidation("block_name", validateBlockName)

	// // Discover local workspaces and load to localWorkspaceIndex
	// err = discoverWorkspaces()
	// if err != nil {
	// 	log.Fatal(err)
	// }

}
