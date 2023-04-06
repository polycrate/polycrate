package cmd

import (
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var pruneCmd = &cobra.Command{
	Use:   "prune $BLOCK",
	Short: "Run the 'prune' action of a block",
	Run: func(cmd *cobra.Command, args []string) {
		actionsRunCmd.Run(cmd, []string{"prune", args[0]})
	},
	Args: cobra.ExactArgs(1), // https://github.com/spf13/cobra/blob/master/user_guide.md
}

func init() {
	rootCmd.AddCommand(pruneCmd)
}
