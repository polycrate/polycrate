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

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "polycrate",
	Short: "Polycrate ist ein Framework zum Entwickeln von Plattformen",
	Long: `Polycrate
	
Polycrate ist ein Framework zum Entwickeln von Plattformen.
	
Erfahre mehr unter https://accelerator.ayedo.de/polycrate
	`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
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

		currentHistoryItem.UpdateStatus("cancelled")
		state.WriteHistory()

		os.Exit(0)
	}()

	cobra.OnInitialize(initConfig)

	//rootCmd.PersistentFlags().StringVar(&command, "command", "", "command")
	//rootCmd.PersistentFlags().StringVar(&statefile, "statefile", "", "Polycrate state")
	//rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", filepath.Join(home, ".kube/config"), "kubeconfig")
	//rootCmd.PersistentFlags().StringVar(&plugins, "plugins", "", "plugins")
	//rootCmd.PersistentFlags().StringVar(&workspaceConfig, "workspace-config", "", "Polycrate Workspace Config")
	rootCmd.PersistentFlags().StringVar(&logLevel, "loglevel", "0", "loglevel")

	rootCmd.PersistentFlags().BoolVarP(&pull, "pull", "p", true, "Pull images upfront")

	rootCmd.PersistentFlags().BoolVarP(&local, "local", "l", false, "Run commands locally instead of the container")

	rootCmd.PersistentFlags().BoolVarP(&force, "force", "f", false, "Force execution")

	rootCmd.PersistentFlags().StringSliceVarP(&overrides, "set", "s", []string{}, "Development directory for this component")

	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output-format", "o", "yaml", "Output format")

	rootCmd.PersistentFlags().BoolVarP(&interactive, "interactive", "i", false, "Interactive container session")
	rootCmd.PersistentFlags().StringVarP(&workspaceDir, "workspace", "w", cwd, "Polycrate Workspace directory")

	rootCmd.PersistentFlags().StringVar(&imageRef, "image-ref", "ghcr.io/polycrate/polycrate", "image reference")

	rootCmd.PersistentFlags().StringVar(&imageVersion, "image-version", version, "image version")

	rootCmd.PersistentFlags().StringVar(&blocksRoot, "blocks-root", "blocks", "Blocks root directory")

	rootCmd.PersistentFlags().StringVar(&workflowsRoot, "workflows-root", "workflows", "Workflows root directory")

	rootCmd.PersistentFlags().StringVar(&workflowsRoot, "state-root", "state", "State root directory")

	rootCmd.PersistentFlags().StringVar(&workspaceContainerDir, "container-root", "/workspace", "Workspace container root directory")

	rootCmd.PersistentFlags().StringVar(&sshPrivateKey, "ssh-private-key", "id_rsa", "Workspace ssh private key")

	rootCmd.PersistentFlags().StringVar(&sshPublicKey, "ssh-public-key", "id_rsa.pub", "Workspace ssh public key")

	rootCmd.PersistentFlags().StringVar(&remoteRoot, "remote-root", "/polycrate", "Remote root")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	var logrusLogLevel string
	switch logLevel {
	case "0":
		logrusLogLevel = "Info"
	case "1":
		logrusLogLevel = "Debug"
	case "2":
		logrusLogLevel = "Trace"
	default:
		logrusLogLevel = "Warn"
	}
	ll, err := logrus.ParseLevel(logrusLogLevel)
	if err != nil {
		ll = logrus.InfoLevel
	}
	// set global log level
	log.SetLevel(ll)

	//discoverKubeconfig()

	// // Additionally, check for KUBECONFIG env var
	// kubeconfigEnv := os.Getenv("KUBECONFIG")

	// if kubeconfigEnv != "" {
	// 	kubeconfig = kubeconfigEnv
	// 	log.Debug("Setting kubeconfig from KUBECONFIG env var to ", kubeconfigEnv)
	// }

	// log.Debug("Trying to find a kubeconfig in ", context)

	// // Get stack kubeconfig
	// stackKubeConfigPath := getKubeconfigPath(context)

	// // Overwrite config if kubeconfig has been found in stack_dir
	// if stackKubeConfigPath != "" {
	// 	//viper.Set("kubeconfig", stackKubeConfigPath)
	// 	kubeconfig = stackKubeConfigPath
	// 	log.Debug("Setting kubeconfig to ", stackKubeConfigPath)
	// }

	// Check if dev-dir is set
	// and set cloudstackVersion to "latest"
	polycrateVersion = version

	if imageVersion == "development" {
		imageVersion = "latest"
		log.Debug("Setting image version to latest")
	}

	// Load Workspace
	//loadWorkspace()
	//_loadWorkspace()
}
