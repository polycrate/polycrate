package kubectl

import (
	"os"
	"strings"

	"k8s.io/component-base/cli"
	kcmd "k8s.io/kubectl/pkg/cmd"
	"k8s.io/kubectl/pkg/cmd/util"

	// Import to initialize client auth plugins.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func Main(args []string) {
	main(args)
}

func main(args []string) {
	// Build new args
	// Remove first Arg (k)
	// Remove everything with a -- that comes before k
	var new_args []string
	var passed_k bool
	for _, arg := range args {
		if arg == "k" {
			passed_k = true
			continue
		}

		if passed_k {
			new_args = append(new_args, arg)
		} else {
			if strings.HasPrefix(arg, "--") {
				// Ignore the arg
				//} else if (arg == "k") && i+1 < len(os.Args) {
			} else {
				new_args = append(new_args, arg)
			}
		}
	}
	os.Args = new_args
	command := kcmd.NewDefaultKubectlCommand()
	if err := cli.RunNoErrOutput(command); err != nil {
		util.CheckErr(err)
	}
}
