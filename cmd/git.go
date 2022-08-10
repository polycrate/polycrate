package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5"
	gitConfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	log "github.com/sirupsen/logrus"
)

func GitGetRemoteUrl(path string, name string) (string, error) {
	remoteArgs := []string{
		"remote",
		"get-url",
		name,
	}
	output, err := GitExecute(path, remoteArgs)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

func GitUpdateRemoteUrl(path string, remote string, url string) error {
	err := GitDeleteRemote(path, remote)
	if err != nil {
		return err
	}

	err = GitCreateRemote(path, remote, url)
	if err != nil {
		return err
	}
	return nil
}

func GitCreateRepo(path string, remote string, branch string, url string) error {
	err := CreateDir(path)
	if err != nil {
		return err
	}

	_, err = GitInit(path)
	if err != nil {
		return err
	}

	_, err = GitCheckout(path, branch, true)
	if err != nil {
		return err
	}

	err = GitCreateRemote(path, remote, url)
	if err != nil {
		return err
	}

	//Make an initial commit
	_, err = GitCommitAll(path, "Initial sync")
	if err != nil {
		return err
	}

	_, err = GitPush(path, remote, branch)
	if err != nil {
		return err
	}

	// err = GitCreateAndCheckoutBranch(r, GitDefaultBranch, remote)
	// if err != nil {
	// 	return nil, err
	// }

	return nil
}

func GitCheckout(path string, branch string, create bool) (string, error) {
	checkoutArgs := []string{
		"checkout",
	}

	if create {
		checkoutArgs = append(checkoutArgs, "-b")
	}

	checkoutArgs = append(checkoutArgs, branch)

	output, err := GitExecute(path, checkoutArgs)
	if err != nil {
		return "", err
	}

	return output, nil
}

func GitInit(path string) (string, error) {
	log.WithFields(log.Fields{
		"path": path,
	}).Debugf("Initializing git repository")

	initArgs := []string{
		"init",
	}

	output, err := GitExecute(path, initArgs)
	if err != nil {
		return "", err
	}
	log.Debugf(output)

	return output, nil
}

func GitCreateAndCheckoutBranch(repository *git.Repository, branch string, remote string) error {
	localRef := plumbing.NewBranchReferenceName(branch)
	w, err := repository.Worktree()
	if err != nil {
		return err
	}

	log.Warn("Marker")

	opts := &gitConfig.Branch{
		Name:   branch,
		Remote: remote,
	}

	log.WithFields(log.Fields{
		"branch": branch,
	}).Debugf("Creating branch")

	if err := repository.CreateBranch(opts); err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"branch": branch,
	}).Debugf("Checking out branch")
	if err := w.Checkout(&git.CheckoutOptions{Branch: localRef, Create: true}); err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"branch": branch,
		"remote": remote,
	}).Debugf("Setting remote")

	remoteRef := plumbing.NewRemoteReferenceName(remote, branch)

	newReference := plumbing.NewSymbolicReference(localRef, remoteRef)

	log.WithFields(log.Fields{
		"branch": branch,
		"remote": remote,
	}).Debugf("Syncing with remote")
	if err := repository.Storer.SetReference(newReference); err != nil {
		return err
	}

	return nil
}

func GitDeleteRemote(path string, name string) error {
	log.WithFields(log.Fields{
		"path":   workspace.LocalPath,
		"remote": name,
	}).Debugf("Deleting remote from repository")

	args := []string{
		"remote",
		"remove",
		name,
	}

	output, err := GitExecute(path, args)
	if err != nil {
		return err
	}
	log.Debugf(output)
	return nil
}

// func GitHasRemote(repository *git.Repository, remote string) bool {
// 	r, err := repository.Remote(remote)
// 	if err != nil {
// 		return false
// 	}

// 	url := r.Config().URLs[0]
// 	return url != ""
// }

func GitHasRemote(path string, remote string) bool {
	args := []string{
		"remote",
		"get-url",
		remote,
	}
	output, err := GitExecute(path, args)
	if err != nil {
		return false
	}

	if output != "" {
		return true
	}
	return false
}

func GitIsRepo(path string) bool {
	args := []string{
		"status",
	}
	_, err := GitExecute(path, args)

	return err == nil
}

func GitOpenRepo(path string) (*git.Repository, error) {
	repo, err := git.PlainOpen(path)

	if err != nil {
		return nil, err
	}
	return repo, nil
}

func GitRepoExists(path string) error {
	repo, err := GitOpenRepo(path)
	if err != nil {
		return err
	}

	revHash, err := repo.ResolveRevision(plumbing.Revision(config.Sync.DefaultBranch))
	if err != nil {
		return err
	}
	revCommit, err := repo.CommitObject(*revHash)
	fmt.Println(revCommit)

	if err != nil {
		return err
	}
	return nil
}

func GitGetStatus(r *git.Repository) (*git.Status, error) {
	w, err := r.Worktree()
	if err != nil {
		return nil, err
	}

	// git status
	ws, err := w.Status()
	if err != nil {
		return nil, err
	}

	return &ws, nil
}

func GitIsAncestor(r *git.Repository) (bool, error) {
	remoteRev := fmt.Sprintf("%s/%s", GitDefaultRemote, GitDefaultBranch)
	remoteRevHash, err := r.ResolveRevision(plumbing.Revision(remoteRev))
	if err != nil {
		return false, err
	}

	log.Debugf("Remote Rev Hash: %s", remoteRevHash)

	remoteRevCommit, err := r.CommitObject(*remoteRevHash)
	if err != nil {
		return false, err
	}
	log.Debugf("Remote Rev Commit: %s", remoteRevCommit.Hash)

	headRef, err := r.Head()
	if err != nil {
		return false, err
	}
	log.Debugf("HEAD REF: %s", headRef)

	// ... retrieving the commit object
	headCommit, err := r.CommitObject(headRef.Hash())
	if err != nil {
		return false, err
	}
	log.Debugf("HEAD Commit: %s", headCommit.Hash)

	isAncestor, err := headCommit.IsAncestor(remoteRevCommit)

	if err != nil {
		return false, err
	}
	return isAncestor, nil
}

func GitHasChanges(path string) bool {
	statusArgs := []string{
		"status",
		"--porcelain",
	}

	output, err := GitExecute(path, statusArgs)
	if err != nil {
		return false
	}

	if output != "" {
		return true
	}
	return false
}

func GitCommitAll(path string, message string) (string, error) {
	addArgs := []string{
		"add",
		".",
	}

	_, err := GitExecute(path, addArgs)
	if err != nil {
		return "", err
	}

	commitArgs := []string{
		"commit",
		"--all",
		fmt.Sprintf("--message=%s", message),
	}

	_, err = GitExecute(path, commitArgs)
	if err != nil {
		return "", err
	}

	// Get current commit hash
	hashArgs := []string{
		"rev-parse",
		"HEAD",
	}

	_, err = GitExecute(path, hashArgs)
	if err != nil {
		return "", err
	}

	return "", nil
}

func GitExecute(path string, args []string) (string, error) {
	pwd, _ := os.Getwd()

	// Change to commit dir
	err := os.Chdir(path)
	if err != nil {
		return "", err
	}

	_, output, err := RunCommandWithOutput("git", args...)
	if err != nil {
		return output, err
	}

	err = os.Chdir(pwd)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(output), nil
}

func GitFetch(path string) (string, error) {
	fetchArgs := []string{
		"fetch",
	}
	output, err := GitExecute(path, fetchArgs)
	if err != nil {
		return "", err
	}

	return output, nil
}

func GitPush(path string, remote string, branch string) (string, error) {
	pushArgs := []string{
		"push",
		"--set-upstream",
		remote,
		branch,
	}
	output, err := GitExecute(path, pushArgs)

	if err != nil {
		return "", err
	}

	return output, nil
}
func GitSetUpstreamTracking(path string, remote string, branch string) (string, error) {
	pushArgs := []string{
		"branch",
		"--set-upstream-to",
		remote,
		branch,
	}
	output, err := GitExecute(path, pushArgs)

	if err != nil {
		return "", err
	}

	return output, nil
}

func GitPull(path string, remote string, branch string) (string, error) {
	pullArgs := []string{
		"pull",
		remote,
		branch,
	}
	output, err := GitExecute(path, pullArgs)

	if err != nil {
		return "", err
	}

	return output, nil
}

func GitBehindBy(path string) (int, error) {
	// behind_count = $(git rev-list --count HEAD..@{u}).
	revArgs := []string{
		"rev-list",
		"--count",
		"HEAD..@{u}",
	}

	output, err := GitExecute(path, revArgs)
	if err != nil {
		return 0, err
	}
	int, err := strconv.Atoi(output)
	if err != nil {
		return 0, err
	}
	return int, nil
}

func GitGetHeadCommit(path string, revision string) (string, error) {
	revArgs := []string{
		"rev-parse",
		revision,
	}

	output, err := GitExecute(path, revArgs)
	if err != nil {
		return "", err
	}
	return output, nil
}

func GitAheadBy(path string) (int, error) {
	// ahead_count = $(git rev-list --count @{u}..HEAD)
	revArgs := []string{
		"rev-list",
		"--count",
		"@{u}..HEAD",
	}

	output, err := GitExecute(path, revArgs)
	if err != nil {
		return 0, err
	}

	int, err := strconv.Atoi(output)
	if err != nil {
		return 0, err
	}
	return int, nil
}

func GitCreateRemote(path string, name string, url string) error {
	log.WithFields(log.Fields{
		"path":   workspace.LocalPath,
		"remote": name,
		"url":    url,
	}).Debugf("Adding remote to repository")

	remoteArgs := []string{
		"remote",
		"add",
		name,
		url,
	}

	_, err := GitExecute(path, remoteArgs)

	// _, err := r.CreateRemote(&gitConfig.RemoteConfig{
	// 	Name:  name,
	// 	URLs:  []string{url},
	// 	Fetch: []gitConfig.RefSpec{"+refs/heads/*:refs/remotes/origin/*"},
	// })

	if err != nil {
		return err
	}

	// localRef := plumbing.NewBranchReferenceName(fmt.Sprintf("refs/heads/%s", GitDefaultBranch))
	// remoteRef := plumbing.NewRemoteReferenceName(GitDefaultRemote, GitDefaultBranch)
	// ref, err := r.Reference(remoteRef, true)
	// fmt.Println(ref)
	// log.Warn("Marker")
	// if err != nil {
	// 	return err
	// }
	// fmt.Println(ref.Hash())
	// if err := r.Storer.SetReference(plumbing.NewHashReference(localRef, ref.Hash())); err != nil {
	// 	return err
	// }

	// currentConfig, err := r.Config()
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// main := gitConfig.Branch{
	// 	Name:   GitDefaultBranch,
	// 	Remote: GitDefaultRemote,
	// 	Merge:  "origin",
	// }
	// currentConfig.Branches[GitDefaultBranch] = &main

	// printObject(currentConfig)
	// r.SetConfig(currentConfig)

	log.WithFields(log.Fields{
		"branch": GitDefaultBranch,
		"remote": GitDefaultRemote,
	}).Debugf("Fetching remote")

	_, err = GitFetch(path)

	if err != nil {
		return err
	}

	// err = r.Fetch(&git.FetchOptions{
	// 	RemoteName: GitDefaultRemote,
	// })

	// if err != nil {
	// 	if errors.Is(err, transport.ErrEmptyRemoteRepository) {
	// 		log.WithFields(log.Fields{
	// 			"branch": GitDefaultBranch,
	// 			"remote": GitDefaultRemote,
	// 		}).Debugf("Remote repository is empty")
	// 	} else if errors.Is(err, git.NoErrAlreadyUpToDate) {
	// 		log.WithFields(log.Fields{
	// 			"branch": GitDefaultBranch,
	// 			"remote": GitDefaultRemote,
	// 		}).Debugf("Remote local repository is up to date")
	// 	} else {
	// 		return err
	// 	}
	// }

	// localRef := plumbing.NewBranchReferenceName(GitDefaultBranch)
	// remoteRef := plumbing.NewRemoteReferenceName(GitDefaultRemote, GitDefaultBranch)

	// newReference := plumbing.NewSymbolicReference(localRef, remoteRef)

	// if err := repository.Storer.SetReference(newReference); err != nil {
	// 	return err
	// }

	return nil
}
