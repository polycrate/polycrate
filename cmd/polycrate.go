package cmd

import (
	"bytes"
	"context"
	"os/signal"
	"syscall"
	"time"

	// "encoding/json"
	goErrors "errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"sync"

	"github.com/Songmu/prompter"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	semver "github.com/hashicorp/go-version"
	"github.com/imdario/mergo"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v2"
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

type ExperimentalConfig struct {
	MergeV2 bool `yaml:"merge_v2,omitempty" mapstructure:"merge_v2,omitempty" json:"merge_v2,omitempty"`
}

type PolycrateConfig struct {
	//Sync      PolycrateSyncConfig     `yaml:"sync,omitempty" mapstructure:"sync,omitempty" json:"sync,omitempty"`
	//Providers []PolycrateProvider     `yaml:"providers,omitempty" mapstructure:"providers,omitempty" json:"providers,omitempty"`
	//Gitlab   PolycrateGitlabProvider `yaml:"gitlab,omitempty" mapstructure:"gitlab,omitempty" json:"gitlab,omitempty"`
	Registry     Registry           `yaml:"registry,omitempty" mapstructure:"registry,omitempty" json:"registry,omitempty"`
	Sync         SyncOptions        `yaml:"sync,omitempty" mapstructure:"sync,omitempty" json:"sync,omitempty"`
	Loglevel     int                `yaml:"loglevel,omitempty" mapstructure:"loglevel,omitempty" json:"loglevel,omitempty"`
	Logformat    string             `yaml:"logformat,omitempty" mapstructure:"logformat,omitempty" json:"logformat,omitempty"`
	Kubeconfig   string             `yaml:"kubeconfig,omitempty" mapstructure:"kubeconfig,omitempty" json:"kubeconfig,omitempty"`
	Webhooks     []Webhook          `yaml:"webhooks,omitempty" mapstructure:"webhooks,omitempty" json:"webhooks,omitempty"`
	CheckUpdates bool               `yaml:"check_updates,omitempty" mapstructure:"check_updates,omitempty" json:"check_updates,omitempty"`
	Experimental ExperimentalConfig `yaml:"experimental,omitempty" mapstructure:"experimental,omitempty" json:"experimental,omitempty"`
	//Workspace PolycrateWorkspaceDefaults `yaml:"workspace,omitempty" mapstructure:"workspace,omitempty" json:"workspace,omitempty"`
}

type PolycrateEvent struct {
	Labels      map[string]string    `yaml:"labels,omitempty" mapstructure:"labels,omitempty" json:"labels,omitempty"`
	Data        interface{}          `yaml:"data,omitempty" mapstructure:"data,omitempty" json:"data,omitempty"`
	Workspace   string               `yaml:"workspace,omitempty" mapstructure:"workspace,omitempty" json:"workspace,omitempty"`
	Workflow    string               `yaml:"workflow,omitempty" mapstructure:"workflow,omitempty" json:"workflow,omitempty"`
	Step        string               `yaml:"step,omitempty" mapstructure:"step,omitempty" json:"step,omitempty"`
	Block       string               `yaml:"block,omitempty" mapstructure:"block,omitempty" json:"block,omitempty"`
	Action      string               `yaml:"action,omitempty" mapstructure:"action,omitempty" json:"action,omitempty"`
	Command     string               `yaml:"command,omitempty" mapstructure:"command,omitempty" json:"command,omitempty"`
	ExitCode    int                  `yaml:"exit_code,omitempty" mapstructure:"exit_code,omitempty" json:"exit_code,omitempty"`
	UserEmail   string               `yaml:"user_email,omitempty" mapstructure:"user_email,omitempty" json:"user_email,omitempty"`
	UserName    string               `yaml:"user_name,omitempty" mapstructure:"user_name,omitempty" json:"user_name,omitempty"`
	Date        string               `yaml:"date,omitempty" mapstructure:"date,omitempty" json:"date,omitempty"`
	Transaction string               `yaml:"transaction,omitempty" mapstructure:"transaction,omitempty" json:"transaction,omitempty"`
	Version     string               `yaml:"version,omitempty" mapstructure:"version,omitempty" json:"version,omitempty"`
	Output      string               `yaml:"output,omitempty" mapstructure:"output,omitempty" json:"output,omitempty"`
	Config      WorkspaceEventConfig `yaml:"config,omitempty" mapstructure:"config,omitempty" json:"config,omitempty"`
	Message     string               `yaml:"message,omitempty" mapstructure:"message,omitempty" json:"message,omitempty"`
}

type Webhook struct {
	Endpoint string            `yaml:"endpoint,omitempty" mapstructure:"endpoint,omitempty" json:"endpoint,omitempty"`
	Labels   map[string]string `yaml:"labels,omitempty" mapstructure:"labels,omitempty" json:"labels,omitempty"`
}

type PolycrateLog struct {
	log     *log.Entry
	Level   string
	History string
}

type PolycrateTransactionTask struct {
	Name string
	Job  func(tx *PolycrateTransaction) error
	// One of:
	// - created
	// - running
	// - stopped
	// - failed
	// - killed
	Status string
}

type PolycrateTransaction struct {
	Context     context.Context
	TXID        uuid.UUID `yaml:"txid,omitempty" mapstructure:"txid,omitempty" json:"txid,omitempty"`
	CancelFunc  func()
	Log         PolycrateLog
	RuntimeDir  string            `yaml:"runtime_dir,omitempty" mapstructure:"runtime_dir,omitempty" json:"runtime_dir,omitempty"`
	Command     string            `yaml:"command,omitempty" mapstructure:"command,omitempty" json:"command,omitempty"`
	UserEmail   string            `yaml:"user_email,omitempty" mapstructure:"user_email,omitempty" json:"user_email,omitempty"`
	UserName    string            `yaml:"user_name,omitempty" mapstructure:"user_name,omitempty" json:"user_name,omitempty"`
	Date        string            `yaml:"date,omitempty" mapstructure:"date,omitempty" json:"date,omitempty"`
	Transaction uuid.UUID         `yaml:"transaction,omitempty" mapstructure:"transaction,omitempty" json:"transaction,omitempty"`
	Version     string            `yaml:"version,omitempty" mapstructure:"version,omitempty" json:"version,omitempty"`
	ExitCode    int               `yaml:"exit_code,omitempty" mapstructure:"exit_code,omitempty" json:"exit_code,omitempty"`
	Output      string            `yaml:"output,omitempty" mapstructure:"output,omitempty" json:"output,omitempty"`
	Snapshot    WorkspaceSnapshot `yaml:"snapshot,omitempty" mapstructure:"snapshot,omitempty" json:"snapshot,omitempty"`
	Job         func(tx *PolycrateTransaction) error
	Tasks       []*PolycrateTransactionTask
	// One of:
	// - created
	// - running
	// - stopped
	// - failed
	// - killed
	Status string
}

type Polycrate struct {
	lock         sync.Mutex
	Config       PolycrateConfig `yaml:"config,omitempty" mapstructure:"config,omitempty" json:"config,omitempty"`
	Workspaces   []*Workspace
	Transactions []*PolycrateTransaction
}

type Prompt struct {
	Message string `yaml:"message,omitempty" mapstructure:"message,omitempty" json:"message,omitempty"`
}

func (p *Prompt) Validate() bool {
	return force || prompter.YN(p.Message, false)

}

// func (p *Polycrate) GetTransaction(txid uuid.UUID) (*PolycrateTransaction, error) {
// 	for i := 0; i < len(p.Transactions); i++ {
// 		transaction := &p.Transactions[i]
// 		if transaction.TXID == txid {
// 			return transaction, nil
// 		}
// 	}
// 	return nil, fmt.Errorf("transaction not found: %s", txid.String())
// }

func NewEvent(tx *PolycrateTransaction) *PolycrateEvent {
	event := &PolycrateEvent{
		Command:     tx.Command,
		Version:     tx.Version,
		Transaction: tx.TXID.String(),
		Labels: map[string]string{
			"monk.event.class": "polycrate",
		},
		UserEmail: tx.UserEmail,
		UserName:  tx.UserName,
		Date:      tx.Date,
		Output:    tx.Output,
	}

	if tx.Snapshot.Workspace != nil {
		event.Workspace = tx.Snapshot.Workspace.Name
		event.Config = tx.Snapshot.Workspace.Events
	}
	if tx.Snapshot.Block != nil {
		event.Block = tx.Snapshot.Block.Name
	}
	if tx.Snapshot.Action != nil {
		event.Action = tx.Snapshot.Action.Name
	}
	if tx.Snapshot.Workflow != nil {
		event.Workflow = tx.Snapshot.Workflow.Name
	}
	if tx.Snapshot.Step != nil {
		event.Step = tx.Snapshot.Step.Name
	}

	switch tx.Status {
	case "created":
		event.Labels["monk.event.level"] = "Warning"
		event.Message = "Transaction created but not finished"
	case "running":
		event.Labels["monk.event.level"] = "Warning"
		event.Message = "Transaction running but not finished"
	case "stopped":
		event.Labels["monk.event.level"] = "Info"
		event.Message = "Transaction stopped"
	case "failed":
		event.Labels["monk.event.level"] = "Error"
		event.Message = "Transaction failed"
	case "killed":
		event.Labels["monk.event.level"] = "Error"
		event.Message = "Transaction killed"
	}
	return event
}

func (e *PolycrateEvent) ToJSON() ([]byte, error) {
	eventData, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	return eventData, nil
}

func (e *PolycrateEvent) ToYAML() ([]byte, error) {
	eventData, err := yaml.Marshal(e)
	if err != nil {
		return nil, err
	}
	return eventData, nil
}

func (e *PolycrateEvent) MergeInLabels(labels map[string]string) error {
	_labels := &e.Labels
	if err := mergo.Merge(_labels, labels); err != nil {
		return err
	}
	return nil
}

func (e *PolycrateEvent) Save(tx *PolycrateTransaction) error {
	workspace = *tx.Snapshot.Workspace

	date_formatted := time.Now().Format(WorkspaceLogDateFormat)
	logDir := filepath.Join(workspace.LocalPath, workspace.Config.LogsRoot, date_formatted)

	tx.Log.Debugf("Preparing log directory at %s", logDir)

	err := os.MkdirAll(logDir, os.ModePerm)
	if err != nil {
		return err
	}

	logFile := strings.Join([]string{tx.TXID.String(), "yml"}, ".")
	logPath := filepath.Join(logDir, logFile)
	tx.Log.Debugf("Saving log file at %s", logPath)

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	// Export revision to yaml
	yaml, err := yaml.Marshal(e)
	if err != nil {
		return err
	}

	// Write yaml export to file
	_, err = f.Write(yaml)
	if err != nil {
		return err
	}

	// Check if git repo; if yes, commit
	if GitIsRepo(tx, workspace.LocalPath) {
		// Check if committing is enabled in workspace config
		if workspace.Events.Commit {
			_, err = GitCommitAll(tx, workspace.LocalPath, fmt.Sprintf("chore: saved event %s", tx.TXID.String()))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (e *PolycrateEvent) Submit(tx *PolycrateTransaction) error {
	if e.Config.Handler == "workspace" {
		tx.Log.Debugf("Saving event to workspace logs")
		e.Save(tx)
	} else if e.Config.Handler == "webhook" {

		tx.Log.Debugf("Submitting event to webhook at '%s'", e.Config.Endpoint)

		eventData, err := e.ToJSON()
		if err != nil {
			return err
		}
		request, err := http.NewRequest("POST", e.Config.Endpoint, bytes.NewBuffer(eventData))
		if err != nil {
			return err
		}
		request.Header.Set("Content-Type", "application/json; charset=UTF-8")

		client := &http.Client{}
		response, err := client.Do(request)
		if err != nil {
			return err
		}
		defer response.Body.Close()

		body, _ := ioutil.ReadAll(response.Body)
		log.Tracef("Webhook returned status '%s'", response.Status)
		log.Tracef(string(body))
	} else {
		log.Debugf("No valid event handler found.")
	}
	return nil
}

// type PolycrateTransactionFn func(context.Context, string) error

// func (p *Polycrate) WithTransaction(fn PolycrateTransactionFn, s string) error {
// 	ctx, cancelFunc := context.WithCancel(context.Background())
// 	ctx, _, err := p.StartTransaction(ctx, cancelFunc)
// 	if err != nil {
// 		return err
// 	}

// 	// Call function
// 	if err := fn(ctx, s); err != nil {
// 		if err := polycrate.StopTransaction(ctx, cancelFunc); err != nil {
// 			return err
// 		}
// 		return err
// 	}

// 	// Stop transaction
// 	if err := polycrate.StopTransaction(ctx, cancelFunc); err != nil {
// 		return err
// 	}

// 	return nil

// 	//log := polycrate.GetContextLogger(ctx)
// }

func (p *Polycrate) EventHandler(tx *PolycrateTransaction) error {
	// Exit if there's no event in the context
	event := NewEvent(tx)

	eg := new(errgroup.Group)

	eg.Go(func() error {
		if err := event.Submit(tx); err != nil {
			return err
		}
		// for _, webhook := range p.Config.Webhooks {
		// 	if err := event.MergeInLabels(webhook.Labels); err != nil {
		// 		return err
		// 	}

		// 	if err := event.Submit(ctx, webhook); err != nil {
		// 		return err
		// 	}
		// }
		return nil
	})
	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

func (p *Polycrate) GetUserInfo() map[string]string {
	email, _ := GitGetUserEmail()
	name, _ := GitGetUserName()
	user := map[string]string{
		"email": email,
		"name":  name,
	}

	return user
}

// func (p *Polycrate) PublishEvent(ctx context.Context, data *WorkspaceRevision, webhook Webhook) error {
// 	//log := polycrate.GetContextLogger(ctx)

// 	event := PolycrateEvent{}
// 	event.Data = data

// 	if err := event.Submit(ctx); err != nil {
// 		return err
// 	}

// 	return nil
// }

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
	// if polycrate.Config.Loglevel == 3 {
	// 	log.SetReportCaller(true)
	// }

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
	p.WaitForGracefulShutdown()

	return nil
}

func (p *Polycrate) GetTransaction(TXID uuid.UUID) *PolycrateTransaction {
	for _, t := range polycrate.Transactions {
		if t.TXID == TXID {
			return t
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

// func (p *Polycrate) UpdateTransaction(txid uuid.UUID, ctx context.Context) error {
// 	// Lock
// 	p.lock.Lock()

// 	if txid == uuid.Nil {
// 		return fmt.Errorf("no TXID found")
// 	}

// 	// Check if transaction exists
// 	t := p.GetTransaction(txid)
// 	if t == nil {
// 		return fmt.Errorf("transaction not registered: %s", txid)
// 	}

// 	err := t.Update(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	// Unlock
// 	p.lock.Unlock()

// 	return nil
// }

func (p *Polycrate) RegisterTransaction(transaction *PolycrateTransaction) {
	// Lock
	p.lock.Lock()

	if transaction.TXID == uuid.Nil {
		panic("no TXID found")
	}

	// Check if transaction exists
	if t := p.GetTransaction(transaction.TXID); t == nil {
		p.Transactions = append(p.Transactions, transaction)
	}

	// Unlock
	p.lock.Unlock()
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

func (p *Polycrate) Transaction() *PolycrateTransaction {
	ctx, cancel := context.WithCancel(context.Background())
	txid := uuid.New()

	TXIDKey := ContextKey("TXID")
	ctx = context.WithValue(ctx, TXIDKey, txid)

	versionKey := ContextKey("version")
	ctx = context.WithValue(ctx, versionKey, version)

	// Load Transaction Log
	pl := PolycrateLog{}
	pl.Load(ctx)
	// Set the current transaction id as a field of the log
	pl.SetField("txid", txid.String())

	// Prepare Transaction runtime dir
	txRuntimeDir := filepath.Join([]string{polycrateRuntimeDir, txid.String()}...)
	err := os.MkdirAll(txRuntimeDir, os.ModePerm)
	if err != nil {
		panic(err)
	}

	// Obtain user info
	userInfo := polycrate.GetUserInfo()

	// Create Transaction
	tx := PolycrateTransaction{
		Context:    ctx,
		TXID:       txid,
		CancelFunc: cancel,
		Log:        pl,
		RuntimeDir: txRuntimeDir,
		Date:       time.Now().Format(time.RFC3339),
		Version:    version,
		UserEmail:  userInfo["email"],
		UserName:   userInfo["name"],
	}

	// Register Transaction in index
	p.RegisterTransaction(&tx)

	return &tx
}

func (tx *PolycrateTransaction) SetCommand(cmd *cobra.Command) *PolycrateTransaction {
	tx.Command = FormatCommand(cmd)
	return tx
}
func (tx *PolycrateTransaction) SetOutput(output string) *PolycrateTransaction {
	tx.Output = output
	return tx
}
func (tx *PolycrateTransaction) SetExitCode(exitCode int) *PolycrateTransaction {
	tx.ExitCode = exitCode
	return tx
}
func (tx *PolycrateTransaction) SetJob(job func(*PolycrateTransaction) error) *PolycrateTransaction {
	tx.Job = job
	return tx
}

func (tx *PolycrateTransaction) Stop() *PolycrateTransaction {
	tx.Log.Debug("Stopping transaction")
	//log = log.WithField("txid", txid)

	// Call cancelFunc
	defer tx.CancelFunc()

	// Unregister the transaction
	if err := polycrate.UnregisterTransaction(tx); err != nil {
		panic(err)
	}

	tx.Status = "stopped"

	if err := polycrate.EventHandler(tx); err != nil {
		// We're not terminating here to not block further execution
		tx.Log.Warnf("Event handler failed: %s", err)
	}

	tx.Log.Debug("Stopped transaction")
	return tx
}
func (tx *PolycrateTransaction) Run() (err error) {
	if tx.Job != nil {
		return tx.Job(tx)
	}

	return fmt.Errorf("no job defined")
}

// func (p *Polycrate) NewTransaction(ctx context.Context, cmd *cobra.Command) (context.Context, *PolycrateTransaction, context.CancelFunc, error) {
// 	ctx, cancel := context.WithCancel(ctx)
// 	ctx, tx, err := p.StartTransaction(ctx, cancel)
// 	if err != nil {
// 		return ctx, nil, nil, err
// 	}

// 	revision := &WorkspaceRevision{}

// 	revision.Date = time.Now().Format(time.RFC3339)
// 	revision.Transaction = p.GetContextTXID(ctx)
// 	revision.Version = p.GetContextVersion(ctx)

// 	if cmd != nil {
// 		cmdKey := ContextKey("cmd")
// 		ctx = context.WithValue(ctx, cmdKey, cmd)
// 		revision.Command = FormatCommand(cmd)
// 	}

// 	userInfo := polycrate.GetUserInfo()
// 	revision.UserEmail = userInfo["email"]
// 	revision.UserName = userInfo["name"]

// 	revisionKey := ContextKey("revision")
// 	ctx = context.WithValue(ctx, revisionKey, revision)

// 	eventKey := ContextKey("event")
// 	ctx = context.WithValue(ctx, eventKey, NewEvent(tx))

// 	return ctx, tx, cancel, err
// }

// func (t *PolycrateTransaction) Update(ctx context.Context) error {
// 	t.Context = ctx
// 	return nil
// }

// func (p *Polycrate) StartTransaction(ctx context.Context, cancelFunc func()) (context.Context, *PolycrateTransaction, error) {
// 	TXIDKey := ContextKey("TXID")
// 	txid := uuid.New()
// 	versionKey := ContextKey("version")

// 	ctx = context.WithValue(ctx, TXIDKey, txid)
// 	ctx = context.WithValue(ctx, versionKey, version)

// 	transaction := PolycrateTransaction{
// 		Context:    ctx,
// 		TXID:       txid,
// 		CancelFunc: cancelFunc,
// 	}

// 	p.RegisterTransaction(&transaction)

// 	// Register transaction aware logger
// 	log := log.WithField("txid", txid)
// 	ctx = p.SetContextLogger(ctx, log)

// 	txRuntimeDir := filepath.Join([]string{polycrateRuntimeDir, txid.String()}...)
// 	err := os.MkdirAll(txRuntimeDir, os.ModePerm)
// 	if err != nil {
// 		return nil, nil, err
// 	}

// 	log.Debug("Started transaction")

// 	return ctx, &transaction, nil
// }

// func (p *Polycrate) StopTransaction(ctx context.Context, cancel func()) error {
// 	txid := p.GetContextTXID(ctx)

// 	tx := p.GetTransaction(txid)

// 	// Update with latest context
// 	ctx = tx.Context

// 	log.Debug("Stopping transaction")
// 	//log = log.WithField("txid", txid)

// 	// Call cancelFunc
// 	defer cancel()

// 	// Unregister the transaction
// 	if err := p.UnregisterTransaction(tx); err != nil {
// 		return err
// 	}

// 	if err := polycrate.EventHandler(ctx); err != nil {
// 		// We're not terminating here to not block further execution
// 		log.Warnf("Event handler failed: %s", err)
// 	}

// 	log.Debug("Stopped transaction")

// 	return nil
// }

func (p *Polycrate) GetContextTXID(ctx context.Context) uuid.UUID {
	k := ContextKey("TXID")
	t := ctx.Value(k).(uuid.UUID)
	return t
}

// func (p *Polycrate) LoadWorkspaces() error {
// 	// Discover local workspaces and load to localWorkspaceIndex
// 	err := p.DiscoverWorkspaces()
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

func (p *Polycrate) LoadConfigFromFile(ctx context.Context, path string) error {
	// 1. load $HOME/.polycrate/config.yml
	// 2. Unmarshal into polycrate.Config

	// Load Polycrate config
	var c = viper.New()

	// Match CLI Flags with Config options
	// CLI Flags have precedence
	c.BindPFlag("loglevel", rootCmd.Flags().Lookup("loglevel"))
	c.BindPFlag("kubeconfig", rootCmd.Flags().Lookup("kubeconfig"))
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

// func (p *Polycrate) ContextExit(ctx context.Context, cancelFunc context.CancelFunc, err error) {
// 	if err := p.StopTransaction(ctx, cancelFunc); err != nil {
// 		log.Fatal(err)
// 	}

// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	return
// }

func (p *Polycrate) InitWorkspace(tx *PolycrateTransaction, path string, name string, withSSHKeys bool, withConfig bool) (*Workspace, error) {

	if _, err := os.Stat(path); os.IsNotExist(err) {
		tx.Log.Debugf("Creating workspace directory at %s", path)

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
		tx.Log.Debugf("Creating blocks directory at %s", blocksDir)

		if err := CreateDir(blocksDir); err != nil {
			return nil, err
		}
	}

	if withConfig {
		if err := workspace.Save(tx); err != nil {
			log.Debugf("Config already exists at %s", workspace.Config.WorkspaceConfig)
		} else {
			log.Infof("Config created at %s", workspace.Config.WorkspaceConfig)
		}
	}

	if withSSHKeys {
		err := workspace.CreateSshKeys(tx.Context)
		if err != nil {
			log.Debug(err)
		}
	}

	return workspace, nil
}

func (p *Polycrate) CleanupRuntimeDir(tx *PolycrateTransaction) error {

	err := os.MkdirAll(polycrateRuntimeDir, os.ModePerm)
	if err != nil {
		return err
	}

	// Create runtime dir
	tx.Log.Debugf("Cleaning runtime directory at %s", polycrateRuntimeDir)

	// Purge all contents of runtime dir
	dir, err := ioutil.ReadDir(polycrateRuntimeDir)
	if err != nil {
		return err
	}
	for _, d := range dir {
		tx.Log.Debugf("Removing directory at %s/%s", polycrateRuntimeDir, d.Name())
		err := os.RemoveAll(filepath.Join([]string{polycrateRuntimeDir, d.Name()}...))
		if err != nil {
			return err
		}
	}

	return nil
}

// func (p *Polycrate) PreloadWorkspaceWithContext(ctx context.Context, path string, validate bool) (*Workspace, error) {
// 	workspace, err := p.PreloadWorkspace(&PolycrateTransaction{}, path, validate)
// 	if err != nil {
// 		return nil, err
// 	}

// 	workspaceKey := ContextKey("workspace")
// 	ctx = context.WithValue(ctx, workspaceKey, workspace)

// 	log := polycrate.GetContextLogger(ctx)
// 	log = log.WithField("workspace", workspace.Name)
// 	ctx = polycrate.SetContextLogger(ctx, log)

// 	return workspace, nil
// }

// func (p *Polycrate) GetWorkspaceWithContext(ctx context.Context, path string, validate bool) (context.Context, *Workspace, error) {
// 	workspace, err := p.LoadWorkspace(&PolycrateTransaction{}, path, validate)
// 	if err != nil {
// 		return ctx, nil, err
// 	}

// 	workspaceKey := ContextKey("workspace")
// 	ctx = context.WithValue(ctx, workspaceKey, workspace)

// 	log := polycrate.GetContextLogger(ctx)
// 	log = log.WithField("workspace", workspace.Name)
// 	ctx = polycrate.SetContextLogger(ctx, log)

// 	return ctx, workspace, nil
// }

func (p *Polycrate) PreloadWorkspace(tx *PolycrateTransaction, path string, validate bool) (*Workspace, error) {
	var workspace *Workspace
	var err error

	log := tx.Log.log.WithField("path", path)

	log.Debugf("Loading workspace from path")

	workspace = new(Workspace)

	// Make a hard copy of the defaultWorkspace
	*workspace = defaultWorkspace

	// Set the path to the given path
	workspace, err = workspace.Preload(tx, path, validate)
	if err != nil {
		return nil, err
	}

	return workspace, nil
}

// Loads a workspace from a given path
// Optionally: validate the workspace (default: false)
func (p *Polycrate) LoadWorkspace(tx *PolycrateTransaction, path string, validate bool) (*Workspace, error) {
	// 1. Check if its in polycrate.Workspaces via GetWorkspace
	// 2. If yes, load it from there, lock Polycrate, run workspace.Load(), and unlock
	// 3. If not, bootstrap from defaultWorkspace and add to polycrate.Workspaces
	var workspace *Workspace
	var err error

	workspace, err = p.GetWorkspace(path)
	if err != nil {
		tx.Log.Debugf("Loading workspace from %s", path)

		workspace = new(Workspace)

		// Make a hard copy of the defaultWorkspace
		*workspace = defaultWorkspace

		// Set the path to the given path
		workspace, err = workspace.Load(tx, path, validate)
		if err != nil {
			return nil, err
		}
	} else {
		// Reload workspace
		tx.Log.Debugf("Reloading workspace")

		//workspace.Reload(ctx)
		workspace, err = workspace.Reload(tx, validate)
		if err != nil {
			return nil, err
		}
	}
	//w.Load(ctx, path)

	//w.load().Flush()
	tx.Log.Debug("Workspace loaded")

	// RegisterWorkspace
	if err = p.RegisterWorkspace(workspace); err != nil {
		return nil, err
	}

	return workspace, nil
}

func (p *Polycrate) HighjackSigint(ctx context.Context, tx *PolycrateTransaction) {
	tx.Log.Debugf("Starting signal handler")

	//signals := make(chan os.Signal, 1)

	//signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	HighJackCTX, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	//defer stop()

	go func() {
		select {
		case <-tx.Context.Done():
			log.Debugf("Received ctx-done")
			return
		case <-HighJackCTX.Done():
			log.Errorf("Received CTRL-C")
			//stop()

			err := p.PruneContainer(tx)
			if err != nil {
				tx.Log.Fatal(err)
			}

			tx.Stop()
			tx.Log.Debugf("Stopping signal handler")
			stop()
		}
	}()

	// go func() {
	// 	<-signals

	// 	log.Errorf("Received CTRL-C")
	// 	err := p.PruneContainer(ctx)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	// Call the cancel func
	// 	cancel := p.GetContextCancel(ctx)
	// 	err = p.StopTransaction(ctx, cancel)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	//}()
}
func (p *Polycrate) WaitForGracefulShutdown() {
	SigIntCTX, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	go func() {
		select {
		case <-SigIntCTX.Done():
			// Loop over transactions and stop them
			for _, tx := range p.Transactions {
				tx.Log.Warn("Received CTRL-C. Stopping transaction")
				tx.Stop()
			}
			stop()
		}
	}()
}

// func (p *Polycrate) DiscoverWorkspaces() error {
// 	workspacesDir := polycrateWorkspaceDir

// 	if _, err := os.Stat(workspacesDir); !os.IsNotExist(err) {
// 		log.WithFields(log.Fields{
// 			"path": workspacesDir,
// 		}).Debugf("Discovering local workspaces")

// 		// This function adds all valid Blocks to the list of
// 		err := filepath.WalkDir(workspacesDir, p.WalkWorkspacesDir)
// 		if err != nil {
// 			return err
// 		}
// 	} else {
// 		log.WithFields(log.Fields{
// 			"path": workspacesDir,
// 		}).Debugf("Skipping workspace discovery. Local workspaces directory not found")
// 	}

// 	return nil
// }

// This functions looks for a specific workspace in polycrate.Workspaces
func (p *Polycrate) GetWorkspace(id string) (*Workspace, error) {

	for i := 0; i < len(p.Workspaces); i++ {

		workspace := p.Workspaces[i]
		//workspace := workspace
		if workspace != nil {
			if workspace.Name == id || workspace.LocalPath == id {
				return workspace, nil
			}
		}
	}

	err := fmt.Errorf("workspace not found in index")
	return nil, err
}

// func (p *Polycrate) WalkWorkspacesDir(path string, d fs.DirEntry, err error) error {
// 	if err != nil {
// 		return err
// 	}

// 	if !d.IsDir() {
// 		fileinfo, _ := d.Info()

// 		if fileinfo.Name() == WorkspaceConfigFile {
// 			workspaceConfigFileDir := filepath.Dir(path)
// 			log.WithFields(log.Fields{
// 				"path": workspaceConfigFileDir,
// 			}).Tracef("Local workspace detected")

// 			w := Workspace{}
// 			w.LocalPath = workspaceConfigFileDir
// 			w.Path = workspaceConfigFileDir

// 			log.WithFields(log.Fields{
// 				"path": w.Path,
// 			}).Tracef("Reading in local workspace")

// 			w.SoftloadWorkspaceConfig().Flush()

// 			// Check if the workspace has already been loaded to the local workspace index
// 			_i, err := p.GetWorkspace(w.Name)
// 			if err != nil {
// 				// Workspace not indexed
// 				p.Workspaces = append(p.Workspaces, _i)
// 			}
// 		}
// 	}
// 	return nil
// }

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

func (p *Polycrate) PullImage(ctx context.Context, image string) error {
	log.Infof("Pulling image: %s", image)

	err := PullImageGo(ctx, image)
	if err != nil {
		return err
	}

	// _, _, err = PullImage(ctx, image)
	// if err != nil {
	// 	return err
	// }
	return nil
}

// func (p *Polycrate) CreateContainer(ctx context.Context, name string, image string, command []string, env []string, mounts []string, workdir string, ports []string, labels []string) (string, error) {
// 	_, name, err := CreateContainer(ctx, name, image, command, env, mounts, workdir, ports, labels)
// 	if err != nil {
// 		return "", err
// 	}
// 	return name, nil
// }

// func (p *Polycrate) RemoveContainer(ctx context.Context, name string) error {
// 	err := RemoveContainer(ctx, name)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (p *Polycrate) CopyFromImage(ctx context.Context, image string, src string, dst string) error {
// 	txid := p.GetContextTXID(ctx)

// 	name, err := p.CreateContainer(ctx, txid.String(), image, nil, nil, nil, "", nil, nil)
// 	if err != nil {
// 		return err
// 	}

// 	err = CopyFromContainer(ctx, name, src, dst)
// 	if err != nil {
// 		return err
// 	}

// 	// err = p.RemoveContainer(ctx, name)
// 	// if err != nil {
// 	// 	return err
// 	// }
// 	return nil
// }

func (p *Polycrate) GetStableVersion(ctx context.Context) (string, error) {
	log.Debugf("Getting stable version from %s", polycrate.Config.Registry.Url)

	url := fmt.Sprintf("https://%s/api/v2.0/projects/library/repositories/polycrate/artifacts?q=%s&page=1&page_size=100", p.Config.Registry.Url, url.QueryEscape("tags=latest"))
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	// Unmarshal artifacts
	var repositoryArtifactList []HarborRepositoryArtifact
	err = json.NewDecoder(resp.Body).Decode(&repositoryArtifactList)
	if err != nil {
		return "", err
	}

	// Above query should only ever return 1 artifact
	if len(repositoryArtifactList) > 1 {
		return "", fmt.Errorf("unable to determine latest version. too many artifacts")
	}
	artifact := repositoryArtifactList[0]

	var v string
	for _, tag := range artifact.Tags {
		if tag.Name != "latest" && !strings.HasPrefix(tag.Name, "v") {
			v = tag.Name
			_, err := semver.NewVersion(v)
			if err != nil {
				return "", err
			}
		}
	}

	return v, nil
}

// func (p *Polycrate) UpdateCLI(ctx context.Context) error {
// 	log.Warn("Starting Self-Update")
// 	stableVersion, err := p.GetStableVersion(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	image := strings.Join([]string{WorkspaceConfigImageRef, stableVersion}, ":")
// 	err = p.CopyFromImage(ctx, image, "/usr/local/bin/polycrate", "/usr/local/bin/polycrate")

// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

func (p *Polycrate) PruneContainer(tx *PolycrateTransaction) error {
	log.Infof("Removing container")

	// docker container prune --filter label=polycrate.workspace.revision.transaction=%sw.
	filters := []string{
		fmt.Sprintf("label=polycrate.txid=%s", tx.TXID.String()),
	}

	exitCode, _, err := PruneContainer(tx,
		filters,
	)

	log := tx.Log.log.WithField("exit_code", exitCode)
	log.Debugf("Pruned container")

	// Handle pruning error
	if err != nil {
		return err
	}

	return nil
}

func (p *Polycrate) RunContainer(tx *PolycrateTransaction, mounts []string, env []string, ports []string, name string, labels []string, workdir string, image string, command []string) (int, string, error) {

	return RunContainer(
		tx,
		image,
		command,
		env,
		mounts,
		workdir,
		ports,
		labels)
}
func (p *Polycrate) BuildContainer(ctx context.Context, contextDir string, dockerfile string, tags []string) (string, error) {
	image, err := buildContainerImage(ctx, contextDir, dockerfile, tags)
	if err != nil {
		return "", err
	}
	return image, nil
}

func (pl *PolycrateLog) Load(ctx context.Context) *PolycrateLog {
	var log = logrus.New()

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

	logrusLevel, err := logrus.ParseLevel(logrusLogLevel)
	if err != nil {
		logrusLevel = logrus.InfoLevel
	}

	// Set global log level
	log.SetLevel(logrusLevel)

	// Report the calling function if loglevel == 3
	// if polycrate.Config.Loglevel == 3 {
	// 	log.SetReportCaller(true)
	// }

	// Set Formatter
	log.SetFormatter(&logrus.TextFormatter{})
	if polycrate.Config.Logformat == "json" {
		log.SetFormatter(&logrus.JSONFormatter{})
	}

	pl.log = log.WithContext(ctx)
	return pl
}

func (pl *PolycrateLog) SetField(key string, value string) *PolycrateLog {
	pl.log = pl.log.WithField(key, value)
	return pl
}

func (pl *PolycrateLog) Info(args ...interface{}) *PolycrateLog {
	pl.log.Info(args...)
	return pl
}
func (pl *PolycrateLog) Debug(args ...interface{}) *PolycrateLog {
	pl.log.Debug(args...)
	return pl
}
func (pl *PolycrateLog) Error(args ...interface{}) *PolycrateLog {
	pl.log.Error(args...)
	return pl
}
func (pl *PolycrateLog) Warn(args ...interface{}) *PolycrateLog {
	pl.log.Warn(args...)
	return pl
}
func (pl *PolycrateLog) Fatal(args ...interface{}) *PolycrateLog {
	pl.log.Fatal(args...)
	return pl
}
func (pl *PolycrateLog) Warnf(format string, args ...interface{}) *PolycrateLog {
	pl.log.Warnf(format, args...)
	return pl
}
func (pl *PolycrateLog) Infof(format string, args ...interface{}) *PolycrateLog {
	pl.log.Infof(format, args...)
	return pl
}
func (pl *PolycrateLog) Debugf(format string, args ...interface{}) *PolycrateLog {
	pl.log.Debugf(format, args...)
	return pl
}
func (pl *PolycrateLog) Tracef(format string, args ...interface{}) *PolycrateLog {
	pl.log.Tracef(format, args...)
	return pl
}
func (pl *PolycrateLog) Errorf(format string, args ...interface{}) *PolycrateLog {
	pl.log.Errorf(format, args...)
	return pl
}
func (pl *PolycrateLog) Fatalf(format string, args ...interface{}) *PolycrateLog {
	pl.log.Fatalf(format, args...)
	return pl
}
