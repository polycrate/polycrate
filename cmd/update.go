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
	"archive/tar"
	"context"
	"fmt"

	"compress/gzip"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var consent bool
var dryRun bool
var latestUrl string
var tempDownloadPath string
var packageRegistry string

// releaseCmd represents the release command
var updateCmd = &cobra.Command{
	Use:   "update [version] [--yes]",
	Args:  cobra.MaximumNArgs(1),
	Short: "Update Polycrate",
	Long: `
Update polycrate.

The first argument given to this command is used as the version you want to update to. By default, polycrate looks up the latest version from GitHub (https://github.com/polycrate/polycrate/releases/).

Use --yes/-y to automatically accept the consent promopt.

Use --force to re-install or downgrade to a specific version

`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		tx := polycrate.Transaction().SetCommand(cmd)
		defer tx.Stop()

		// err = polycrate.UpdateCLI(ctx)
		// if err != nil {
		// 	polycrate.ContextExit(ctx, cancelFunc, err)
		// }
		// log.Fatal("End update")

		stableVersion, err := polycrate.GetStableVersion(ctx)
		if err != nil {
			log.Fatal(err)
		}

		//stableVersion := getStableVersion()

		// WorkspaceConfigImageRef
		downloadVersion := stableVersion
		if len(args) == 1 {
			// args[0] is a concrete version the user can request
			// TODO: check this with semver upfront
			downloadVersion = args[0]
		}

		log.Warn("Current version: " + version)
		log.Warn("Current stable version: " + stableVersion)
		log.Warn("Requested version: " + downloadVersion)

		if downloadVersion == version && !force {
			log.Warn("Already up to date")
			os.Exit(0)
		}

		if !consent {
			updateConsentPrompt := promptContent{
				"Please select Yes or No",
				"Do you want to update from " + version + " to " + downloadVersion + "?",
			}

			updateConsent := promptYesNo(updateConsentPrompt)

			if updateConsent == "Yes" {
				consent = true
			}
		}

		if consent {
			log.Warn("You might be asked for your sudo password")
			err := downloadPolycrateCLI(downloadVersion)
			if err != nil {
				log.Fatal(err)
			}
			log.Warn("Update sucessful")
		} else {
			log.Warn("Update cancelled")
		}

	},
}

func init() {
	rootCmd.AddCommand(updateCmd)

	// https://github.com/polycrate/polycrate/releases/download/v0.2.2/polycrate_0.2.2_darwin_amd64.tar.gz

	updateCmd.PersistentFlags().BoolVarP(&consent, "yes", "y", false, "Consent to update")
	updateCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "d", false, "Don't actually do anything")
	updateCmd.PersistentFlags().StringVar(&latestUrl, "latest-url", "https://api.github.com/repos/polycrate/polycrate/releases/latest", "Latest URL")
	updateCmd.PersistentFlags().StringVar(&tempDownloadPath, "temp-download-path", "/tmp/polycrate", "Temporary download path")

}

type GitHubTag struct {
	Name       string `json:"name"`
	ZipBallUrl string `json:"zipball_url"`
	TarBallUrl string `json:"tarball_url"`
	Commit     struct {
		Sha string `json:"sha"`
		Url string `json:"url"`
	}
	NodeID string `json:"node_id"`
}

type CLIDownload struct {
	Version         string
	Arch            string
	Os              string
	PackageRegistry string
}

func ExtractTarGz(gzipStream io.Reader) error {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return err
	}

	tarReader := tar.NewReader(uncompressedStream)

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("ExtractTarGz: Next() failed: %s", err.Error())
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(header.Name, 0755); err != nil {
				log.Fatalf("ExtractTarGz: Mkdir() failed: %s", err.Error())
			}
		case tar.TypeReg:
			//outFile, err := os.Create(header.Name)
			outFile, err := os.OpenFile(header.Name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
			if err != nil {
				log.Fatalf("ExtractTarGz: Create() failed: %s", err.Error())
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				log.Fatalf("ExtractTarGz: Copy() failed: %s", err.Error())
			}
			outFile.Close()

		default:
			log.Fatalf(
				"ExtractTarGz: uknown type: %b in %s",
				header.Typeflag,
				header.Name)
		}

	}
	return nil
}

func Untar(tarball, target string) error {
	reader, err := os.Open(tarball)
	if err != nil {
		return err
	}
	defer reader.Close()
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		path := filepath.Join(target, header.Name)
		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(file, tarReader)
		if err != nil {
			return err
		}
	}
	return nil
}

func downloadPolycrateCLI(packageVersion string) error {
	// Discover arch/platform
	runtimeOS := runtime.GOOS
	runtimeArch := runtime.GOARCH
	downloadUrl := fmt.Sprintf("%s/get/polycrate/%s/%s_%s/polycrate_%s_%s_%s.tar.gz", polycrate.Config.Hub.Url, packageVersion, runtimeOS, runtimeArch, packageVersion, runtimeOS, runtimeArch)

	if !dryRun {
		// Create temp file
		packageDownload, err := os.CreateTemp("/tmp", "polycrate-"+packageVersion+"-*")
		CheckErr(err)

		// Download to tempfile
		err = DownloadFile(downloadUrl, packageDownload.Name())
		CheckErr(err)

		defer os.Remove(packageDownload.Name())

		log.Debug("Downloaded version " + packageVersion + " from " + downloadUrl + " to " + packageDownload.Name())

		// Unpack
		os.Chdir("/tmp")
		if _, err := os.Stat(tempDownloadPath); !os.IsNotExist(err) {
			err := os.Remove(tempDownloadPath)
			if err != nil {
				return err
			}
		}
		ExtractTarGz(packageDownload) // creates /tmp/polycrate
		if err != nil {
			return err
		}

		// Check existing CLI
		executable, err := os.Executable()
		CheckErr(err)

		if version == "development" || version == "latest" || version == "dev" {
			// in development, set executable to /usr/local/bin/cloudstack
			log.Debug("Development mode: Hard-wiring executable to /usr/local/bin/polycrate")
			executable = "/usr/local/bin/polycrate"
		}
		log.Debug("Current Executable: " + executable)

		// Check if executable exists
		if _, err := os.Stat(executable); os.IsNotExist(err) {
			// executable does not exist
			log.Debug("Executable not found at " + executable)
		} else {
			// executable exists
			executableBackup, err := ioutil.TempFile("/tmp", "polycrate-backup-*")
			CheckErr(err)
			defer os.Remove(executableBackup.Name())

			// backup executable to tempfile
			cmd := exec.Command("/bin/sh", "-c", "sudo mv "+executable+" "+executableBackup.Name())
			cmd.Run()
		}
		// move new package in place
		cmd := exec.Command("/bin/sh", "-c", "sudo mv "+tempDownloadPath+" "+executable)
		cmd.Run()
		cmd = exec.Command("/bin/sh", "-c", "sudo chmod +x "+executable)
		cmd.Run()
		executableSymlink := "/usr/local/bin/poly"
		cmd = exec.Command("/bin/sh", "-c", "sudo ln -s "+executable+" "+executableSymlink)
		cmd.Run()

		log.Info("Downloaded Polycrate version " + packageVersion + " to " + executable)
		// err = os.Rename(packageDownload.Name(), executable)
		// CheckErr(err)
	}

	return nil
}
