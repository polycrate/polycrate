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
	"encoding/json"
	goErrors "errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jedib0t/go-pretty/v6/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/xeonx/timeago"
	"gopkg.in/yaml.v2"
)

var activeFlags []map[string]string
var activeParents []string

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show history",
	Long:  `Show history`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			// Show specific history item
			itemNumber, _ := strconv.Atoi(args[0])

			if len(state.History) < itemNumber {
				log.Fatal("History item " + strconv.Itoa(itemNumber) + " does not exist")
			}
			stateHistoryItem := state.History[itemNumber]
			stateHistoryItem.Show()
		} else {
			if len(state.History) < 1 {
				log.Fatal("History is empty")
			}
			state.showHistory()
		}
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		loadWorkspace()
	},
}

func init() {
	rootCmd.AddCommand(historyCmd)
}

type PluginState struct {
	Name    string `mapstructure:"name,omitempty" json:"name,omitempty"`
	Command string `mapstructure:"command,omitempty" json:"command,omitempty"`
	Version string `mapstructure:"version,omitempty" json:"version,omitempty"`
	//Config  PluginConfig `mapstructure:"config,omitempty" json:"config,omitempty"`
}

type Environment struct {
	Arch     string `mapstructure:"arch,omitempty" json:"arch,omitempty"`
	Os       string `mapstructure:"os,omitempty" json:"os,omitempty"`
	Hostname string `mapstructure:"hostname,omitempty" json:"hostname,omitempty"`
	User     string `mapstructure:"user,omitempty" json:"user,omitempty"`
}

type StateHistoryItem struct {
	Date        string              `mapstructure:"date,omitempty" json:"date,omitempty" validate:"datetime=2006-01-02T15:04:05-0700"`
	Command     string              `mapstructure:"command,omitempty" json:"command,omitempty"`
	Flags       []map[string]string `mapstructure:"flags,omitempty" json:"flags,omitempty"`
	Args        []string            `mapstructure:"args,omitempty" json:"args,omitempty"`
	Version     string              `mapstructure:"version,omitempty" json:"version,omitempty"`
	Status      string              `mapstructure:"status,omitempty" json:"status,omitempty"`
	Environment Environment         `mapstructure:"environment,omitempty" json:"environment,omitempty"`
}

type Statefile struct {
	History []StateHistoryItem `mapstructure:"history" json:"history" validate:"dive,required"`
	//Plugins map[string]PluginConfig `mapstructure:"plugins,omitempty" json:"plugins,omitempty"`
	Changed string `mapstructure:"changed" json:"changed" validate:"required"`
}

func showHistory() {
	fmt.Println("Stack Directory: ", workspaceDir)
}

func checkFlags(f *pflag.Flag) {
	activeFlags = append(activeFlags, map[string]string{
		"name":  f.Name,
		"value": f.Value.String(),
	})
}
func checkParents(cmd *cobra.Command) {
	log.Debug("Adding active parent " + cmd.Name())
	activeParents = append(activeParents, cmd.Name())
}

func (c *Statefile) StartHistoryItem(cmd *cobra.Command, status string) {
	runtimeArch := runtime.GOARCH
	runtimeOs := runtime.GOOS
	hostname, _ := os.Hostname()

	_runtimeUser, _ := user.Current()
	runtimeUser := _runtimeUser.Username

	// Load flags into activeFlags variable
	cmd.Flags().Visit(checkFlags)

	// Find parents
	cmd.VisitParents(checkParents)

	// Reverse parents
	for i, j := 0, len(activeParents)-1; i < j; i, j = i+1, j-1 {
		activeParents[i], activeParents[j] = activeParents[j], activeParents[i]
	}

	fullCommand := cmd.CalledAs()
	if len(activeParents) > 0 {
		parentCommand := strings.Join(activeParents, " ")
		fullCommand = strings.Join([]string{parentCommand, cmd.CalledAs()}, " ")
	}

	environment := Environment{
		Arch:     runtimeArch,
		Os:       runtimeOs,
		Hostname: hostname,
		User:     runtimeUser,
	}

	currentHistoryItem = StateHistoryItem{
		Date:        now,
		Version:     version,
		Status:      "in progress",
		Command:     fullCommand,
		Flags:       activeFlags,
		Args:        cmd.Flags().Args(),
		Environment: environment,
	}
}

func addHistoryItem(cmd *cobra.Command, status string) {
	runtimeArch := runtime.GOARCH
	runtimeOs := runtime.GOOS
	hostname, _ := os.Hostname()

	_runtimeUser, _ := user.Current()
	runtimeUser := _runtimeUser.Username

	// Load flags into activeFlags variable
	cmd.Flags().Visit(checkFlags)

	// Find parents
	cmd.VisitParents(checkParents)

	// Reverse parents
	for i, j := 0, len(activeParents)-1; i < j; i, j = i+1, j-1 {
		activeParents[i], activeParents[j] = activeParents[j], activeParents[i]
	}

	fullCommand := cmd.CalledAs()
	if len(activeParents) > 0 {
		parentCommand := strings.Join(activeParents, " ")
		fullCommand = strings.Join([]string{parentCommand, cmd.CalledAs()}, " ")
	}

	environment := Environment{
		Arch:     runtimeArch,
		Os:       runtimeOs,
		Hostname: hostname,
		User:     runtimeUser,
	}

	currentHistoryItem = StateHistoryItem{
		Date:        now,
		Version:     version,
		Status:      "in progress",
		Command:     fullCommand,
		Flags:       activeFlags,
		Args:        cmd.Flags().Args(),
		Environment: environment,
	}
}

func (c *StateHistoryItem) UpdateStatus(status string) {
	c.Status = status
}

func (c *Statefile) WriteHistory() error {
	if currentHistoryItem.Date != "" {
		c.History = append(c.History, currentHistoryItem)
		c.Changed = now
		err := c.Save()
		return err
	} else {
		log.Warn("No history item found, not writing history today :(")
	}
	return nil
}

func writeHistory() error {
	if currentHistoryItem.Date != "" {
		state.History = append(state.History, currentHistoryItem)
		state.Changed = now
		err := state.Save()
		return err
	} else {
		log.Warn("No history item found, not writing history today :(")
	}
	return nil
}

func (c *Statefile) Save() error {
	log.Debug("Saving Statefile: " + statefile)

	fileData, _ := yaml.Marshal(c)

	err := ioutil.WriteFile(statefile, fileData, 0644)

	return err
}

func (c *Statefile) showHistory() {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"", "Long Date", "When", "Who", "Version", "Command", "Args", "Flags", "Plugin Version", "Status"})

	for historyItem := range c.History {
		longDate := c.History[historyItem].Date
		layout := "2006-01-02T15:04:05-0700"
		dateObject, _ := time.Parse(layout, longDate)
		shortDate := timeago.English.Format(dateObject)

		// Compose flags
		flagList := []string{}
		for _, v := range c.History[historyItem].Flags {
			flag := strings.Join([]string{"--" + v["name"], v["value"]}, " ")
			flagList = append(flagList, flag)
		}
		t.AppendRow([]interface{}{
			historyItem,
			longDate,
			shortDate,
			c.History[historyItem].Environment.User,
			c.History[historyItem].Version,
			c.History[historyItem].Command,
			strings.Join(c.History[historyItem].Args, " "),
			strings.Join(flagList, " "),
			c.History[historyItem].Status})
	}

	t.AppendFooter(table.Row{"", "", "", "", "", "", "", "", "Total", len(c.History)})

	t.SortBy([]table.SortBy{
		{Name: "Long Date", Mode: table.Dsc},
	})

	t.SetColumnConfigs([]table.ColumnConfig{
		{
			Name:   "Long Date",
			Hidden: true,
		},
	})

	t.Render()
}

func (c *StateHistoryItem) Show() {
	showStateHistoryItem(c, outputFormat)
}

func showStateHistoryItem(stateHistoryitem *StateHistoryItem, format string) {
	if format == "json" {
		data, err := json.Marshal(stateHistoryitem)
		CheckErr(err)
		fmt.Printf("%s\n", data)
	}
	if format == "yaml" {
		data, err := yaml.Marshal(stateHistoryitem)
		CheckErr(err)
		fmt.Printf("%s\n", data)
	}
}

func (c *Statefile) Validate() error {
	validate := validator.New()

	err := validate.Struct(c)

	if err != nil {

		// this check is only needed when your code could produce
		// an invalid value for validation such as interface with nil
		// value most including myself do not usually have code like this.
		if _, ok := err.(*validator.InvalidValidationError); ok {
			log.Error(err)
			return nil
		}

		for _, err := range err.(validator.ValidationErrors) {
			log.Error("Configuration option `" + strings.ToLower(err.Namespace()) + "` failed to validate: " + err.Tag())
		}

		// from here you can create your own error messages in whatever language you wish
		return goErrors.New("Error validating Statefile")
	}
	return nil
}
