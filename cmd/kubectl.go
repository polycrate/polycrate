package cmd

import (
	"os"

	"polycrate/pkg/kubectl"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	// Import to initialize client auth plugins.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var kubectlCmd = &cobra.Command{
	Use:                "k",
	Short:              "kubectl wrapper",
	Long:               ``,
	DisableFlagParsing: true,
	//Args:  cobra.ExactArgs(2), // https://github.com/spf13/cobra/blob/master/user_guide.md
	Run: func(cmd *cobra.Command, args []string) {
		// _w := cmd.Flags().Lookup("workspace").Value.String()

		tx := polycrate.Transaction()
		tx.SetCommand(cmd)
		defer tx.Stop()

		workspace, err := polycrate.LoadWorkspace(tx, cwd, true)
		if err != nil {
			tx.Log.Fatal(err)
		}

		os.Setenv("KUBECONFIG", workspace.Kubeconfig.LocalPath)
		//os.Setenv("KUBECONFIG", "/Users/derfabianpeter/.polycrate/workspaces/ayedo/aycloud-platform-1/artifacts/blocks/k8s/kubeconfig.yml")

		//cmd.Flags().VisitAll(hideFlag)

		kubectl.Main(os.Args)
	},
}

func hideFlag(flag *pflag.Flag) {
	flag.Hidden = true
}

func init() {
	rootCmd.AddCommand(kubectlCmd)
}
