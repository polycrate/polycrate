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
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"golang.org/x/crypto/ssh"
)

// releaseCmd represents the release command
var CreateSSHKeyCmd = &cobra.Command{
	Use:   "ssh-keys",
	Args:  cobra.ExactArgs(0),
	Short: "Generate SSH Keys for a stack",
	Long:  `Generate SSH Keys for a stack`,
	Run: func(cmd *cobra.Command, args []string) {
		loadWorkspace()
		err := CreateSSHKeys()
		//CheckErr(err)
		if err != nil {
			log.Warn(err)
		}
	},
}

func init() {
	createCmd.AddCommand(CreateSSHKeyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// releaseCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// releaseCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func marshalRSAPrivate(priv *rsa.PrivateKey) string {
	return string(pem.EncodeToMemory(&pem.Block{
		Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv),
	}))
}

func generateKey() (string, string, error) {
	reader := rand.Reader
	bitSize := 2048

	key, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		return "", "", err
	}

	pub, err := ssh.NewPublicKey(key.Public())
	if err != nil {
		return "", "", err
	}
	pubKeyStr := string(ssh.MarshalAuthorizedKey(pub))
	privKeyStr := marshalRSAPrivate(key)

	return pubKeyStr, privKeyStr, nil
}

func CreateSSHKeys() error {

	privKeyPath := filepath.Join(workspace.path, workspace.Config.SshPrivateKey)
	pubKeyPath := filepath.Join(workspace.path, workspace.Config.SshPublicKey)

	log.Debug("Asserting private ssh key at ", privKeyPath)
	log.Debug("Asserting public ssh key at ", pubKeyPath)

	_, privKeyErr := os.Stat(privKeyPath)
	_, pubKeyErr := os.Stat(pubKeyPath)

	// Check if keys do already exist
	if os.IsNotExist(privKeyErr) && os.IsNotExist(pubKeyErr) {
		// No keys found
		// Generate new ones
		pubKeyStr, privKeyStr, err := generateKey()
		CheckErr(err)

		// Save private key
		privKeyFile, err := os.Create(privKeyPath)
		CheckErr(err)

		defer privKeyFile.Close()

		_, errPrivKey := privKeyFile.WriteString(privKeyStr)
		CheckErr(errPrivKey)

		err = os.Chmod(privKeyPath, 0600)
		CheckErr(err)

		// Save public key
		pubKeyFile, err := os.Create(pubKeyPath)
		CheckErr(err)

		defer pubKeyFile.Close()

		_, errPubKey := pubKeyFile.WriteString(pubKeyStr)
		CheckErr(errPubKey)

		err = os.Chmod(pubKeyPath, 0644)
		CheckErr(err)
	} else {
		return errors.New("SSH Keys already exist")
	}

	return nil
}
