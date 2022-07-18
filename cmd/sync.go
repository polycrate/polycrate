package cmd

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type SyncProviderFactory func() PolycrateProvider

type SyncBranch struct {
	Name string `yaml:"name,omitempty" mapstructure:"name,omitempty" json:"name,omitempty"`
}

type SyncRemoteOptions struct {
	Branch SyncBranch
	Name   string `yaml:"name,omitempty" mapstructure:"name,omitempty" json:"name,omitempty"`
	Url    string `yaml:"url,omitempty" mapstructure:"url,omitempty" json:"url,omitempty"`
}

type SyncLocalOptions struct {
	Branch SyncBranch `yaml:"branch,omitempty" mapstructure:"branch,omitempty" json:"branch,omitempty"`
}

type SyncOptions struct {
	Local   SyncLocalOptions  `yaml:"local,omitempty" mapstructure:"local,omitempty" json:"local,omitempty"`
	Remote  SyncRemoteOptions `yaml:"remote,omitempty" mapstructure:"remote,omitempty" json:"remote,omitempty"`
	Enabled bool              `yaml:"enabled,omitempty" mapstructure:"enabled,omitempty" json:"enabled,omitempty"`
	Auto    bool              `yaml:"auto,omitempty" mapstructure:"auto,omitempty" json:"auto,omitempty"`
}

type SyncReference struct {
	Commit    string     `yaml:"commit,omitempty" mapstructure:"commit,omitempty" json:"commit,omitempty"`
	Branch    SyncBranch `yaml:"branch,omitempty" mapstructure:"branch,omitempty" json:"branch,omitempty"`
	Reference string     `yaml:"reference,omitempty" mapstructure:"reference,omitempty" json:"reference,omitempty"`
}

type Sync struct {
	UUID      uuid.UUID         `yaml:"uuid,omitempty" mapstructure:"uuid,omitempty" json:"uuid,omitempty"`
	Provider  PolycrateProvider `yaml:"provider,omitempty" mapstructure:"provider,omitempty" json:"provider,omitempty"`
	Path      string            `yaml:"path,omitempty" mapstructure:"path,omitempty" json:"path,omitempty"`
	Options   SyncOptions       `yaml:"options,omitempty" mapstructure:"options,omitempty" json:"options,omitempty"`
	err       error
	loaded    bool
	Status    string        `yaml:"status,omitempty" mapstructure:"status,omitempty" json:"status,omitempty"`
	LocalRef  SyncReference `yaml:"local_ref,omitempty" mapstructure:"local_ref,omitempty" json:"local_ref,omitempty"`
	RemoteRef SyncReference `yaml:"remote_ref,omitempty" mapstructure:"remote_ref,omitempty" json:"remote_ref,omitempty"`
	History   HistoryLog    `yaml:"history,omitempty" mapstructure:"history,omitempty" json:"history,omitempty"`
}

func (s *Sync) Print() {
	printObject(s)
}

func (s *Sync) Sync() *Sync {
	if !s.loaded {
		s.err = fmt.Errorf("sync module not loaded")
		return s
	}

	if s.Options.Enabled {

		s.UpdateStatus().Flush()

		log.WithFields(log.Fields{
			"workspace": workspace.Name,
			"module":    "sync",
		}).Infof("Updating status")

		switch status := s.Status; status {
		case "changed":
			log.WithFields(log.Fields{
				"workspace": workspace.Name,
				"module":    "sync",
			}).Infof("Changes found. Comitting and syncing again")
			s.Log("Sync auto-commit").Flush()
			s.Sync().Flush()
		case "synced":
			log.WithFields(log.Fields{
				"workspace": workspace.Name,
				"module":    "sync",
			}).Infof("Up-to-date")
		case "diverged":
			log.WithFields(log.Fields{
				"workspace": workspace.Name,
				"module":    "sync",
			}).Fatalf("Sync error - run `polycrate sync status` for more information")
		case "ahead":
			log.WithFields(log.Fields{
				"workspace": workspace.Name,
				"module":    "sync",
			}).Infof("New commits found. Syncing")
			s.Push().Flush()
		case "behind":
			log.WithFields(log.Fields{
				"workspace": workspace.Name,
				"module":    "sync",
			}).Infof("New commits found. Syncing")
			s.Pull().Flush()
		}
	} else {
		s.err = fmt.Errorf("sync disabled")
		return s
	}

	return s
}

func (s *Sync) Push() *Sync {
	_, err := GitPush(s.Path, s.Options.Remote.Name, s.Options.Remote.Branch.Name)
	if err != nil {
		s.err = err
		return s
	}
	return s
}

func (s *Sync) Pull() *Sync {
	_, err := GitPull(s.Path, s.Options.Remote.Name, s.Options.Remote.Branch.Name)
	if err != nil {
		s.err = err
		return s
	}
	return s
}

func (s *Sync) Load() *Sync {
	if workspace.SyncOptions.Enabled {
		s.UUID = uuid.New()
		log.WithFields(log.Fields{
			"path":      workspace.LocalPath,
			"workspace": workspace.Name,
			"module":    "sync",
		}).Debugf("Loading sync module")

		s.LoadProvider().Flush()
		s.LoadRepo().Flush()
		s.UpdateStatus().Flush()
		s.loaded = true
	} else {
		log.WithFields(log.Fields{
			"path":      workspace.LocalPath,
			"workspace": workspace.Name,
			"module":    "sync",
		}).Debugf("Not loading sync module. Sync disabled")
		s.loaded = false
	}
	return s
}

func (s *Sync) UpdateStatus() *Sync {
	s.Options.Local.Branch.Name = workspace.SyncOptions.Local.Branch.Name
	s.Options.Remote.Branch.Name = workspace.SyncOptions.Remote.Branch.Name
	s.Options.Remote.Name = workspace.SyncOptions.Remote.Name
	s.Options.Remote.Url = workspace.SyncOptions.Remote.Url
	s.Options.Enabled = workspace.SyncOptions.Enabled
	s.Options.Auto = workspace.SyncOptions.Auto

	if s.Options.Enabled {
		log.WithFields(log.Fields{
			"workspace": workspace.Name,
			"module":    "sync",
			"path":      s.Path,
		}).Debugf("Fetching status of remote repository")
		// https://stackoverflow.com/posts/68187853/revisions
		// Get remote reference
		_, err := GitFetch(s.Path)
		if err != nil {
			s.err = err
			return s
		}

		s.LocalRef.Reference = s.Options.Local.Branch.Name
		s.RemoteRef.Reference = fmt.Sprintf("%s/%s", s.Options.Remote.Name, s.Options.Remote.Branch.Name)

		s.LocalRef.Branch.Name = s.Options.Local.Branch.Name
		s.LocalRef.Commit, err = GitGetHeadCommit(s.Path, s.LocalRef.Reference)
		if err != nil {
			s.err = err
			return s
		}
		s.RemoteRef.Branch.Name = s.Options.Remote.Branch.Name
		s.RemoteRef.Commit, err = GitGetHeadCommit(s.Path, s.RemoteRef.Reference)
		if err != nil {
			s.err = err
			return s
		}

		behindBy, err := GitBehindBy(s.Path)
		if err != nil {
			s.err = err
			return s
		}

		aheadBy, err := GitAheadBy(s.Path)
		if err != nil {
			s.err = err
			return s
		}

		log.WithFields(log.Fields{
			"workspace": workspace.Name,
			"module":    "sync",
			"path":      s.Path,
			"behind":    behindBy,
			"ahead":     aheadBy,
		}).Debugf("Comparing with remote repository")

		// ahead > 0, behind == 0
		if aheadBy != 0 && behindBy == 0 {
			s.Status = "ahead"
		}

		// ahead == 0, behind > 0
		if behindBy != 0 && aheadBy == 0 {
			s.Status = "behind"
		}

		// ahead == 0, behind == 0
		if behindBy == 0 && aheadBy == 0 {
			s.Status = "synced"
		}

		// ahead > 0, behind > 0
		if behindBy != 0 && aheadBy != 0 {
			s.Status = "diverged"
		}

		// Has uncommited changes?
		if GitHasChanges(s.Path) {
			s.Status = "changed"
		}
	}
	return s
}

func (s *Sync) LoadProvider() *Sync {
	provider, err := s.getProvider()
	if err != nil {
		s.err = err
		return s
	}
	s.Provider = provider
	return s
}

func (s *Sync) Flush() *Sync {
	if s.err != nil {
		log.Fatal(s.err)
	}
	return s
}

func (s *Sync) getProvider() (PolycrateProvider, error) {
	if config.Sync.Provider == "gitlab" {
		var pf SyncProviderFactory = NewGitlabSyncProvider
		provider := pf()

		log.WithFields(log.Fields{
			"provider": "gitlab",
			"path":     workspace.LocalPath,
		}).Debugf("Loading sync provider")
		return provider, nil
	}
	return nil, fmt.Errorf("provider not found: %s", config.Sync.Provider)
}

func (s *Sync) LoadRepo() *Sync {

	var err error
	// Check if it's a git repo already
	log.WithFields(log.Fields{
		"path": workspace.LocalPath,
	}).Debugf("Loading local repository")

	if GitIsRepo(workspace.LocalPath) {
		// It's a git repo
		// 1. Get repo's remote
		// 2. Compare with configured remote
		// 2.1 No remote configured? Update configured remote with repo's remote
		// 2.2 No repo remot? Update with configured remote
		// 2.3 Unequal? Update repo remote with configured remote

		// Check remote
		if !GitHasRemote(workspace.LocalPath, GitDefaultRemote) {
			log.WithFields(log.Fields{
				"path": workspace.LocalPath,
			}).Debugf("Local repository has no remote url configured")

			// Check if workspace has a remote url configured
			if workspace.SyncOptions.Remote.Url != "" {
				// Create the remote from the workspace config
				err := GitCreateRemote(workspace.LocalPath, GitDefaultRemote, workspace.SyncOptions.Remote.Url)
				if err != nil {
					s.err = err
					return s
				}
			} else {
				// Exit with error - workspace.SyncOptions.Remote.Url is not configured
				s.err = fmt.Errorf("workspace has no remote configured")
				return s
			}
		} else {
			// Remote is already configured
			// Check if workspace has a remote url configured
			if workspace.SyncOptions.Remote.Url != "" {
				// Check if its url matches the configured remote url
				remoteUrl, err := GitGetRemoteUrl(workspace.LocalPath, GitDefaultRemote)
				if err != nil {
					s.err = err
					return s
				}

				if remoteUrl != workspace.SyncOptions.Remote.Url {
					// Urls don't match
					// Update the repository with the configured remote
					log.WithFields(log.Fields{
						"path": workspace.LocalPath,
					}).Debugf("Local repository remote url doesn't match workspace remote url")

					err := GitUpdateRemoteUrl(workspace.LocalPath, GitDefaultRemote, workspace.SyncOptions.Remote.Url)
					if err != nil {
						s.err = err
						return s
					}
				}
			} else {
				remoteUrl, err := GitGetRemoteUrl(workspace.LocalPath, GitDefaultRemote)
				if err != nil {
					s.err = err
					return s
				}

				// Update the workspace remote with the local remote
				log.WithFields(log.Fields{
					"path": workspace.LocalPath,
				}).Debugf("Workspace has no remote url configured. Updating with local repository remote url")
				log.WithFields(log.Fields{
					"path": workspace.LocalPath,
				}).Warnf("Updating workspace remote url with local repository remote url")
				workspace.updateConfig("sync.remote.url", remoteUrl).Flush()
			}
		}
	} else {
		// Not a git repo
		// Check if a remote url is configured

		if workspace.SyncOptions.Remote.Url != "" {
			// We have a remote url configured
			// Create a repository with the given url
			log.WithFields(log.Fields{
				"path": workspace.LocalPath,
				"url":  workspace.SyncOptions.Remote.Url,
			}).Debugf("Creating new repository with remote url from workspace config")

			err = GitCreateRepo(workspace.LocalPath, workspace.SyncOptions.Remote.Name, workspace.SyncOptions.Remote.Branch.Name, workspace.SyncOptions.Remote.Url)
			if err != nil {
				s.err = err
				return s
			}
		} else {
			// No remote url configured
			log.WithFields(log.Fields{
				"path": workspace.LocalPath,
			}).Warnf("Workspace has no remote url configured.")
			s.err = errors.New("cannot sync this repository. No remote configured in workspace or repository")
			return s

			// // Check for default provider
			// if config.Sync.CreateRepo {
			// 	log.WithFields(log.Fields{
			// 		"path":     workspace.LocalPath,
			// 		"provider": s.Provider.GetName(),
			// 	}).Warnf("Creating project at configured provider")

			// 	group, err := s.Provider.GetDefaultGroup()
			// 	if err != nil {
			// 		s.err = err
			// 		return s
			// 	}

			// 	if group.name != "" {
			// 		// Create project in default group
			// 		project, err := s.Provider.CreateProject(group, workspace.Name)
			// 		if err != nil {
			// 			s.err = err
			// 			return s
			// 		}
			// 		printObject(project)

			// 		// Initialize repository from project data
			// 		var remote_url string
			// 		if config.Gitlab.Transport == "ssh" {
			// 			remote_url = project.remote_ssh
			// 		} else {
			// 			remote_url = project.remote_http
			// 		}

			// 		log.WithFields(log.Fields{
			// 			"path": workspace.LocalPath,
			// 			"url":  remote_url,
			// 		}).Debugf("Configure workspace to sync with remote repository")

			// 		err = GitCreateRepo(workspace.LocalPath, workspace.SyncOptions.Remote.Name, workspace.SyncOptions.Remote.Branch.Name, remote_url)
			// 		if err != nil {
			// 			s.err = err
			// 			return s
			// 		}

			// 		// Update workspace.SyncOptions.Remote.Url with remote_url
			// 		// Update the workspace remote with the local remote
			// 		log.WithFields(log.Fields{
			// 			"path": workspace.LocalPath,
			// 		}).Warnf("Updating workspace remote url with local repository remote url")
			// 		workspace.updateConfig("sync.remote.url", remote_url).Flush()
			// 	} else {
			// 		// No default group given
			// 		// Exit
			// 		s.err = fmt.Errorf("cannot create project - no group defined in provider config")
			// 		return s
			// 	}
			// }
		}
	}

	// if true {
	// 	// It's not a git repo
	// 	log.WithFields(log.Fields{
	// 		"path": workspace.LocalPath,
	// 	}).Debugf("No repository found")

	// 	// Check if workspace.SyncOptions.Remote.Url is configured
	// 	if workspace.SyncOptions.Remote.Url != "" {
	// 		// 1.1.2 git repo not found: check remote repo
	// 		// 1.1.2.1 found: update workspace.SyncOptions.Remote.Url
	// 		// 1.1.2.2 not found: CREATE_REMOTE_REPO, update workspace.SyncOptions.Remote.Url
	// 	}

	// 	if config.Sync.CreateRepo {
	// 		group, err := s.Provider.GetDefaultGroup()
	// 		if err != nil {
	// 			s.err = err
	// 			return s
	// 		}

	// 		if group.name != "" {
	// 			// Create project in default group
	// 			project, err := s.Provider.CreateProject(group, workspace.Name)
	// 			if err != nil {
	// 				s.err = err
	// 				return s
	// 			}
	// 			printObject(project)
	// 		}
	// 	}
	// 	// Try to init a git repo at the given path
	// 	log.WithFields(log.Fields{
	// 		"path": workspace.LocalPath,
	// 	}).Debugf("Initializing new repository")

	// 	_, err := git.PlainInit(workspace.LocalPath, false)
	// 	if err != nil {
	// 		s.err = err
	// 		return s
	// 	}
	// 	return s
	// } else {
	// 	// It's a git repo
	// 	// 1. Get repo's remote
	// 	// 2. Compare with configured remote
	// 	// 2.1 No remote configured? Update configured remote with repo's remote
	// 	// 2.2 No repo remot? Update with configured remote
	// 	// 2.3 Unequal? Update repo remote with configured remote
	// 	log.WithFields(log.Fields{
	// 		"path": workspace.LocalPath,
	// 	}).Debugf("Loaded repository")
	// 	printObject(repository)

	// 	remote, err := repository.Remote("polycrate")
	// 	if err != nil {
	// 		// The remote does not exist
	// 		// Let's create it
	// 		GitCreateRemote(repository, "polycrate", workspace.SyncOptions.Remote.Url)
	// 	}

	// 	fmt.Println(remote)
	// }
	log.WithFields(log.Fields{
		"path": workspace.LocalPath,
	}).Debugf("Local repository loaded")

	s.Path = workspace.LocalPath
	return s
}

func (s *Sync) Validate() {
	// Check if a remote is configured
	// if not, fail
	printObject(s)
}

func (s *Sync) Log(message string) *Sync {
	if s.loaded {
		log.WithFields(log.Fields{
			"workspace": workspace.Name,
			"module":    "sync",
		}).Debugf("Writing history")

		err := s.History.Append(message)
		if err != nil {
			s.err = err
			return s
		}

		return s.Commit(message).Flush()
	} else {
		log.WithFields(log.Fields{
			"workspace": workspace.Name,
			"message":   message,
			"module":    "sync",
		}).Debugf("Not writing history. Sync module not loaded")
		return s
	}

}
func (s *Sync) Commit(message string) *Sync {
	if s.loaded {
		hash, err := GitCommitAll(s.Path, message)
		if err != nil {
			s.err = err
			return s
		}

		log.WithFields(log.Fields{
			"workspace": workspace.Name,
			"message":   message,
			"hash":      hash,
			"module":    "sync",
		}).Debugf("Added commit")
	} else {
		log.WithFields(log.Fields{
			"workspace": workspace.Name,
			"message":   message,
			"module":    "sync",
		}).Debugf("Not committing. Sync module not loaded")
	}
	return s
}
