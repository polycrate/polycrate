package cmd

import (
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var pushCmd = &cobra.Command{
	Use:   blocksPushCmd.Use,
	Short: blocksPushCmd.Short,
	Run:   blocksPushCmd.Run,
	Args:  cobra.ExactArgs(1), // https://github.com/spf13/cobra/blob/master/user_guide.md
}

func init() {
	rootCmd.AddCommand(pushCmd)
}
