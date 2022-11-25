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
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var _sshBlock string = ""

// installCmd represents the install command
var sshCmd = &cobra.Command{
	Use:    "ssh",
	Short:  "SSH into a node",
	Hidden: true,
	Long:   ``,
	Args:   cobra.ExactArgs(1), // https://github.com/spf13/cobra/blob/master/user_guide.md
	Run: func(cmd *cobra.Command, args []string) {

		hostname := args[0]

		workspace.load().Flush()

		block := workspace.GetBlockFromIndex(_sshBlock)
		block.SSH(hostname).Flush()

		//workspace.RunAction(args[0]).Flush()
	},
}

func init() {
	rootCmd.AddCommand(sshCmd)

	sshCmd.PersistentFlags().StringVar(&_sshBlock, "block", "", "Block to load inventory from")
}

func connectWithSSH(node string) {
	var nodeInfo *viper.Viper
	var nodePath string

	// find node as is
	nodePath = "all.hosts." + node
	if !inventoryConfigObject.IsSet(nodePath) {
		// compile new node ID
		nodeLong := strings.Join([]string{workspace.Name, node}, "-")
		nodePath = "all.hosts." + nodeLong
		if !inventoryConfigObject.IsSet("all.hosts." + nodeLong) {
			log.Fatal("Node " + node + " and " + nodeLong + " not found")
		} else {
			log.Info("Found node " + nodeLong)
		}
	} else {
		log.Info("Found node " + node)
	}

	nodeInfo = inventoryConfigObject.Sub(nodePath)
	log.Debug(nodeInfo)

	// set ssh params
	var sshUser string
	if nodeInfo.IsSet("ansible_ssh_user") {
		sshUser = nodeInfo.GetString("ansible_ssh_user")
	} else {
		sshUser = "root"
	}

	var sshPort string
	if nodeInfo.IsSet("ansible_ssh_port") {
		sshPort = nodeInfo.GetString("ansible_ssh_port")
	} else {
		sshPort = "22"
	}

	var sshHost string
	if nodeInfo.IsSet("ansible_host") {
		sshHost = nodeInfo.GetString("ansible_host")
	} else {
		log.Fatal("ansible_host not set")
	}

	sshPrivateKey := filepath.Join(workspace.LocalPath, "id_rsa")

	args := []string{
		"-l",
		sshUser,
		"-i",
		sshPrivateKey,
		"-p",
		sshPort,
		sshHost,
	}

	log.Debug("ssh -l " + sshUser + " -i " + sshPrivateKey + " -p " + sshPort + " " + sshHost)

	log.Info("Starting ssh session: " + sshUser + "@" + sshHost + ":" + sshPort)

	_, err := RunCommand("ssh", args...)
	CheckErr(err)
}
