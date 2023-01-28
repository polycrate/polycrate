package cmd

import (
	"context"
	goErrors "errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// type PolycrateSync struct {
// 	Provider PolycrateProvider
// 	Repo     *git.Repository
// 	//err      error
// }

// type PolycrateSyncConfig struct {
// 	CreateRepo      bool   `yaml:"create_repo,omitempty" mapstructure:"create_repo,omitempty" json:"create_repo,omitempty"`
// 	DeleteRepo      bool   `yaml:"delete_repo,omitempty" mapstructure:"delete_repo,omitempty" json:"delete_repo,omitempty"`
// 	AutoSync        bool   `yaml:"auto_sync,omitempty" mapstructure:"auto_sync,omitempty" json:"auto_sync,omitempty"`
// 	Mode            string `yaml:"mode,omitempty" mapstructure:"mode,omitempty" json:"mode,omitempty"`

// 	Provider        string `yaml:"provider,omitempty" mapstructure:"provider,omitempty" json:"provider,omitempty"`
// 	DefaultBranch   string `yaml:"default_branch,omitempty" mapstructure:"default_branch,omitempty" json:"default_branch,omitempty"`
// }

type ContextKey string

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

type PolycrateConfig struct {
	//Sync      PolycrateSyncConfig     `yaml:"sync,omitempty" mapstructure:"sync,omitempty" json:"sync,omitempty"`
	//Providers []PolycrateProvider     `yaml:"providers,omitempty" mapstructure:"providers,omitempty" json:"providers,omitempty"`
	//Gitlab   PolycrateGitlabProvider `yaml:"gitlab,omitempty" mapstructure:"gitlab,omitempty" json:"gitlab,omitempty"`
	Registry  Registry    `yaml:"registry,omitempty" mapstructure:"registry,omitempty" json:"registry,omitempty"`
	Sync      SyncOptions `yaml:"sync,omitempty" mapstructure:"sync,omitempty" json:"sync,omitempty"`
	Loglevel  int         `yaml:"loglevel,omitempty" mapstructure:"loglevel,omitempty" json:"loglevel,omitempty"`
	Logformat string      `yaml:"logformat,omitempty" mapstructure:"logformat,omitempty" json:"logformat,omitempty"`
	//Workspace PolycrateWorkspaceDefaults `yaml:"workspace,omitempty" mapstructure:"workspace,omitempty" json:"workspace,omitempty"`
}

type PolycrateTransaction struct {
	Context    context.Context
	TXID       uuid.UUID
	CancelFunc func()
}

type Polycrate struct {
	lock         sync.Mutex
	Config       PolycrateConfig `yaml:"config,omitempty" mapstructure:"config,omitempty" json:"config,omitempty"`
	Workspaces   []*Workspace
	Transactions []PolycrateTransaction
}

func (p *Polycrate) Load(ctx context.Context) error {
	var logrusLogLevel string
	switch polycrate.Config.Loglevel {
	case 1:
		logrusLogLevel = "Info"
	case 2:
		logrusLogLevel = "Debug"
	case 3:
		logrusLogLevel = "Trace"
	default:
		logrusLogLevel = "Info"
	}

	logrusLevel, err := log.ParseLevel(logrusLogLevel)
	if err != nil {
		logrusLevel = log.InfoLevel
	}

	// Set global log level
	log.SetLevel(logrusLevel)

	// Report the calling function if loglevel == 3
	if polycrate.Config.Loglevel == 3 {
		log.SetReportCaller(true)
	}

	// Set Formatter
	log.SetFormatter(&log.TextFormatter{})
	if p.Config.Logformat == "json" {
		log.SetFormatter(&log.JSONFormatter{})
	}

	// Register the custom validators to the global validator variable
	validate.RegisterValidation("metadata_name", validateMetadataName)
	validate.RegisterValidation("block_name", validateBlockName)

	if _, err := os.Stat(polycrateConfigFilePath); os.IsNotExist(err) {
		// Seems config wasn't found
		// Let's initialize it
		if err := p.CreateConfigDir(ctx); err != nil {
			return err
		}
		if err := p.CreateConfigFile(ctx); err != nil {
			return err
		}
	}

	log.Debugf("Loading Polycrate config from %s", polycrateConfigFilePath)
	if err := p.LoadConfigFromFile(ctx, polycrateConfigFilePath); err != nil {
		return err
	}

	// if err := p.LoadWorkspaces(); err != nil {
	// 	return err
	// }

	return nil
}

func (p *Polycrate) GetTransaction(TXID uuid.UUID) *PolycrateTransaction {
	for _, t := range polycrate.Transactions {
		if t.TXID == TXID {
			return &t
		}
	}
	return nil
}
func (p *Polycrate) DeleteTransaction(TXID uuid.UUID) error {
	deleted := false
	for i := len(p.Transactions) - 1; i >= 0; i-- {
		if p.Transactions[i].TXID == TXID {
			p.Transactions[i] = p.Transactions[len(p.Transactions)-1]
			deleted = true
		}
	}
	if deleted {
		return nil
	} else {
		return fmt.Errorf("unable to delete entry: not found")
	}
}

func (p *Polycrate) RegisterWorkspace(workspace *Workspace) error {
	// Lock
	p.lock.Lock()

	if workspace.LocalPath == "" {
		return fmt.Errorf("workspace path not set")
	}

	// Check if workspace exists
	if _, err := p.GetWorkspace(workspace.LocalPath); err == nil {
		return fmt.Errorf("workspace already registered: %s", workspace.LocalPath)
	}
	p.Workspaces = append(p.Workspaces, workspace)

	// Unlock
	p.lock.Unlock()

	return nil
}

func (p *Polycrate) RegisterTransaction(transaction PolycrateTransaction) error {
	// Lock
	p.lock.Lock()

	if transaction.TXID == uuid.Nil {
		return fmt.Errorf("no TXID found")
	}

	// Check if transaction exists
	if t := p.GetTransaction(transaction.TXID); t != nil {
		return fmt.Errorf("transaction already registered: %s", transaction.TXID)
	}
	p.Transactions = append(p.Transactions, transaction)

	// Unlock
	p.lock.Unlock()

	return nil
}
func (p *Polycrate) UnregisterTransaction(transaction *PolycrateTransaction) error {
	// Lock
	p.lock.Lock()

	if transaction.TXID == uuid.Nil {
		return fmt.Errorf("no TXID found")
	}

	if err := p.DeleteTransaction(transaction.TXID); err != nil {
		return err
	}

	// Unlock
	p.lock.Unlock()

	return nil
}

func (p *Polycrate) StartTransaction(ctx context.Context, cancelFunc func()) (context.Context, error) {
	TXIDKey := ContextKey("TXID")
	txid := uuid.New()

	ctx = context.WithValue(ctx, TXIDKey, txid)

	transaction := PolycrateTransaction{
		Context:    ctx,
		TXID:       txid,
		CancelFunc: cancelFunc,
	}

	// Register transaction aware logger
	log := log.WithField("txid", txid)
	ctx = p.SetContextLogger(ctx, log)

	err := p.RegisterTransaction(transaction)
	if err != nil {
		return nil, err
	}

	txRuntimeDir := filepath.Join([]string{polycrateRuntimeDir, txid.String()}...)
	err = os.MkdirAll(txRuntimeDir, os.ModePerm)
	if err != nil {
		return nil, err
	}

	log.Debug("Started transaction")

	return ctx, nil
}
func (p *Polycrate) StopTransaction(ctx context.Context, cancelFunc func()) error {
	txid := p.GetContextTXID(ctx)

	tx := p.GetTransaction(txid)

	log := p.GetContextLogger(ctx)
	log = log.WithField("txid", txid)

	// Call cancelFunc
	cancelFunc()

	// Unregister the transaction
	if err := p.UnregisterTransaction(tx); err != nil {
		return err
	}

	log.Debug("Stopped transaction")

	return nil
}

func (p *Polycrate) GetContextTXID(ctx context.Context) uuid.UUID {
	k := ContextKey("TXID")
	t := ctx.Value(k).(uuid.UUID)
	return t
}

func (p *Polycrate) GetContextLogger(ctx context.Context) *log.Entry {
	LoggerKey := ContextKey("Logger")
	log := ctx.Value(LoggerKey).(*log.Entry)
	return log
}

func (p *Polycrate) SetContextLogger(ctx context.Context, log *log.Entry) context.Context {
	LoggerKey := ContextKey("Logger")
	ctx = context.WithValue(ctx, LoggerKey, log)
	return ctx
}

func (p *Polycrate) LoadWorkspaces() error {
	// Discover local workspaces and load to localWorkspaceIndex
	err := p.DiscoverWorkspaces()
	if err != nil {
		return err
	}
	return nil
}

func (p *Polycrate) LoadConfigFromFile(ctx context.Context, path string) error {
	// 1. load $HOME/.polycrate/config.yml
	// 2. Unmarshal into polycrate.Config

	// Load Polycrate config
	var c = viper.New()

	// Match CLI Flags with Config options
	// CLI Flags have precedence
	c.BindPFlag("loglevel", rootCmd.Flags().Lookup("loglevel"))
	c.BindPFlag("registry.url", rootCmd.Flags().Lookup("registry-url"))
	c.BindPFlag("registry.base_image", rootCmd.Flags().Lookup("registry-base-image"))
	c.BindPFlag("workspace.config.image.version", rootCmd.Flags().Lookup("image-version"))
	c.BindPFlag("workspace.config.image.reference", rootCmd.Flags().Lookup("image-ref"))
	c.BindPFlag("workspace.config.blocksroot", rootCmd.Flags().Lookup("blocks-root"))
	c.BindPFlag("workspace.config.blocksconfig", rootCmd.Flags().Lookup("blocks-config"))
	c.BindPFlag("workspace.config.workflowsroot", rootCmd.Flags().Lookup("workflows-root"))
	c.BindPFlag("workspace.config.workspaceconfig", rootCmd.Flags().Lookup("workspace-config"))
	c.BindPFlag("workspace.config.artifactsroot", rootCmd.Flags().Lookup("artifacts-root"))
	c.BindPFlag("workspace.config.containerroot", rootCmd.Flags().Lookup("container-root"))
	c.BindPFlag("workspace.config.remoteroot", rootCmd.Flags().Lookup("remote-root"))
	c.BindPFlag("workspace.config.sshprivatekey", rootCmd.Flags().Lookup("ssh-private-key"))
	c.BindPFlag("workspace.config.sshpublickey", rootCmd.Flags().Lookup("ssh-public-key"))
	c.BindPFlag("workspace.config.dockerfile", rootCmd.Flags().Lookup("dockerfile"))
	c.BindPFlag("workspace.sync.local.branch.name", rootCmd.Flags().Lookup("sync-local-branch"))
	c.BindPFlag("workspace.sync.remote.branch.name", rootCmd.Flags().Lookup("sync-remote-branch"))
	c.BindPFlag("workspace.sync.remote.name", rootCmd.Flags().Lookup("sync-remote-name"))
	c.BindPFlag("workspace.sync.enabled", rootCmd.Flags().Lookup("sync-enabled"))
	c.BindPFlag("workspace.sync.auto", rootCmd.Flags().Lookup("sync-auto"))
	c.BindPFlag("workspace.extraenv", rootCmd.Flags().Lookup("env"))
	c.BindPFlag("workspace.extramounts", rootCmd.Flags().Lookup("mount"))
	c.BindPFlag("workspace.localpath", rootCmd.Flags().Lookup("workspace"))
	c.BindPFlag("workspace.containerpath", rootCmd.Flags().Lookup("container-root"))

	c.SetEnvPrefix(EnvPrefix)
	c.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	c.AutomaticEnv()

	c.SetConfigType("yaml")
	c.SetConfigFile(path)

	err := c.ReadInConfig()
	if err != nil {
		return err
	}

	if err = c.Unmarshal(&p.Config); err != nil {
		return err
	}

	if err = p.Config.validate(); err != nil {
		return err
	}

	return nil
}

func (p *Polycrate) CreateConfigDir(ctx context.Context) error {
	err := CreateDir(polycrateConfigDir)
	if err != nil {
		return err
	}
	return nil
}

func (p *Polycrate) CreateConfigFile(ctx context.Context) error {
	err := CreateFile(polycrateConfigFilePath)
	if err != nil {
		return err
	}
	return nil
}

func (p *Polycrate) ContextExit(ctx context.Context, cancelFunc context.CancelFunc, err error) {
	if err := polycrate.StopTransaction(ctx, cancelFunc); err != nil {
		log.Fatal(err)
	}

	if err != nil {
		log.Fatal(err)
	}
	return
}

func (p *Polycrate) InitWorkspace(ctx context.Context, path string, name string, withSSHKeys bool, withConfig bool) (*Workspace, error) {
	log := p.GetContextLogger(ctx)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.WithField("path", path)
		log.Info("Creating directory")

		if err := CreateDir(path); err != nil {
			return nil, err
		}
	}

	workspace := new(Workspace)

	// Make a hard copy of the defaultWorkspace
	*workspace = defaultWorkspace

	workspace.Name = name
	workspace.LocalPath = path
	workspace.ContainerPath = workspace.Config.ContainerRoot

	blocksDir := filepath.Join([]string{workspace.LocalPath, workspace.Config.BlocksRoot}...)
	if _, err := os.Stat(blocksDir); os.IsNotExist(err) {
		log.WithField("path", blocksDir)
		log.Info("Creating directory")

		if err := CreateDir(blocksDir); err != nil {
			return nil, err
		}
	}

	log.WithField("workspace", workspace.Name)

	if withConfig {
		log.WithField("config", workspace.Config.WorkspaceConfig)
		log.Info("Saving config")
		if err := workspace.Save(ctx); err != nil {
			return nil, err
		}
	}

	if withSSHKeys {
		err := workspace.CreateSshKeys(ctx)
		if err != nil {
			return nil, err
		}
	}

	return workspace, nil
}

func (p *Polycrate) CleanupRuntimeDir(ctx context.Context) error {
	log := p.GetContextLogger(ctx)

	err := os.MkdirAll(polycrateRuntimeDir, os.ModePerm)
	if err != nil {
		return err
	}

	// Create runtime dir
	log.WithField("path", polycrateRuntimeDir).Debugf("Cleaning runtime directory")

	// Purge all contents of runtime dir
	dir, err := ioutil.ReadDir(polycrateRuntimeDir)
	if err != nil {
		return err
	}
	for _, d := range dir {
		log.WithField("file", d.Name()).WithField("path", polycrateRuntimeDir).Debugf("Removing directory")
		err := os.RemoveAll(filepath.Join([]string{polycrateRuntimeDir, d.Name()}...))
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Polycrate) LoadWorkspace(ctx context.Context, path string) (*Workspace, error) {
	// 1. Check if its in polycrate.Workspaces via GetWorkspace
	// 2. If yes, load it from there, lock Polycrate, run workspace.Load(), and unlock
	// 3. If not, bootstrap from defaultWorkspace and add to polycrate.Workspaces
	var workspace *Workspace
	var err error
	log := p.GetContextLogger(ctx)
	log = log.WithField("path", path)

	workspace, err = p.GetWorkspace(path)
	if err != nil {
		log.Debugf("Loading workspace")

		workspace = new(Workspace)

		// Make a hard copy of the defaultWorkspace
		*workspace = defaultWorkspace

		// Set the path to the given path
		workspace, err = workspace.Load(ctx, path)
		if err != nil {
			return nil, err
		}
	} else {
		// Reload workspace
		log.Debugf("Reloading workspace")

		//workspace.Reload(ctx)
		workspace, err = workspace.Reload(ctx)
		if err != nil {
			return nil, err
		}
	}
	//w.Load(ctx, path)

	//w.load().Flush()
	log.WithField("workspace", workspace.Name).WithField("blocks", len(workspace.Blocks)).WithField("workflows", len(workspace.Workflows)).Debugf("Workspace loaded")

	// RegisterWorkspace
	if err = p.RegisterWorkspace(workspace); err != nil {
		return nil, err
	}

	return workspace, nil
}

func (p *Polycrate) DiscoverWorkspaces() error {
	workspacesDir := polycrateWorkspaceDir

	if _, err := os.Stat(workspacesDir); !os.IsNotExist(err) {
		log.WithFields(log.Fields{
			"path": workspacesDir,
		}).Debugf("Discovering local workspaces")

		// This function adds all valid Blocks to the list of
		err := filepath.WalkDir(workspacesDir, p.WalkWorkspacesDir)
		if err != nil {
			return err
		}
	} else {
		log.WithFields(log.Fields{
			"path": workspacesDir,
		}).Debugf("Skipping workspace discovery. Local workspaces directory not found")
	}

	return nil
}

// This functions looks for a specific workspace in polycrate.Workspaces
func (p *Polycrate) GetWorkspace(id string) (*Workspace, error) {

	for i := 0; i < len(p.Workspaces); i++ {

		workspace := p.Workspaces[i]
		//workspace := workspace
		if workspace != nil {
			log.Debug(workspace.Name)

			if workspace.Name == id || workspace.LocalPath == id {
				return workspace, nil
			}
		}
	}

	err := fmt.Errorf("workspace not found in index")
	return nil, err
}

func (p *Polycrate) WalkWorkspacesDir(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}

	if !d.IsDir() {
		fileinfo, _ := d.Info()

		if fileinfo.Name() == WorkspaceConfigFile {
			workspaceConfigFileDir := filepath.Dir(path)
			log.WithFields(log.Fields{
				"path": workspaceConfigFileDir,
			}).Tracef("Local workspace detected")

			w := Workspace{}
			w.LocalPath = workspaceConfigFileDir
			w.Path = workspaceConfigFileDir

			log.WithFields(log.Fields{
				"path": w.Path,
			}).Tracef("Reading in local workspace")

			w.SoftloadWorkspaceConfig().Flush()

			// Check if the workspace has already been loaded to the local workspace index
			_i, err := p.GetWorkspace(w.Name)
			if err != nil {
				// Workspace not indexed
				p.Workspaces = append(p.Workspaces, _i)
			}
		}
	}
	return nil
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

func (p *Polycrate) getTempFile(ctx context.Context, filename string) (*os.File, error) {
	txid := polycrate.GetContextTXID(ctx)

	fp := filepath.Join(polycrateRuntimeDir, txid.String(), filename)
	f, err := os.Create(fp)
	if err != nil {
		return nil, err
	}
	return f, nil
}
