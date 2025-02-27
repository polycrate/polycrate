package cmd

import (
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var veleroCmd = &cobra.Command{
	Use:   "velero",
	Short: "velero wrapper",
	Long:  ``,
	//DisableFlagParsing: true,
	//Args:  cobra.ExactArgs(2), // https://github.com/spf13/cobra/blob/master/user_guide.md
	Run: func(cmd *cobra.Command, args []string) {
		_w := cmd.Flags().Lookup("workspace").Value.String()

		tx := polycrate.Transaction()
		tx.SetCommand(cmd)
		defer tx.Stop()

		workspace, err := polycrate.LoadWorkspace(tx, _w, true)
		if err != nil {
			tx.Log.Fatal(err)
		}

		os.Setenv("KUBECONFIG", workspace.Kubeconfig.LocalPath)

		var _cmd []string
		_cmd = append(_cmd, "-c")

		var __cmd []string
		__cmd = append(__cmd, "/usr/local/bin/velero")
		__cmd = append(__cmd, args...)
		result := strings.Join(__cmd, " ")

		_cmd = append(_cmd, result)

		err = workspace.RunContainer(tx, "polycrate-velero", workspace.LocalPath, _cmd)
		if err != nil {
			log.Fatal(err)
		}

	},
}

func init() {
	rootCmd.AddCommand(veleroCmd)
}
