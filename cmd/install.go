package cmd

import (
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install $BLOCK",
	Short: "Run the 'install' action of a block",
	Run: func(cmd *cobra.Command, args []string) {
		actionsRunCmd.Run(cmd, []string{args[0], "install"})
	},
	Args: cobra.ExactArgs(1), // https://github.com/spf13/cobra/blob/master/user_guide.md
}

func init() {
	rootCmd.AddCommand(installCmd)
}
