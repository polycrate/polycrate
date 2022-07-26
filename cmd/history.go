package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var logCmd = &cobra.Command{
	// Deprecated: "This command is deprecated as it has limited functionality. Use `cloudstack pipelines` instead",
	Use:   "log 'this is a logmessage'",
	Short: "Adds a message to the log file",
	Args:  cobra.ExactArgs(1),
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		message := args[0]
		workspace.load().Flush()

		sync.Log(message)
	},
}

func init() {
	rootCmd.AddCommand(logCmd)
}

type HistoryLog struct {
	Msg         string    `yaml:"msg,omitempty" mapstructure:"msg,omitempty" json:"msg,omitempty"`
	Command     string    `yaml:"command,omitempty" mapstructure:"command,omitempty" json:"command,omitempty"`
	ExitCode    int       `yaml:"exit_code,omitempty" mapstructure:"exit_code,omitempty" json:"exit_code,omitempty"`
	Commit      string    `yaml:"commit,omitempty" mapstructure:"commit,omitempty" json:"commit,omitempty"`
	Version     string    `yaml:"version,omitempty" mapstructure:"version,omitempty" json:"version,omitempty"`
	Datetime    string    `yaml:"datetime,omitempty" mapstructure:"datetime,omitempty" json:"datetime,omitempty"`
	Transaction uuid.UUID `yaml:"transaction,omitempty" mapstructure:"transaction,omitempty" json:"transaction,omitempty"`
	cmd         *cobra.Command
}

func (l *HistoryLog) SetCommand(cmd *cobra.Command) *HistoryLog {

	commandPath := cmd.CommandPath()
	localArgs := cmd.Flags().Args()

	localFlags := []string{}

	cmd.Flags().Visit(func(flag *pflag.Flag) {
		//fmt.Printf("--%s=%s\n", flag.Name, flag.Value)
		localFlags = append(localFlags, fmt.Sprintf("--%s=%s", flag.Name, flag.Value))
	})

	command := strings.Join([]string{
		commandPath,
		strings.Join(localArgs, " "),
		strings.Join(localFlags, " "),
	}, " ")

	l.Command = command
	return l
}

func (l *HistoryLog) GetCommit() {
	c, err := GitGetHeadCommit(sync.Path, workspace.SyncOptions.Local.Branch.Name)
	if err != nil {
		// No repo
		log.Warn(err)
		l.Commit = ""
		return
	}
	l.Commit = c
}

func (l *HistoryLog) Append(message string) error {
	l.Transaction = sync.UUID
	l.Msg = message
	l.Version = version
	l.SetCommand(l.cmd)
	l.GetCommit()
	l.Datetime = time.Now().Format(time.RFC3339)

	json, err := json.Marshal(l)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(filepath.Join(workspace.LocalPath, "history.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(json)
	if err != nil {
		return err
	}

	if _, err = f.WriteString("\n"); err != nil {
		return err
	}

	return nil
}

func (l *HistoryLog) AppendAndCommit(message string) error {
	err := l.Append(message)
	if err != nil {
		return err
	}

	_, err = GitCommitAll(sync.Path, message)
	if err != nil {
		return err
	}

	return nil
}
