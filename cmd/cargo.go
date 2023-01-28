// curl -X 'GET' \
//   'https://cargo.ayedo.cloud/api/v2.0/projects/ayedo/repositories?page=1&page_size=10' \
//   -H 'accept: application/json'
package cmd

import "github.com/spf13/cobra"

var cargoCmd = &cobra.Command{
	Use:   "cargo",
	Short: "Deal with the Polycrate registry",
	Long:  ``,
	Args:  cobra.ExactArgs(0), // https://github.com/spf13/cobra/blob/master/user_guide.md
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cargoCmd)
}
