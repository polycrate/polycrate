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
	"io/fs"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	semver "github.com/hashicorp/go-version"
	"github.com/imdario/mergo"
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

type PolycrateConfig struct {
	//Sync      PolycrateSyncConfig     `yaml:"sync,omitempty" mapstructure:"sync,omitempty" json:"sync,omitempty"`
	//Providers []PolycrateProvider     `yaml:"providers,omitempty" mapstructure:"providers,omitempty" json:"providers,omitempty"`
	//Gitlab   PolycrateGitlabProvider `yaml:"gitlab,omitempty" mapstructure:"gitlab,omitempty" json:"gitlab,omitempty"`
	Registry     Registry    `yaml:"registry,omitempty" mapstructure:"registry,omitempty" json:"registry,omitempty"`
	Sync         SyncOptions `yaml:"sync,omitempty" mapstructure:"sync,omitempty" json:"sync,omitempty"`
	Loglevel     int         `yaml:"loglevel,omitempty" mapstructure:"loglevel,omitempty" json:"loglevel,omitempty"`
	Logformat    string      `yaml:"logformat,omitempty" mapstructure:"logformat,omitempty" json:"logformat,omitempty"`
	Webhooks     []Webhook   `yaml:"webhooks,omitempty" mapstructure:"webhooks,omitempty" json:"webhooks,omitempty"`
	CheckUpdates bool        `yaml:"check_updates,omitempty" mapstructure:"check_updates,omitempty" json:"check_updates,omitempty"`
	//Workspace PolycrateWorkspaceDefaults `yaml:"workspace,omitempty" mapstructure:"workspace,omitempty" json:"workspace,omitempty"`
}

type PolycrateEvent struct {
	Labels      map[string]string `yaml:"labels,omitempty" mapstructure:"labels,omitempty" json:"labels,omitempty"`
	Data        interface{}       `yaml:"data,omitempty" mapstructure:"data,omitempty" json:"data,omitempty"`
	Workspace   string            `yaml:"workspace,omitempty" mapstructure:"workspace,omitempty" json:"workspace,omitempty"`
	Workflow    string            `yaml:"workflow,omitempty" mapstructure:"workflow,omitempty" json:"workflow,omitempty"`
	Step        string            `yaml:"step,omitempty" mapstructure:"step,omitempty" json:"step,omitempty"`
	Block       string            `yaml:"block,omitempty" mapstructure:"block,omitempty" json:"block,omitempty"`
	Action      string            `yaml:"action,omitempty" mapstructure:"action,omitempty" json:"action,omitempty"`
	Command     string            `yaml:"command,omitempty" mapstructure:"command,omitempty" json:"command,omitempty"`
	ExitCode    int               `yaml:"exit_code,omitempty" mapstructure:"exit_code,omitempty" json:"exit_code,omitempty"`
	UserEmail   string            `yaml:"user_email,omitempty" mapstructure:"user_email,omitempty" json:"user_email,omitempty"`
	UserName    string            `yaml:"user_name,omitempty" mapstructure:"user_name,omitempty" json:"user_name,omitempty"`
	Date        string            `yaml:"date,omitempty" mapstructure:"date,omitempty" json:"date,omitempty"`
	Transaction uuid.UUID         `yaml:"transaction,omitempty" mapstructure:"transaction,omitempty" json:"transaction,omitempty"`
	Version     string            `yaml:"version,omitempty" mapstructure:"version,omitempty" json:"version,omitempty"`
	Output      string            `yaml:"output,omitempty" mapstructure:"output,omitempty" json:"output,omitempty"`
}

type Webhook struct {
	Endpoint string            `yaml:"endpoint,omitempty" mapstructure:"endpoint,omitempty" json:"endpoint,omitempty"`
	Labels   map[string]string `yaml:"labels,omitempty" mapstructure:"labels,omitempty" json:"labels,omitempty"`
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

// func (p *Polycrate) GetTransaction(txid uuid.UUID) (*PolycrateTransaction, error) {
// 	for i := 0; i < len(p.Transactions); i++ {
// 		transaction := &p.Transactions[i]
// 		if transaction.TXID == txid {
// 			return transaction, nil
// 		}
// 	}
// 	return nil, fmt.Errorf("transaction not found: %s", txid.String())
// }

func NewEvent(ctx context.Context) *PolycrateEvent {
	event := &PolycrateEvent{
		Version:     polycrate.GetContextVersion(ctx),
		Transaction: polycrate.GetContextTXID(ctx),
		Labels: map[string]string{
			"monk.event.class": "polycrate",
		},
	}

	action, err := polycrate.GetContextAction(ctx)
	if err == nil {
		event.Action = action.Name
	}

	block, err := polycrate.GetContextBlock(ctx)
	if err == nil {
		event.Block = block.Name
	}

	workspace, err := polycrate.GetContextWorkspace(ctx)
	if err == nil {
		event.Workspace = workspace.Name
	}

	cmd, err := polycrate.GetContextCmd(ctx)
	if err == nil {
		event.Command = FormatCommand(cmd)
	}

	userInfo := polycrate.GetUserInfo(ctx)
	event.UserEmail = userInfo["email"]
	event.UserName = userInfo["name"]

	revision, err := polycrate.GetContextRevision(ctx)
	if err == nil {
		event.UserEmail = revision.UserEmail
		event.UserName = revision.UserName
		event.Date = revision.Date
		event.Output = revision.Output
	}
	event.Labels["monk.event.level"] = "Info"

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
	log.Tracef("Merging labels to event labels")
	_labels := &e.Labels
	if err := mergo.Merge(_labels, labels); err != nil {
		return err
	}
	return nil
}

func (e *PolycrateEvent) Submit(ctx context.Context, webhook Webhook) error {
	log := polycrate.GetContextLogger(ctx)
	log.WithField("endpoint", webhook.Endpoint)
	log.Debugf("Submitting event to webhook endpoint")

	eventData, err := e.ToJSON()
	if err != nil {
		return err
	}
	request, err := http.NewRequest("POST", webhook.Endpoint, bytes.NewBuffer(eventData))
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
	return nil
}

type PolycrateTransactionFn func(context.Context, string) error

func (p *Polycrate) WithTransaction(fn PolycrateTransactionFn, s string) error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	ctx, err := p.StartTransaction(ctx, cancelFunc)
	if err != nil {
		return err
	}

	// Call function
	if err := fn(ctx, s); err != nil {
		if err := polycrate.StopTransaction(ctx, cancelFunc); err != nil {
			return err
		}
		return err
	}

	// Stop transaction
	if err := polycrate.StopTransaction(ctx, cancelFunc); err != nil {
		return err
	}

	return nil

	//log := polycrate.GetContextLogger(ctx)
}

func (p *Polycrate) EventHandler(ctx context.Context) error {
	// Exit if there's no event in the context
	event, err := polycrate.GetContextEvent(ctx)
	if err != nil {
		return err
	}

	eg := new(errgroup.Group)

	eg.Go(func() error {
		for _, webhook := range p.Config.Webhooks {
			if err := event.MergeInLabels(webhook.Labels); err != nil {
				return err
			}

			if err := event.Submit(ctx, webhook); err != nil {
				return err
			}
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

func (p *Polycrate) GetUserInfo(ctx context.Context) map[string]string {
	email, _ := GitGetUserEmail(ctx)
	name, _ := GitGetUserName(ctx)
	user := map[string]string{
		"email": email,
		"name":  name,
	}

	return user
}

func (p *Polycrate) PublishEvent(ctx context.Context, data *WorkspaceRevision, webhook Webhook) error {
	//log := polycrate.GetContextLogger(ctx)

	event := PolycrateEvent{}
	event.Data = data

	if err := event.Submit(ctx, webhook); err != nil {
		return err
	}

	return nil
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

func (p *Polycrate) NewTransaction(ctx context.Context, cmd *cobra.Command) (context.Context, context.CancelFunc, error) {
	ctx, cancel := context.WithCancel(ctx)
	ctx, err := p.StartTransaction(ctx, cancel)
	if err != nil {
		return ctx, nil, err
	}

	revision := &WorkspaceRevision{}

	revision.Date = time.Now().Format(time.RFC3339)
	revision.Transaction = p.GetContextTXID(ctx)
	revision.Version = p.GetContextVersion(ctx)

	if cmd != nil {
		cmdKey := ContextKey("cmd")
		ctx = context.WithValue(ctx, cmdKey, cmd)
		revision.Command = FormatCommand(cmd)
	}

	userInfo := polycrate.GetUserInfo(ctx)
	revision.UserEmail = userInfo["email"]
	revision.UserName = userInfo["name"]

	revisionKey := ContextKey("revision")
	ctx = context.WithValue(ctx, revisionKey, revision)

	eventKey := ContextKey("event")
	ctx = context.WithValue(ctx, eventKey, NewEvent(ctx))

	return ctx, cancel, err
}

func (p *Polycrate) StartTransaction(ctx context.Context, cancelFunc func()) (context.Context, error) {
	TXIDKey := ContextKey("TXID")
	txid := uuid.New()
	versionKey := ContextKey("version")

	ctx = context.WithValue(ctx, TXIDKey, txid)
	ctx = context.WithValue(ctx, versionKey, version)

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

func (p *Polycrate) StopTransaction(ctx context.Context, cancel func()) error {
	txid := p.GetContextTXID(ctx)

	tx := p.GetTransaction(txid)

	log := p.GetContextLogger(ctx)
	//log = log.WithField("txid", txid)

	// Call cancelFunc
	defer cancel()

	// Unregister the transaction
	if err := p.UnregisterTransaction(tx); err != nil {
		return err
	}

	if err := polycrate.EventHandler(ctx); err != nil {
		// We're not terminating here to not block further execution
		log.Warnf("Event handler failed: %s", err)
	}

	log.Debug("Stopped transaction")

	return nil
}

func (p *Polycrate) GetContextTXID(ctx context.Context) uuid.UUID {
	k := ContextKey("TXID")
	t := ctx.Value(k).(uuid.UUID)
	return t
}

func (p *Polycrate) GetContextWorkspace(ctx context.Context) (*Workspace, error) {
	k := ContextKey("workspace")
	workspace := ctx.Value(k)

	if workspace == nil {
		err := fmt.Errorf("workspace not found in context")
		return nil, err
	}
	return workspace.(*Workspace), nil
}

func (p *Polycrate) GetContextBlock(ctx context.Context) (*Block, error) {
	k := ContextKey("block")
	block := ctx.Value(k)

	if block == nil {
		return nil, fmt.Errorf("block not found in context")
	}
	return block.(*Block), nil
}

func (p *Polycrate) GetContextAction(ctx context.Context) (*Action, error) {
	k := ContextKey("action")
	action := ctx.Value(k)

	if action == nil {
		return nil, fmt.Errorf("action not found in context")
	}
	return action.(*Action), nil
}

func (p *Polycrate) GetContextCmd(ctx context.Context) (*cobra.Command, error) {
	k := ContextKey("cmd")
	cmd := ctx.Value(k)

	if cmd == nil {
		return nil, fmt.Errorf("command not found in context")
	}
	return cmd.(*cobra.Command), nil
}

func (p *Polycrate) GetContextExitCode(ctx context.Context) (int, error) {
	k := ContextKey("exit_code")
	ExitCode := ctx.Value(k)

	if ExitCode == nil {
		return 0, fmt.Errorf("exit_code not found in context")
	}
	return ExitCode.(int), nil
}
func (p *Polycrate) GetContextOutput(ctx context.Context) (string, error) {
	k := ContextKey("output")
	output := ctx.Value(k)

	if output == nil {
		return "", fmt.Errorf("output not found in context")
	}
	return output.(string), nil
}

func (p *Polycrate) GetContextWorkflow(ctx context.Context) (*Workflow, error) {
	k := ContextKey("workflow")
	workflow := ctx.Value(k)

	if workflow == nil {
		return nil, fmt.Errorf("workflow not found in context")
	}
	return workflow.(*Workflow), nil
}
func (p *Polycrate) GetContextStep(ctx context.Context) (*Step, error) {
	k := ContextKey("step")
	step := ctx.Value(k)

	if step == nil {
		return nil, fmt.Errorf("step not found in context")
	}
	return step.(*Step), nil
}

func (p *Polycrate) GetContextRevision(ctx context.Context) (*WorkspaceRevision, error) {
	k := ContextKey("revision")
	revision := ctx.Value(k)

	if revision == nil {
		return nil, fmt.Errorf("revision not found in context")
	}
	return revision.(*WorkspaceRevision), nil
}

func (p *Polycrate) SetContextOutput(ctx context.Context, output string) context.Context {
	outputKey := ContextKey("output")
	ctx = context.WithValue(ctx, outputKey, output)
	return ctx
}

func (p *Polycrate) SetContextExitCode(ctx context.Context, exitCode int) context.Context {
	exitCodeKey := ContextKey("exit_code")
	ctx = context.WithValue(ctx, exitCodeKey, exitCode)
	return ctx
}

func (p *Polycrate) SetContextEvent(ctx context.Context, event *PolycrateEvent) context.Context {
	if action, err := p.GetContextAction(ctx); err == nil {
		event.Action = action.Name
	}
	if block, err := p.GetContextBlock(ctx); err == nil {
		event.Block = block.Name
	}
	if workspace, err := p.GetContextWorkspace(ctx); err == nil {
		event.Workspace = workspace.Name
	}
	if workflow, err := p.GetContextWorkflow(ctx); err == nil {
		event.Workflow = workflow.Name
	}
	if step, err := p.GetContextStep(ctx); err == nil {
		event.Step = step.Name
	}
	if exit_code, err := p.GetContextExitCode(ctx); err == nil {
		event.ExitCode = exit_code
	}
	if cmd, err := p.GetContextCmd(ctx); err == nil {
		event.Command = FormatCommand(cmd)
	}
	if output, err := p.GetContextOutput(ctx); err == nil {
		event.Output = output
	}

	eventKey := ContextKey("event")
	ctx = context.WithValue(ctx, eventKey, event)
	return ctx
}

func (p *Polycrate) GetContextEvent(ctx context.Context) (*PolycrateEvent, error) {
	k := ContextKey("event")
	event := ctx.Value(k)

	if event == nil {
		return nil, fmt.Errorf("event not found in context")
	}
	return event.(*PolycrateEvent), nil
}

func (p *Polycrate) GetContextVersion(ctx context.Context) string {
	k := ContextKey("version")
	t := ctx.Value(k).(string)
	return t
}

func (p *Polycrate) GetContextLogger(ctx context.Context) *log.Entry {
	LoggerKey := ContextKey("Logger")
	_log := ctx.Value(LoggerKey)

	return _log.(*log.Entry)
}

func (p *Polycrate) GetContextCancel(ctx context.Context) context.CancelFunc {
	txid := polycrate.GetContextTXID(ctx)
	cancel := p.GetTransaction(txid).CancelFunc
	return cancel
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
	if err := p.StopTransaction(ctx, cancelFunc); err != nil {
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
		log = log.WithField("path", path)
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
		log = log.WithField("path", blocksDir)
		log.Info("Creating directory")

		if err := CreateDir(blocksDir); err != nil {
			return nil, err
		}
	}

	log = log.WithField("workspace", workspace.Name)

	if withConfig {
		log = log.WithField("config", workspace.Config.WorkspaceConfig)
		if err := workspace.Save(ctx); err != nil {
			log.Warn("Config already exists")
		} else {
			log.Info("Config created")
		}
	}

	if withSSHKeys {
		err := workspace.CreateSshKeys(ctx)
		if err != nil {
			log.Warn(err)
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

func (p *Polycrate) GetWorkspaceWithContext(ctx context.Context, path string, validate bool) (context.Context, *Workspace, error) {
	workspace, err := p.LoadWorkspace(ctx, path, validate)
	if err != nil {
		return ctx, nil, err
	}

	workspaceKey := ContextKey("workspace")
	ctx = context.WithValue(ctx, workspaceKey, workspace)

	log := polycrate.GetContextLogger(ctx)
	log = log.WithField("workspace", workspace.Name)
	ctx = polycrate.SetContextLogger(ctx, log)

	return ctx, workspace, nil
}

func (p *Polycrate) LoadWorkspace(ctx context.Context, path string, validate bool) (*Workspace, error) {
	// 1. Check if its in polycrate.Workspaces via GetWorkspace
	// 2. If yes, load it from there, lock Polycrate, run workspace.Load(), and unlock
	// 3. If not, bootstrap from defaultWorkspace and add to polycrate.Workspaces
	var workspace *Workspace
	var err error
	log := p.GetContextLogger(ctx)
	log = log.WithField("path", path)

	workspace, err = p.GetWorkspace(path)
	if err != nil {
		log.Debugf("Loading workspace from path")

		workspace = new(Workspace)

		// Make a hard copy of the defaultWorkspace
		*workspace = defaultWorkspace

		// Set the path to the given path
		workspace, err = workspace.Load(ctx, path, validate)
		if err != nil {
			return nil, err
		}
	} else {
		// Reload workspace
		log.Debugf("Reloading workspace")

		//workspace.Reload(ctx)
		workspace, err = workspace.Reload(ctx, validate)
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

func (p *Polycrate) HighjackSigint(ctx context.Context) {
	log := polycrate.GetContextLogger(ctx)
	log.Debugf("Starting signal handler")

	//signals := make(chan os.Signal, 1)

	//signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	HighJackCTX, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	//defer stop()

	go func() {
		select {
		case <-ctx.Done():
			return
		case <-HighJackCTX.Done():
			log.Errorf("Received CTRL-C")
			//stop()

			err := p.PruneContainer(ctx)
			if err != nil {
				log.Fatal(err)
			}

			// Call the cancel func
			cancel := p.GetContextCancel(ctx)
			err = p.StopTransaction(ctx, cancel)
			if err != nil {
				log.Error("After Stop")
				log.Fatal(err)
			}
			stop()
			log.Debugf("Stopping signal handler")
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

func (p *Polycrate) PullImage(ctx context.Context, image string) error {
	log.Infof("Pulling image: %s", image)
	_, _, err := PullImage(ctx, image)
	if err != nil {
		return err
	}
	return nil
}

func (p *Polycrate) CreateContainer(ctx context.Context, name string, image string, command []string, env []string, mounts []string, workdir string, ports []string, labels []string) (string, error) {
	_, name, err := CreateContainer(ctx, name, image, command, env, mounts, workdir, ports, labels)
	if err != nil {
		return "", err
	}
	return name, nil
}

func (p *Polycrate) RemoveContainer(ctx context.Context, name string) error {
	err := RemoveContainer(ctx, name)
	if err != nil {
		return err
	}
	return nil
}

func (p *Polycrate) CopyFromImage(ctx context.Context, image string, src string, dst string) error {
	txid := p.GetContextTXID(ctx)

	name, err := p.CreateContainer(ctx, txid.String(), image, nil, nil, nil, "", nil, nil)
	if err != nil {
		return err
	}

	err = CopyFromContainer(ctx, name, src, dst)
	if err != nil {
		return err
	}

	// err = p.RemoveContainer(ctx, name)
	// if err != nil {
	// 	return err
	// }
	return nil
}

func (p *Polycrate) GetStableVersion(ctx context.Context) (string, error) {
	log := p.GetContextLogger(ctx)

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

func (p *Polycrate) UpdateCLI(ctx context.Context) error {
	log.Warn("Starting Self-Update")
	stableVersion, err := p.GetStableVersion(ctx)
	if err != nil {
		return err
	}

	image := strings.Join([]string{WorkspaceConfigImageRef, stableVersion}, ":")
	err = p.CopyFromImage(ctx, image, "/usr/local/bin/polycrate", "/usr/local/bin/polycrate")

	if err != nil {
		return err
	}

	return nil
}

func (p *Polycrate) PruneContainer(ctx context.Context) error {
	log := polycrate.GetContextLogger(ctx)
	txid := polycrate.GetContextTXID(ctx)

	log.Infof("Removing container")

	// docker container prune --filter label=polycrate.workspace.revision.transaction=%sw.
	filters := []string{
		fmt.Sprintf("label=polycrate.txid=%s", txid),
	}

	exitCode, _, err := PruneContainer(ctx,
		filters,
	)

	log.WithField("exit_code", exitCode)
	log.Debugf("Pruned container")

	// Handle pruning error
	if err != nil {
		return err
	}

	return nil
}

func (p *Polycrate) RunContainer(ctx context.Context, mounts []string, env []string, ports []string, name string, labels []string, workdir string, image string, command []string) (int, string, error) {

	return RunContainer(
		ctx,
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
