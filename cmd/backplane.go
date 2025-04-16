package cmd

import (
	"bytes"
	goErrors "errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var backplaneCmd = &cobra.Command{
	Use:   "backplane",
	Short: "Show Backplane info",
	Long:  ``,
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		tx := polycrate.Transaction()
		tx.SetCommand(cmd)
		defer tx.Stop()

		fmt.Println("Backplane Info here")

		backplane.SystemInfo(tx)
	},
}

func init() {
	rootCmd.AddCommand(backplaneCmd)
}

//"github.com/apex/log"

type BackplaneWorkspaceResponse struct {
	Count    int                  `json:"count"`
	Next     string               `json:"next,omitempty"`
	Previous string               `json:"previous,omitempty"`
	Results  []BackplaneWorkspace `json:"results"`
}

type BackplaneLogin struct {
	Expiry string `json:"expiry"`
	Token  string `json:"token"`
}

type BackplaneSystemState struct {
	BlockCount                                 int     `json:"block_count"`
	WorkspaceCount                             int     `json:"workspace_count"`
	K8sClusterCount                            int     `json:"k8scluster_count"`
	K8sAppInstanceCount                        int     `json:"k8sappinstance_count"`
	OrganizationCount                          int     `json:"organization_count"`
	SystemCommitID                             string  `json:"system_commit_id"`
	HarborConnectionState                      string  `json:"harbor_connection_state"`
	HarborHelmMirrorProjectID                  string  `json:"harbor_helm_mirror_project_id"`
	HarborHelmMirrorProjectName                string  `json:"harbor_helm_mirror_project_name"`
	HarborHelmMirrorProjectState               string  `json:"harbor_helm_mirror_project_state"`
	SystemLastControlLoopExecution             string  `json:"system_last_control_loop_execution"`
	SystemLastControlLoopDurationSeconds       float32 `json:"system_last_control_loop_duration_seconds"`
	SystemMedianControlLoopDurationSeconds     float32 `json:"system_median_control_loop_duration_seconds"`
	SystemLastSystemStateLoopExecution         string  `json:"system_last_system_state_loop_execution"`
	SystemLastSystemStateLoopDurationSeconds   float32 `json:"system_last_system_state_loop_duration_seconds"`
	SystemMedianSystemStateLoopDurationSeconds float32 `json:"system_median_system_state_loop_duration_seconds"`
}

type BackplaneCondition struct {
	Type    string                 `json:"type"`
	Reason  string                 `json:"reason"`
	Context map[string]interface{} `json:"context"`
}
type BackplaneManagedObject struct {
	ID                    string               `json:"id"`
	Name                  string               `json:"name"`
	Kind                  string               `json:"kind"`
	State                 string               `json:"state"`
	ReconciliationRunning bool                 `json:"reconciliation_running"`
	DebugMode             bool                 `json:"debug_mode"`
	Conditions            []BackplaneCondition `json:"conditions"`
}
type BackplaneOrganization struct {
	LegalName     string               `json:"legal_name"`
	IsSystemOwner bool                 `json:"is_system_owner"`
	Workspaces    []BackplaneWorkspace `json:"workspaces"`
	BackplaneManagedObject
}
type BackplaneWorkspace struct {
	GitBranch         string                `json:"git_branch"`
	GitCommitShortSHA string                `json:"git_commit_short_sha"`
	Organization      BackplaneOrganization `json:"organization"`
	BackplaneManagedObject
}

type Backplane struct {
	Url      string `yaml:"url" mapstructure:"url" json:"url" validate:"required"`
	ApiKey   string `yaml:"api_key,omitempty" mapstructure:"api_key,omitempty" json:"api_key,omitempty"`
	Username string `yaml:"username,omitempty" mapstructure:"username,omitempty" json:"username,omitempty"`
	Password string `yaml:"password,omitempty" mapstructure:"password,omitempty" json:"password,omitempty"`
}

func (b *Backplane) LoadToken(tx *PolycrateTransaction) error {
	tokenFile := ".backplane-token.poly"
	tokenPath := filepath.Join(polycrateConfigDir, tokenFile)
	tx.Log.Debugf("Loading backplane token from %s", tokenPath)

	token, err := os.ReadFile(tokenPath)
	if err != nil {
		return err
	}

	polycrate.Config.Backplane.ApiKey = string(token)

	return nil
}

func (b *Backplane) Login(tx *PolycrateTransaction) error {
	// Post to /api/v1/login to obtain token
	// Save token to $CFG_DIR/.backplane-token.poly

	url := fmt.Sprintf("%s/api/login/", polycrate.Config.Backplane.Url)
	tx.Log.Debugf("Calling URL %s", url)

	// JSON body
	bodyMap := map[string]string{
		"username": polycrate.Config.Backplane.Username,
		"password": polycrate.Config.Backplane.Password,
	}
	bodyBytes, _ := json.Marshal(bodyMap)
	body := bytes.NewBuffer(bodyBytes)

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	backplaneLogin := &BackplaneLogin{}
	derr := json.NewDecoder(resp.Body).Decode(backplaneLogin)
	if derr != nil {
		return derr
	}

	if resp.StatusCode != http.StatusOK {
		panic(resp.Status)
	}

	tx.Log.Infof("Token: %s", backplaneLogin.Token)

	tokenFile := ".backplane-token.poly"
	tokenPath := filepath.Join(polycrateConfigDir, tokenFile)
	tx.Log.Debugf("Saving backplane token to %s", tokenPath)

	f, err := os.OpenFile(tokenPath, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(backplaneLogin.Token)
	if err != nil {
		return err
	}

	return nil
}

func (b *Backplane) SystemInfo(tx *PolycrateTransaction) error {
	b.LoadToken(tx)

	url := fmt.Sprintf("%s/api/v1/system-state/", polycrate.Config.Backplane.Url)
	tx.Log.Debugf("Calling URL %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", polycrate.Config.Backplane.ApiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	backplaneSystemState := &BackplaneSystemState{}
	derr := json.NewDecoder(resp.Body).Decode(backplaneSystemState)
	if derr != nil {
		return derr
	}

	if resp.StatusCode != http.StatusOK {
		panic(resp.Status)
	}

	outputFormat = "json"
	printObject(backplaneSystemState)

	return nil
}

func (b *Backplane) GetWorkspace(tx *PolycrateTransaction, name string, organization string) (*BackplaneWorkspace, error) {
	b.LoadToken(tx)

	url := fmt.Sprintf("%s/api/v1/workspaces/?name=%s&organization=%s", polycrate.Config.Backplane.Url, name, organization)
	tx.Log.Debugf("Calling URL %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", polycrate.Config.Backplane.ApiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	//backplaneWorkspaces := &BackplaneWorkspace{}
	backplaneResponse := &BackplaneWorkspaceResponse{}

	// bodyBytes, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// bodyString := string(bodyBytes)
	// log.Info(bodyString)

	derr := json.NewDecoder(resp.Body).Decode(backplaneResponse)
	if derr != nil {
		return nil, derr
	}

	if resp.StatusCode != http.StatusOK {
		panic(resp.Status)
	}

	workspaces := backplaneResponse.Results

	if len(workspaces) > 1 {
		//printObject(workspaces)
		return nil, goErrors.New("more than 1 Workspace matched")
	}

	if len(workspaces) == 0 {
		return nil, goErrors.New("no Workspace matched")
	}

	workspace := &workspaces[0]

	return workspace, nil
}
