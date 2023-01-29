package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5"
	gitConfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	log "github.com/sirupsen/logrus"
)

func GitGetRemoteUrl(ctx context.Context, path string, name string) (string, error) {
	remoteArgs := []string{
		"remote",
		"get-url",
		name,
	}
	output, err := GitExecute(ctx, path, remoteArgs)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

func GitUpdateRemoteUrl(ctx context.Context, path string, remote string, url string) error {
	err := GitDeleteRemote(ctx, path, remote)
	if err != nil {
		return err
	}

	err = GitCreateRemote(ctx, path, remote, url)
	if err != nil {
		return err
	}
	return nil
}

func GitCreateRepo(ctx context.Context, path string, remote string, branch string, url string) error {
	err := CreateDir(path)
	if err != nil {
		return err
	}

	_, err = GitInit(ctx, path)
	if err != nil {
		return err
	}

	_, err = GitCheckout(ctx, path, branch, true)
	if err != nil {
		return err
	}

	err = GitCreateRemote(ctx, path, remote, url)
	if err != nil {
		return err
	}

	//Make an initial commit
	_, err = GitCommitAll(ctx, path, "Initial sync")
	if err != nil {
		return err
	}

	_, err = GitPush(ctx, path, remote, branch)
	if err != nil {
		return err
	}

	return nil
}

func GitCheckout(ctx context.Context, path string, branch string, create bool) (string, error) {
	checkoutArgs := []string{
		"checkout",
	}

	if create {
		checkoutArgs = append(checkoutArgs, "-b")
	}

	checkoutArgs = append(checkoutArgs, branch)

	output, err := GitExecute(ctx, path, checkoutArgs)
	if err != nil {
		return "", err
	}

	return output, nil
}

func GitInit(ctx context.Context, path string) (string, error) {
	log.WithFields(log.Fields{
		"path": path,
	}).Debugf("Initializing git repository")

	initArgs := []string{
		"init",
	}

	output, err := GitExecute(ctx, path, initArgs)
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

func GitDeleteRemote(ctx context.Context, path string, name string) error {
	log.WithFields(log.Fields{
		"path":   workspace.LocalPath,
		"remote": name,
	}).Debugf("Deleting remote from repository")

	args := []string{
		"remote",
		"remove",
		name,
	}

	output, err := GitExecute(ctx, path, args)
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

func GitHasRemote(ctx context.Context, path string) bool {
	args := []string{
		"remote",
		"-v",
	}
	output, err := GitExecute(ctx, path, args)
	if err != nil {
		return false
	}

	if output != "" {
		return true
	}
	return false
}

func GitIsRepo(ctx context.Context, path string) bool {
	args := []string{
		"status",
	}
	_, err := GitExecute(ctx, path, args)

	return err == nil
}

func GitOpenRepo(path string) (*git.Repository, error) {
	repo, err := git.PlainOpen(path)

	if err != nil {
		return nil, err
	}
	return repo, nil
}

// func GitRepoExists(path string) error {
// 	repo, err := GitOpenRepo(path)
// 	if err != nil {
// 		return err
// 	}

// 	revHash, err := repo.ResolveRevision(plumbing.Revision(polycrateConfig.Sync.DefaultBranch))
// 	if err != nil {
// 		return err
// 	}
// 	_, err = repo.CommitObject(*revHash)

// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

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

func GitHasChanges(ctx context.Context, path string) bool {
	statusArgs := []string{
		"status",
		"--porcelain",
	}

	output, err := GitExecute(ctx, path, statusArgs)
	if err != nil {
		return false
	}

	if output != "" {
		return true
	}
	return false
}

func GitGetUserEmail(ctx context.Context) (string, error) {
	args := []string{
		"config",
		"user.email",
	}
	output, err := GitExecute(ctx, polycrateConfigDir, args)
	if err != nil {
		log.Error(err)
		return "", err
	}

	return output, nil
}
func GitGetUserName(ctx context.Context) (string, error) {
	args := []string{
		"config",
		"user.name",
	}

	output, err := GitExecute(ctx, polycrateConfigDir, args)
	if err != nil {
		log.Error(err)
		return "", err
	}

	return output, nil
}

func GitCommitAll(ctx context.Context, path string, message string) (string, error) {
	addArgs := []string{
		"add",
		".",
	}

	_, err := GitExecute(ctx, path, addArgs)
	if err != nil {
		return "", err
	}

	commitArgs := []string{
		"commit",
		"--all",
		fmt.Sprintf("--message=%s", message),
	}

	_, err = GitExecute(ctx, path, commitArgs)
	if err != nil {
		return "", err
	}

	// Get current commit hash
	hashArgs := []string{
		"rev-parse",
		"HEAD",
	}

	hash, err := GitExecute(ctx, path, hashArgs)
	if err != nil {
		return "", err
	}

	return hash, nil
}

func GitExecute(ctx context.Context, path string, args []string) (string, error) {
	pwd, _ := os.Getwd()

	// Change to commit dir
	err := os.Chdir(path)
	if err != nil {
		return "", err
	}

	_, output, err := RunCommandWithOutput(ctx, nil, "git", args...)
	if err != nil {
		return output, err
	}

	err = os.Chdir(pwd)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(output), nil
}

func GitFetch(ctx context.Context, path string) (string, error) {
	fetchArgs := []string{
		"fetch",
	}
	output, err := GitExecute(ctx, path, fetchArgs)
	if err != nil {
		return "", err
	}

	return output, nil
}

func GitPush(ctx context.Context, path string, remote string, branch string) (string, error) {
	pushArgs := []string{
		"push",
		"--set-upstream",
		remote,
		branch,
	}
	output, err := GitExecute(ctx, path, pushArgs)

	if err != nil {
		return "", err
	}

	return output, nil
}
func GitSetUpstreamTracking(ctx context.Context, path string, remote string, branch string) (string, error) {
	pushArgs := []string{
		"branch",
		"-u",
		strings.Join([]string{remote, branch}, "/"),
		branch,
	}
	output, err := GitExecute(ctx, path, pushArgs)

	if err != nil {
		return "", err
	}

	return output, nil
}

func GitPull(ctx context.Context, path string, remote string, branch string) (string, error) {
	pullArgs := []string{
		"pull",
		remote,
		branch,
	}
	output, err := GitExecute(ctx, path, pullArgs)

	if err != nil {
		return "", err
	}

	return output, nil
}

func GitBehindBy(ctx context.Context, path string) (int, error) {
	// behind_count = $(git rev-list --count HEAD..@{u}).
	revArgs := []string{
		"rev-list",
		"--count",
		"HEAD..@{u}",
	}

	output, err := GitExecute(ctx, path, revArgs)
	if err != nil {
		return 0, err
	}
	int, err := strconv.Atoi(output)
	if err != nil {
		return 0, err
	}
	return int, nil
}

func GitGetHeadCommit(ctx context.Context, path string, revision string) (string, error) {
	revArgs := []string{
		"rev-parse",
		revision,
	}

	output, err := GitExecute(ctx, path, revArgs)
	if err != nil {
		return "", err
	}
	return output, nil
}

func GitAheadBy(ctx context.Context, path string) (int, error) {
	// ahead_count = $(git rev-list --count @{u}..HEAD)
	revArgs := []string{
		"rev-list",
		"--count",
		"@{u}..HEAD",
	}

	output, err := GitExecute(ctx, path, revArgs)
	if err != nil {
		return 0, err
	}

	int, err := strconv.Atoi(output)
	if err != nil {
		return 0, err
	}
	return int, nil
}

func GitCreateRemote(ctx context.Context, path string, name string, url string) error {
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

	_, err := GitExecute(ctx, path, remoteArgs)

	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"branch": GitDefaultBranch,
		"remote": GitDefaultRemote,
	}).Debugf("Fetching remote")

	_, err = GitFetch(ctx, path)

	if err != nil {
		return err
	}

	return nil
}
