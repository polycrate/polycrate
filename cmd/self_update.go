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
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var packageRegistry string
var packageName string
var packageStableFile string
var consent bool
var dryRun bool

// releaseCmd represents the release command
var selfUpdateCmd = &cobra.Command{
	Use:   "self-update",
	Args:  cobra.MaximumNArgs(1),
	Short: "Update Cloudstack CLI",
	Long:  `Update Cloudstack CLI`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Running Self-Update")

		stableVersion := getStableVersion()
		downloadVersion := stableVersion
		if len(args) == 1 {
			// args[0] is a concrete version the user can request
			// TODO: check this with semver upfront
			downloadVersion = args[0]
		}

		log.Info("Current version: " + version)
		log.Info("Current stable version: " + stableVersion)
		log.Info("Requested version: " + downloadVersion)

		if downloadVersion == version && !force {
			log.Info("Already up to date")
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
			downloadCloudstackCLI(downloadVersion)
		}
	},
}

//
func init() {
	rootCmd.AddCommand(selfUpdateCmd)

	selfUpdateCmd.PersistentFlags().BoolVarP(&consent, "yes", "y", false, "Consent to update")
	selfUpdateCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "d", false, "Don't actually do anything")
	selfUpdateCmd.PersistentFlags().StringVar(&packageRegistry, "package-registry", "https://s3.ayedo.dev/packages", "Package Registry")
	selfUpdateCmd.PersistentFlags().StringVar(&packageName, "package-name", "cloudstack", "Package name")
	selfUpdateCmd.PersistentFlags().StringVar(&packageStableFile, "package-stable-file", "stable", "Stable file")
}

func getStableVersion() string {
	// Download stable file from package registry
	// Load from basic example
	u, err := url.Parse(packageRegistry)
	CheckErr(err)

	u.Path = path.Join(u.Path, packageName)
	u.Path = path.Join(u.Path, packageStableFile)

	s := u.String()

	// Get content of stable version file
	log.Debug("Determining stable version from " + s)
	_stableVersion, err := getRemoteFileContent(s)
	CheckErr(err)

	stableVersion := strings.Trim(_stableVersion, "\n")

	return stableVersion
}

func downloadCloudstackCLI(packageVersion string) error {
	// Discover arch/platform
	runtimeArch := runtime.GOARCH
	runtimeOS := runtime.GOOS

	// Craft download URL
	u, err := url.Parse(packageRegistry)
	CheckErr(err)

	// Add Package name
	u.Path = path.Join(u.Path, packageName)

	// Add Package version
	u.Path = path.Join(u.Path, packageVersion)

	// Add Package file
	packageFilename := packageName + "-" + runtimeOS + "-" + runtimeArch
	if runtimeOS == "windows" {
		packageFilename = packageFilename + ".exe"
	}
	u.Path = path.Join(u.Path, packageFilename)

	fileUrl := u.String()
	log.Info("Preparing to download Cloudstack CLI from " + fileUrl)
	log.Info("NOTE: You might be prompted for your sudo password")

	if !dryRun {
		// Create temp file
		packageDownload, err := ioutil.TempFile("/tmp", packageFilename+"-*")
		CheckErr(err)

		// Download to tempfile
		err = DownloadFile(fileUrl, packageDownload.Name())
		CheckErr(err)

		log.Debug("Downloaded version " + packageVersion + " to " + packageDownload.Name())

		// Check existing CLI
		executable, err := os.Executable()
		CheckErr(err)

		if version == "development" || version == "latest" {
			// in development, set executable to /usr/local/bin/cloudstack
			log.Debug("Development mode: Hard-wiring executable to /usr/local/bin/cloudstack")
			executable = "/usr/local/bin/cloudstack"
		}
		log.Debug("Current Executable: " + executable)

		// Check if executable exists
		if _, err := os.Stat(executable); os.IsNotExist(err) {
			// executable does not exist
			log.Warn("Executable not found at " + executable)
		} else {
			// executable exists
			executableBackup, err := ioutil.TempFile("/tmp", "cloudstack-backup-*")
			CheckErr(err)
			defer os.Remove(executableBackup.Name())

			// backup executable to tempfile
			cmd := exec.Command("/bin/sh", "-c", "sudo mv "+executable+" "+executableBackup.Name())
			cmd.Run()
		}
		// move new package in place
		cmd := exec.Command("/bin/sh", "-c", "sudo mv "+packageDownload.Name()+" "+executable)
		cmd.Run()
		cmd = exec.Command("/bin/sh", "-c", "sudo chmod +x "+executable)
		cmd.Run()

		log.Info("Downloaded Cloudstack CLI version " + packageVersion + " to " + executable)
		// err = os.Rename(packageDownload.Name(), executable)
		// CheckErr(err)
	}

	return nil
}
