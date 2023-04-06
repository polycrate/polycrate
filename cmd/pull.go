package cmd

import (
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var pullCmd = &cobra.Command{
	Use:   blocksPullCmd.Use,
	Short: blocksPullCmd.Short,
	Run:   blocksPullCmd.Run,
	Args:  cobra.ExactArgs(1), // https://github.com/spf13/cobra/blob/master/user_guide.md
}

func init() {
	rootCmd.AddCommand(pullCmd)
}
