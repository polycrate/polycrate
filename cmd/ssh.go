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
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
		_w := cmd.Flags().Lookup("workspace").Value.String()
		hostname := args[0]

		tx := polycrate.Transaction()
		tx.SetCommand(cmd)
		defer tx.Stop()

		workspace, err := polycrate.LoadWorkspace(tx, _w, true)
		if err != nil {
			tx.Log.Fatal(err)
		}

		if _sshBlock == "" {
			err := fmt.Errorf("no block selected. Use ' --block $BLOCK_NAME' to select an inventory source")
			log.Fatal(err)
		}

		var block *Block
		block, err = workspace.GetBlock(_sshBlock)
		if err != nil {
			log.Fatal(err)
		}

		if block != nil {
			err := block.SSH(tx, hostname, workspace)
			if err != nil {
				log.Error(err)
			}
		} else {
			err := fmt.Errorf("block does not exist: %s", _sshBlock)
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(sshCmd)

	sshCmd.PersistentFlags().StringVar(&_sshBlock, "block", "inventory", "Block to load inventory from")
}

func ConnectWithSSH(tx *PolycrateTransaction, username string, hostname string, port string, privateKey string) error {
	log := tx.Log.log
	log = log.WithField("user", username)
	log = log.WithField("port", port)
	log = log.WithField("host", hostname)
	log = log.WithField("privateKey", privateKey)

	log.Debugf("Starting ssh session")

	args := []string{
		"-l",
		username,
		"-i",
		privateKey,
		"-p",
		port,
		"-o",
		"StrictHostKeyChecking=no",
		"-o",
		"BatchMode=yes",
		hostname,
	}

	_, _, err := RunCommand(tx.Context, nil, "ssh", args...)
	if err != nil {
		return err
	}
	return nil
}

// func connectWithSSH(username string, hostname string, port string, privateKey string) error {
// 	server := strings.Join([]string{hostname, port}, ":")
// 	client, err := sshclient.DialWithKey(server, username, privateKey)
// 	if err != nil {
// 		return err
// 	}
// 	defer client.Close()

// 	log.WithFields(log.Fields{
// 		"workspace": workspace.Name,
// 		"user":      username,
// 		"host":      hostname,
// 		"port":      port,
// 	}).Debugf("Starting ssh session")

// 	if err := client.Terminal(nil).Start(); err != nil {
// 		return err
// 	}

// 	// pk, err := os.ReadFile(privateKey)
// 	// if err != nil {
// 	// 	return err
// 	// }
// 	// signer, err := ssh.ParsePrivateKey(pk)
// 	// if err != nil {
// 	// 	return err
// 	// }

// 	// conf := &ssh.ClientConfig{
// 	// 	User:            username,
// 	// 	HostKeyCallback: ssh.InsecureIgnoreHostKey(),
// 	// 	Auth: []ssh.AuthMethod{
// 	// 		ssh.PublicKeys(signer),
// 	// 	},
// 	// }
// 	// args := []string{
// 	// 	"-l",
// 	// 	username,
// 	// 	"-o",
// 	// 	"StrictHostKeyChecking=no",
// 	// 	"-o",
// 	// 	"BatchMode=yes",
// 	// 	"-i",
// 	// 	privateKey,
// 	// 	"-p",
// 	// 	port,
// 	// 	hostname,
// 	// }

// 	// var conn *ssh.Client

// 	// conn, err = ssh.Dial("tcp", strings.Join([]string{hostname, port}, ":"), conf)
// 	// if err != nil {
// 	// 	return err
// 	// }
// 	// defer conn.Close()

// 	// // Each ClientConn can support multiple interactive sessions,
// 	// // represented by a Session.
// 	// session, err := conn.NewSession()
// 	// if err != nil {
// 	// 	panic("Failed to create session: " + err.Error())
// 	// }
// 	// defer session.Close()

// 	// // Set IO
// 	// session.Stdout = os.Stdout
// 	// session.Stderr = os.Stderr
// 	// in, _ := session.StdinPipe()

// 	// // Set up terminal modes
// 	// modes := ssh.TerminalModes{
// 	// 	ssh.ECHO:          0,     // disable echoing
// 	// 	ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
// 	// 	ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
// 	// }

// 	// // Request pseudo terminal
// 	// if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
// 	// 	log.Fatalf("request for pseudo terminal failed: %s", err)
// 	// }

// 	// // Start remote shell
// 	// if err := session.Shell(); err != nil {
// 	// 	log.Fatalf("failed to start shell: %s", err)
// 	// }

// 	// // Accepting commands
// 	// for {
// 	// 	reader := bufio.NewReader(os.Stdin)
// 	// 	str, _ := reader.ReadString('\n')
// 	// 	fmt.Fprint(in, str)
// 	// }

// 	return err
// }
