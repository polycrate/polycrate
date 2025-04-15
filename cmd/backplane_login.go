package cmd

import (
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var backplanePassword string
var backplaneUsername string

// installCmd represents the install command
var backplaneLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to Backplane",
	Long:  ``,
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		tx := polycrate.Transaction()
		tx.SetCommand(cmd)
		defer tx.Stop()

		//polycrate.Config.Backplane.Url

		if polycrate.Config.Backplane.Username == "" {
			// Ask for a username via prompt
			usernamePrompt := promptui.Prompt{
				Label: "Backplane Username",
			}

			result, err := usernamePrompt.Run()
			if err != nil {
				tx.Log.Fatalf("Failed to obtain Backplane Username: %s", err)
			}
			polycrate.Config.Backplane.Username = result
		}

		if polycrate.Config.Backplane.Password == "" {
			// Ask for a username via prompt
			passwordPrompt := promptui.Prompt{
				Label: "Backplane Password",
			}

			result, err := passwordPrompt.Run()
			if err != nil {
				tx.Log.Fatalf("Failed to obtain Backplane Password: %s", err)
			}
			polycrate.Config.Backplane.Password = result
		}

		tx.Log.Infof("Logging in to Backplane with Username '%s' and Password '%s'", polycrate.Config.Backplane.Username, polycrate.Config.Backplane.Password)

		backplane := new(Backplane)
		err := backplane.Login(tx)
		if err != nil {
			tx.Log.Fatalf("Failed to log in to Backplane: %s", err)
		}
	},
}

func init() {
	backplaneLoginCmd.Flags().StringVar(&backplanePassword, "password", "", "Backplane password")
	backplaneLoginCmd.Flags().StringVar(&backplaneUsername, "username", "", "Backplane username")

	backplaneCmd.AddCommand(backplaneLoginCmd)
}
