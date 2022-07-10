package cmd

import (
	goErrors "errors"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
)

type PolycrateSync struct {
	Provider PolycrateProvider
	Repo     *git.Repository
	err      error
}

type PolycrateSyncConfig struct {
	CreateRepo      bool   `yaml:"create_repo,omitempty" mapstructure:"create_repo,omitempty" json:"create_repo,omitempty"`
	DeleteRepo      bool   `yaml:"delete_repo,omitempty" mapstructure:"delete_repo,omitempty" json:"delete_repo,omitempty"`
	AutoSync        bool   `yaml:"auto_sync,omitempty" mapstructure:"auto_sync,omitempty" json:"auto_sync,omitempty"`
	Mode            string `yaml:"mode,omitempty" mapstructure:"mode,omitempty" json:"mode,omitempty"`
	DefaultProvider string `yaml:"default_provider,omitempty" mapstructure:"default_provider,omitempty" json:"default_provider,omitempty"`
	Provider        string `yaml:"provider,omitempty" mapstructure:"provider,omitempty" json:"provider,omitempty"`
	DefaultBranch   string `yaml:"default_branch,omitempty" mapstructure:"default_branch,omitempty" json:"default_branch,omitempty"`
}

type PolycrateProviderProject struct {
	url         string
	name        string
	id          int
	remote_ssh  string
	remote_http string
	path        string
}

type PolycrateProviderGroup struct {
	url  string
	name string
	path string
	id   int
}

type PolycrateProviderCredentials struct {
	username string
	password string
}

type PolycrateProvider interface {
	GetCredentials() (PolycrateProviderCredentials, error)
	Print()
	GetName() string
	GetDefaultGroup() (PolycrateProviderGroup, error)
	CreateProject(group PolycrateProviderGroup, name string) (PolycrateProviderProject, error)
}

type PolycrateConfig struct {
	Sync      PolycrateSyncConfig     `yaml:"sync,omitempty" mapstructure:"sync,omitempty" json:"sync,omitempty"`
	Providers []PolycrateProvider     `yaml:"providers,omitempty" mapstructure:"providers,omitempty" json:"providers,omitempty"`
	Gitlab    PolycrateGitlabProvider `yaml:"gitlab,omitempty" mapstructure:"gitlab,omitempty" json:"gitlab,omitempty"`
}

func (c *PolycrateProviderCredentials) Print() {
	printObject(c)
}

func (p *PolycrateConfig) validate() error {
	err := validate.Struct(p)

	if err != nil {

		// this check is only needed when your code could produce
		// an invalid value for validation such as interface with nil
		// value most including myself do not usually have code like this.
		if _, ok := err.(*validator.InvalidValidationError); ok {
			log.Error(err)
			return nil
		}

		for _, err := range err.(validator.ValidationErrors) {
			log.WithFields(log.Fields{
				"option": strings.ToLower(err.Namespace()),
				"error":  err.Tag(),
			}).Errorf("Validation error")
		}

		// from here you can create your own error messages in whatever language you wish
		return goErrors.New("error validating Polycrate config")
	}
	return nil
}
