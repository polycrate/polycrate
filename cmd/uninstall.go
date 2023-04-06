package cmd

import (
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall $BLOCK",
	Short: "Run the 'uninstall' action of a block",
	Run: func(cmd *cobra.Command, args []string) {
		actionsRunCmd.Run(cmd, []string{"uninstall", args[0]})
	},
	Args: cobra.ExactArgs(1), // https://github.com/spf13/cobra/blob/master/user_guide.md
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
