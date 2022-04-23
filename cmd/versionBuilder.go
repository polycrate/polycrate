package cmd

import (
	"fmt"
	"os"
)

// not using a git api for go for now as they do not provide all features, falling back to syscalls
// builds a semantic versioning string which is returned by this function
func BuildVersion(path string) string {
	fileInfo, err := os.Stat(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Given path (%s) does not exist.\n", path)
		os.Exit(1)
	}

	if fileInfo.IsDir() {
		os.Chdir(path)
	} else {
		fmt.Fprintf(os.Stderr, "Given path (%s) is not a directory.\n", path)
		os.Exit(1)
	}
	if !isGitRepo() {
		fmt.Fprintf(os.Stderr, "Not a git repo!\n")
		os.Exit(1)
	}
	branch := getBranch()
	if !isValidString(branch) {
		fmt.Fprintf(os.Stderr, "The current branch name contains illegal characters (not in [.-a-zA-Z0-9]): %s\n", branch)
		os.Exit(1)
	}
	hash := getHash()
	tag := getTag()
	return fmt.Sprintf("%s-%s.dev-%s", tag, branch, hash)
}
