package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/filesystem"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	goErrors "errors"

	"github.com/go-git/go-billy/v5/osfs"
)

// checks if the current directory is a git repo by running git status
func isGitRepo() bool {
	_, err := exec.Command("git", "status").Output()

	return err == nil
}

// tries to get the current git tag, if it fails 0.0.1 is returned
func getTag() string {
	out, err := exec.Command("git", "describe", "--abbrev=0", "--tags").Output()

	if err != nil {
		return "0.0.1"
	}
	return strings.TrimSpace(string(out))
}

// tries to get the current git branch, exists the program on failure
func getBranch() string {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while trying to get branch: %s\n", err)
		os.Exit(1)
	}
	return strings.TrimSpace(string(out))
}

// tries to get the current git has in short form, exists the program on failure
func getHash() string {
	out, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while trying to get hash: %s\n", err)
		os.Exit(1)
	}
	return strings.TrimSpace(string(out))
}

func commitContext(message string) error {
	// Opens an already existing repository.
	log.Debug("Opening git repository at ", workspaceDir)
	r, err := git.PlainOpen(workspaceDir)

	if err != nil {
		// we assume the repository does not exist
		log.Debug("Initalizing git repository at ", workspaceDir)
		fs := osfs.New(workspaceDir)
		gitfs := osfs.New(path.Join(workspaceDir, ".git"))
		storer := filesystem.NewStorage(gitfs, cache.NewObjectLRUDefault())
		r, err = git.Init(storer, fs)
		CheckErr(err)
	}

	w, err := r.Worktree()
	CheckErr(err)

	loadGitConfig()

	commit, err := w.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  gitConfigObject.GetString("user.name"),
			Email: gitConfigObject.GetString("user.email"),
			When:  time.Now(),
		},
	})
	CheckErr(err)
	log.Debug("Added commit: '", message, "'")

	_, err = r.CommitObject(commit)
	CheckErr(err)

	return err

}

func loadGitConfig() {
	gitConfigPath := filepath.Join(home, ".gitconfig")

	gitConfigObject.SetConfigFile(gitConfigPath)
	gitConfigObject.SetConfigType("ini")

	if err := gitConfigObject.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Debug(".gitconfig not found")
		} else {
			log.Debug(err)
		}
	}
}

func cloneRepository(repository string, path string, branch string, tag string) error {
	var args = []string{"clone"}
	if branch != "" {
		args = append(args, "-b", branch, "--single-branch")
	} else if tag != "" {
		args = append(args, "-b", tag, "--single-branch")
	}

	args = append(args, "--depth=1", repository, path)

	log.Debug("Running command: git ", strings.Join(args, " "))

	out, err := exec.Command("git", args...).Output()

	log.Debug(string(out))
	if err != nil {
		log.Error(err)
		return goErrors.New(string(out))
	}
	return nil
}
