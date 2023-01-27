package cmd

import (
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:    "sync",
	Short:  "Sync the workspace",
	Hidden: true,
	Long:   ``,
	Args:   cobra.ExactArgs(0), // https://github.com/spf13/cobra/blob/master/user_guide.md
	Run: func(cmd *cobra.Command, args []string) {

		workspace.load().Flush()

		//sync.Sync().Flush()

		//workspace.RunAction(args[0]).Flush()
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}

//type SyncProviderFactory func() PolycrateProvider

type SyncReference struct {
	Commit    string     `yaml:"commit,omitempty" mapstructure:"commit,omitempty" json:"commit,omitempty"`
	Branch    SyncBranch `yaml:"branch,omitempty" mapstructure:"branch,omitempty" json:"branch,omitempty"`
	Reference string     `yaml:"reference,omitempty" mapstructure:"reference,omitempty" json:"reference,omitempty"`
}

// type Sync struct {
// 	UUID uuid.UUID `yaml:"uuid,omitempty" mapstructure:"uuid,omitempty" json:"uuid,omitempty"`
// 	//Provider  PolycrateProvider `yaml:"provider,omitempty" mapstructure:"provider,omitempty" json:"provider,omitempty"`
// 	Path      string      `yaml:"path,omitempty" mapstructure:"path,omitempty" json:"path,omitempty"`
// 	Options   SyncOptions `yaml:"options,omitempty" mapstructure:"options,omitempty" json:"options,omitempty"`
// 	err       error
// 	loaded    bool
// 	Status    string        `yaml:"status,omitempty" mapstructure:"status,omitempty" json:"status,omitempty"`
// 	LocalRef  SyncReference `yaml:"local_ref,omitempty" mapstructure:"local_ref,omitempty" json:"local_ref,omitempty"`
// 	RemoteRef SyncReference `yaml:"remote_ref,omitempty" mapstructure:"remote_ref,omitempty" json:"remote_ref,omitempty"`
// 	History   HistoryLog    `yaml:"history,omitempty" mapstructure:"history,omitempty" json:"history,omitempty"`
// }

// func (s *Sync) Print() {
// 	printObject(s)
// }

// func (s *Sync) Sync() *Sync {
// 	if !s.loaded {
// 		s.err = fmt.Errorf("sync module not loaded")
// 		return s
// 	}

// 	if s.Options.Enabled {
// 		s.UpdateStatus().Flush()

// 		log.WithFields(log.Fields{
// 			"workspace": workspace.Name,
// 		}).Debugf("Syncing")

// 		// log.WithFields(log.Fields{
// 		// 	"workspace": workspace.Name,
// 		// 	"module":    "sync",
// 		// }).Debugf("Updating status")

// 		// s.UpdateStatus().Flush()

// 		switch status := s.Status; status {
// 		case "changed":
// 			log.WithFields(log.Fields{
// 				"workspace": workspace.Name,
// 			}).Debugf("Changes found. Comitting.")

// 			s.Commit("Sync auto-commit").Flush()
// 			s.Sync().Flush()
// 		case "synced":
// 			log.WithFields(log.Fields{
// 				"workspace": workspace.Name,
// 			}).Debugf("Up-to-date")
// 		case "diverged":
// 			// log.WithFields(log.Fields{
// 			// 	"workspace": workspace.Name,
// 			// 	"module":    "sync",
// 			// }).Fatalf("Sync error - run `polycrate sync status` for more information")
// 			s.err = fmt.Errorf("sync error - run `polycrate sync status` for more information")
// 		case "ahead":
// 			log.WithFields(log.Fields{
// 				"workspace": workspace.Name,
// 			}).Debugf("Ahead. Pushing")
// 			s.Push().Flush()
// 			s.Sync().Flush()
// 		case "behind":
// 			log.WithFields(log.Fields{
// 				"workspace": workspace.Name,
// 			}).Debugf("Behind. Pulling")
// 			s.Pull().Flush()
// 			s.Sync().Flush()
// 		}
// 	} else {
// 		s.err = fmt.Errorf("sync disabled")
// 	}

// 	return s
// }

// func (s *Sync) Push() *Sync {
// 	_, err := GitPush(s.Path, s.Options.Remote.Name, s.Options.Remote.Branch.Name)
// 	if err != nil {
// 		s.err = err
// 		return s
// 	}
// 	return s
// }

// func (s *Sync) Pull() *Sync {
// 	_, err := GitPull(s.Path, s.Options.Remote.Name, s.Options.Remote.Branch.Name)
// 	if err != nil {
// 		s.err = err
// 		return s
// 	}
// 	return s
// }

// func (s *Sync) Load() *Sync {
// 	if workspace.SyncOptions.Enabled {
// 		s.UUID = uuid.New()
// 		log.WithFields(log.Fields{
// 			"workspace": workspace.Name,
// 		}).Debugf("Loading sync module")

// 		//s.LoadProvider().Flush()
// 		s.LoadRepo().Flush()

// 		s.Options.Local.Branch.Name = workspace.SyncOptions.Local.Branch.Name
// 		s.Options.Remote.Branch.Name = workspace.SyncOptions.Remote.Branch.Name
// 		s.Options.Remote.Name = workspace.SyncOptions.Remote.Name
// 		s.Options.Remote.Url = workspace.SyncOptions.Remote.Url
// 		s.Options.Enabled = workspace.SyncOptions.Enabled
// 		s.Options.Auto = workspace.SyncOptions.Auto

// 		s.loaded = true

// 		// // pull if we're behind
// 		// switch status := s.Status; status {
// 		// case "behind":
// 		// 	log.WithFields(log.Fields{
// 		// 		"workspace": workspace.Name,
// 		// 		"module":    "sync",
// 		// 	}).Debugf("New remote commits found. Pulling")
// 		// 	s.Commit("Sync auto-commit").Flush()
// 		// 	s.Pull().Flush()
// 		// 	s.Status = "synced"
// 		// }
// 	} else {
// 		log.WithFields(log.Fields{
// 			"workspace": workspace.Name,
// 		}).Debugf("Not loading sync module. Sync disabled")

// 		s.loaded = false
// 	}
// 	return s
// }

// func (s *Sync) UpdateStatus() *Sync {
// 	if s.Options.Enabled {
// 		log.WithFields(log.Fields{
// 			"workspace": workspace.Name,
// 		}).Debugf("Getting remote status")
// 		// https://stackoverflow.com/posts/68187853/revisions
// 		// Get remote reference
// 		_, err := GitFetch(s.Path)
// 		if err != nil {
// 			s.err = err
// 			return s
// 		}

// 		s.LocalRef.Reference = s.Options.Local.Branch.Name
// 		s.RemoteRef.Reference = fmt.Sprintf("%s/%s", s.Options.Remote.Name, s.Options.Remote.Branch.Name)

// 		log.WithFields(log.Fields{
// 			"workspace": workspace.Name,
// 		}).Debugf("Getting last local commit")
// 		s.LocalRef.Branch.Name = s.Options.Local.Branch.Name
// 		s.LocalRef.Commit, err = GitGetHeadCommit(s.Path, s.LocalRef.Reference)
// 		if err != nil {
// 			s.err = err
// 			return s
// 		}

// 		log.WithFields(log.Fields{
// 			"workspace": workspace.Name,
// 		}).Debugf("Getting last remote commit")
// 		s.RemoteRef.Branch.Name = s.Options.Remote.Branch.Name
// 		s.RemoteRef.Commit, err = GitGetHeadCommit(s.Path, s.RemoteRef.Reference)
// 		if err != nil {
// 			s.err = err
// 			return s
// 		}

// 		log.WithFields(log.Fields{
// 			"workspace": workspace.Name,
// 		}).Debugf("Checking if behind remote")
// 		behindBy, err := GitBehindBy(s.Path)
// 		if err != nil {
// 			s.err = err
// 			return s
// 		}

// 		log.WithFields(log.Fields{
// 			"workspace": workspace.Name,
// 		}).Debugf("Checking if ahead of remote")
// 		aheadBy, err := GitAheadBy(s.Path)
// 		if err != nil {
// 			s.err = err
// 			return s
// 		}

// 		// ahead > 0, behind == 0
// 		if aheadBy != 0 && behindBy == 0 {
// 			s.Status = "ahead"
// 		}

// 		// ahead == 0, behind > 0
// 		if behindBy != 0 && aheadBy == 0 {
// 			s.Status = "behind"
// 		}

// 		// ahead == 0, behind == 0
// 		if behindBy == 0 && aheadBy == 0 {
// 			s.Status = "synced"
// 		}

// 		// ahead > 0, behind > 0
// 		if behindBy != 0 && aheadBy != 0 {
// 			s.Status = "diverged"
// 		}

// 		log.WithFields(log.Fields{
// 			"workspace": workspace.Name,
// 		}).Debugf("Checking for uncommited changes")

// 		// Has uncommited changes?
// 		if GitHasChanges(s.Path) {
// 			s.Status = "changed"
// 		}

// 		log.WithFields(log.Fields{
// 			"workspace": workspace.Name,
// 			"status":    s.Status,
// 			"ahead":     aheadBy,
// 			"behind":    behindBy,
// 		}).Debugf("Sync status")
// 	} else {
// 		log.WithFields(log.Fields{
// 			"workspace": workspace.Name,
// 			"status":    "disabled",
// 		}).Debugf("Sync status")
// 	}
// 	return s
// }

// func (s *Sync) LoadProvider() *Sync {
// 	provider, err := s.getProvider()
// 	if err != nil {
// 		s.err = err
// 		return s
// 	}
// 	s.Provider = provider
// 	return s
// }

// func (s *Sync) Flush() *Sync {
// 	if s.err != nil {
// 		log.Fatal(s.err)
// 	}
// 	return s
// }

// func (s *Sync) getProvider() (PolycrateProvider, error) {
// 	if config.Sync.Provider == "gitlab" {
// 		var pf SyncProviderFactory = NewGitlabSyncProvider
// 		provider := pf()

// 		log.WithFields(log.Fields{
// 			"provider": "gitlab",
// 			"path":     workspace.LocalPath,
// 		}).Debugf("Loading sync provider")
// 		return provider, nil
// 	}
// 	return nil, fmt.Errorf("provider not found: %s", config.Sync.Provider)
// }

// func (s *Sync) LoadRepo(path string) *Sync {

// 	var err error
// 	// Check if it's a git repo already
// 	log.WithFields(log.Fields{
// 		"path": path,
// 	}).Debugf("Loading local repository")

// 	if GitIsRepo(path) {
// 		// It's a git repo
// 		// 1. Get repo's remote
// 		// 2. Compare with configured remote
// 		// 2.1 No remote configured? Update configured remote with repo's remote
// 		// 2.2 No repo remot? Update with configured remote
// 		// 2.3 Unequal? Update repo remote with configured remote

// 		// Check remote
// 		remoteUrl, err := GitGetRemoteUrl(path, GitDefaultRemote)
// 		if err != nil {
// 			s.err = err
// 			return s
// 		}
// 		if remoteUrl == "" {
// 			log.WithFields(log.Fields{
// 				"path": path,
// 			}).Debugf("Local repository has no remote url configured")

// 			// Check if workspace has a remote url configured
// 			if workspace.SyncOptions.Remote.Url != "" {
// 				// Create the remote from the workspace config
// 				err := GitCreateRemote(path, GitDefaultRemote, workspace.SyncOptions.Remote.Url)
// 				if err != nil {
// 					s.err = err
// 					return s
// 				}
// 			} else {
// 				// Exit with error - workspace.SyncOptions.Remote.Url is not configured
// 				s.err = fmt.Errorf("workspace has no remote configured")
// 				return s
// 			}
// 		} else {
// 			// Remote is already configured
// 			// Check if workspace has a remote url configured
// 			if workspace.SyncOptions.Remote.Url != "" {
// 				// Check if its url matches the configured remote url

// 				if remoteUrl != workspace.SyncOptions.Remote.Url {
// 					// Urls don't match
// 					// Update the repository with the configured remote
// 					log.WithFields(log.Fields{
// 						"path":      workspace.LocalPath,
// 						"workspace": workspace.Name,
// 					}).Debugf("Local repository remote url doesn't match workspace remote url. Fixing.")

// 					err := GitUpdateRemoteUrl(workspace.LocalPath, GitDefaultRemote, workspace.SyncOptions.Remote.Url)
// 					if err != nil {
// 						s.err = err
// 						return s
// 					}
// 				}
// 			} else {
// 				// Update the workspace remote with the local remote
// 				log.WithFields(log.Fields{
// 					"path": path,
// 				}).Debugf("Workspace has no remote url configured. Updating with local repository remote url")
// 				log.WithFields(log.Fields{
// 					"path": path,
// 				}).Warnf("Updating workspace remote url with local repository remote url")
// 				workspace.updateConfig("sync.remote.url", remoteUrl).Flush()
// 			}
// 		}
// 		log.WithFields(log.Fields{
// 			"path":      workspace.LocalPath,
// 			"workspace": workspace.Name,
// 			"remote":    workspace.SyncOptions.Remote.Name,
// 			"branch":    workspace.SyncOptions.Remote.Branch.Name,
// 		}).Debugf("Tracking remote branch")
// 		_, err = GitSetUpstreamTracking(path, workspace.SyncOptions.Remote.Name, workspace.SyncOptions.Remote.Branch.Name)
// 		if err != nil {
// 			s.err = err
// 			return s
// 		}
// 	} else {
// 		// Not a git repo
// 		// Check if a remote url is configured

// 		if workspace.SyncOptions.Remote.Url != "" {
// 			// We have a remote url configured
// 			// Create a repository with the given url
// 			log.WithFields(log.Fields{
// 				"path": path,
// 				"url":  workspace.SyncOptions.Remote.Url,
// 			}).Debugf("Creating new repository with remote url from workspace config")

// 			err = GitCreateRepo(path, workspace.SyncOptions.Remote.Name, workspace.SyncOptions.Remote.Branch.Name, workspace.SyncOptions.Remote.Url)
// 			if err != nil {
// 				s.err = err
// 				return s
// 			}
// 		} else {
// 			// No remote url configured
// 			log.WithFields(log.Fields{
// 				"path": path,
// 			}).Warnf("Workspace has no remote url configured.")
// 			s.err = errors.New("cannot sync this repository. No remote configured in workspace or repository")
// 			return s
// 		}
// 	}

// 	log.WithFields(log.Fields{
// 		"path": path,
// 	}).Debugf("Local repository loaded")

// 	s.Path = workspace.LocalPath
// 	return s
// }

// func (s *Sync) Validate() {
// 	// Check if a remote is configured
// 	// if not, fail
// 	printObject(s)
// }

// func (s *Sync) Log(message string) *Sync {
// 	if s.loaded {
// 		log.WithFields(log.Fields{
// 			"workspace": workspace.Name,
// 			"module":    "sync",
// 		}).Debugf("Writing history")

// 		err := s.History.Append(message)
// 		if err != nil {
// 			s.err = err
// 			return s
// 		}

// 		return s.Commit(message).Flush()
// 	} else {
// 		log.WithFields(log.Fields{
// 			"workspace": workspace.Name,
// 			"message":   message,
// 			"module":    "sync",
// 		}).Debugf("Not writing history. Sync module not loaded")
// 		return s
// 	}

// }
// func (s *Sync) Commit(message string) *Sync {
// 	if s.loaded {
// 		hash, err := GitCommitAll(s.Path, message)
// 		if err != nil {
// 			s.err = err
// 			return s
// 		}

// 		log.WithFields(log.Fields{
// 			"workspace": workspace.Name,
// 			"message":   message,
// 			"hash":      hash,
// 			"module":    "sync",
// 		}).Debugf("Added commit")
// 	} else {
// 		log.WithFields(log.Fields{
// 			"workspace": workspace.Name,
// 			"message":   message,
// 			"module":    "sync",
// 		}).Debugf("Not committing. Sync module not loaded")
// 	}
// 	return s
// }
