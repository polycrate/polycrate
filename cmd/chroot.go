//go:build linux
// +build linux

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/spf13/cobra"
)

var chrootCmd = &cobra.Command{
	Hidden: true,
	Use:    "chroot",
	Short:  "Change Root",
	Long:   `Change Root`,
	Run: func(cmd *cobra.Command, args []string) {
		_w := cmd.Flags().Lookup("workspace").Value.String()

		tx := polycrate.Transaction().SetCommand(cmd)
		defer tx.Stop()

		_, err := polycrate.LoadWorkspace(tx, _w, true)
		if err != nil {
			tx.Log.Fatal(err)
		}

		run()
		child()
	},
}

func init() {
	rootCmd.AddCommand(chrootCmd)
}

func run() {
	fmt.Printf("Running %v as PID %d\n", os.Args[2:], os.Getpid())

	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}

	if err := cmd.Run(); err != nil {
		fmt.Println("Error running the /proc/self/exe command:", err)
		os.Exit(1)
	}
}

func child() {
	fmt.Printf("Running %v as PID %d\n", "ls -la", os.Getpid())

	cmd := exec.Command("ls", "-la")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := syscall.Mount("proc", "proc", "proc", 0, ""); err != nil {
		fmt.Println("Error mounting proc:", err)
		os.Exit(1)
	}

	if err := cmd.Run(); err != nil {
		fmt.Println("Error running the child command:", err)
		os.Exit(1)
	}

}
