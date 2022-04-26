package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	goErrors "errors"

	"os"
	"os/exec"

	"github.com/google/uuid"
	"github.com/manifoldco/promptui"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

func saveExecutionScript(script []string) (error, string) {
	// Prepare script
	scriptSlice := []string{
		"#!/bin/bash",
		"set -euo pipefail",
	}
	script = append(scriptSlice, script...)

	f, err := ioutil.TempFile("/tmp", "cloudstack."+workspace.Metadata.Name+"."+blockName+".script.*.sh")
	if err != nil {
		return err, ""
	}
	datawriter := bufio.NewWriter(f)

	for _, data := range script {
		_, _ = datawriter.WriteString(data + "\n")
	}

	datawriter.Flush()
	log.Debug("Saved script to " + f.Name())

	err = os.Chmod(f.Name(), 0755)
	if err != nil {
		return err, ""
	}

	// Closing file descriptor
	// Getting fatal errors on windows WSL2 when accessing
	// the mounted script file from inside the container if
	// the file descriptor is still open
	// Works flawlessly with open file descriptor on M1 Mac though
	// It's probably safer to close the fd anyways
	f.Close()
	return nil, f.Name()
}

func RunCommand(name string, args ...string) (exitCode int, err error) {
	log.Debug("Running command: ", name, " ", strings.Join(args, " "))

	var cmd *exec.Cmd
	cmd = exec.Command(name, args...)

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, workspace.DumpEnv()...)

	if !interactive {

		var stdBuffer bytes.Buffer
		mw := io.MultiWriter(os.Stdout, &stdBuffer)

		cmd.Stdout = mw
		cmd.Stderr = mw
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
	}
	err = cmd.Run()

	if err != nil {
		// try to get the exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			exitCode = ws.ExitStatus()
		} else {
			// This will happen (in OSX) if `name` is not available in $PATH,
			// in this situation, exit code could not be get, and stderr will be
			// empty string very likely, so we use the default fail code, and format err
			// to string and set to stderr
			log.Printf("Could not get exit code for failed program: %v, %v", name, args)
			exitCode = defaultFailedCode
		}
	} else {
		// success, exitCode should be 0 if go is ok
		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		exitCode = ws.ExitStatus()
	}
	return exitCode, err
}

func PullContainerImage(image string, version string) error {

	args := []string{"pull", strings.Join([]string{image, version}, ":"), "-q"}
	_, err := RunCommand("docker", args...)
	CheckErr(err)

	return err
}

func RunContainer(imageReference string, imageVersion string, command []string) (int, error) {
	image := strings.Join([]string{imageReference, imageVersion}, ":")

	// Prepare container command
	var runCmd []string

	// https://stackoverflow.com/questions/16248241/concatenate-two-slices-in-go
	runCmd = append(runCmd, []string{"run", "--rm", "-t"}...)

	// Env
	for _, envVar := range workspace.DumpEnv() {
		runCmd = append(runCmd, []string{"-e", envVar}...)
	}

	// Mounts
	for _, bindMount := range mounts {
		runCmd = append(runCmd, []string{"-v", bindMount}...)
	}

	// Ports
	for _, port := range ports {
		runCmd = append(runCmd, []string{"-p", port}...)
	}

	// Workdir
	runCmd = append(runCmd, []string{"--workdir", workdirContainer}...)

	// Hostname + Name
	runCmd = append(runCmd, []string{"--hostname", workspace.Metadata.Name}...)
	runCmd = append(runCmd, []string{"--name", strings.Join([]string{workspace.Metadata.Name, callUUID}, "-")}...)

	// Labels

	runCmd = append(runCmd, []string{"--label", strings.Join([]string{"polycrate.workspace", workspace.Metadata.Name}, "=")}...)
	runCmd = append(runCmd, []string{"--label", strings.Join([]string{"polycrate.block", blockName}, "=")}...)
	runCmd = append(runCmd, []string{"--label", strings.Join([]string{"polycrate.action", actionName}, "=")}...)
	runCmd = append(runCmd, []string{"--label", strings.Join([]string{"polycrate.uuid", callUUID}, "=")}...)

	// Pull
	// if pull {
	// 	runCmd = append(runCmd, []string{"--pull", "always"}...)
	// } else {
	// 	runCmd = append(runCmd, []string{"--pull", "never"}...)
	// }

	// Interactive
	if interactive {
		log.Warn("Running in interactive mode")
		runCmd = append(runCmd, []string{"-it"}...)
	}

	// Platform
	// fixed in cloudstack/cloudstack 1.1.3-main.build-46effead
	// Multi-platform images possible!
	// runCmd = append(runCmd, []string{"--platform", "linux/amd64"}...)

	// Image
	runCmd = append(runCmd, image)

	// Command
	runCmd = append(runCmd, command...)

	// Run container
	exitCode, err := RunCommand("docker", runCmd...)

	return exitCode, err
}

func getKubeconfigPath(configDir string) string {
	var configPath string
	configFiles := []string{"kubeconfig.yml", "kubeconfig", "kubeconfig.yaml"}
	for _, file := range configFiles {
		path := filepath.Join(configDir, file)
		log.Debug("Looking for kubeconfig at ", path)
		if _, err := os.Stat(path); err == nil {
			// Kubeconfig found in the current directory
			configPath = path
			log.Debug("Found kubeconfig at ", path)
			return configPath
		}
	}
	log.Debug("Couldn't find kubeconfig")
	return ""
}

func discoverKubeconfig() error {
	// Additionally, check for KUBECONFIG env var
	kubeconfigEnv := os.Getenv("KUBECONFIG")

	if kubeconfigEnv != "" {
		kubeconfig = kubeconfigEnv
		log.Debug("Setting kubeconfig from KUBECONFIG env var to ", kubeconfigEnv)
	}

	log.Debug("Trying to find a kubeconfig in ", workspace.path)

	// Get stack kubeconfig
	stackKubeConfigPath := getKubeconfigPath(workspace.path)

	// Overwrite config if kubeconfig has been found in stack_dir
	if stackKubeConfigPath != "" {
		//viper.Set("kubeconfig", stackKubeConfigPath)
		kubeconfig = stackKubeConfigPath
		log.Debug("Setting kubeconfig to ", stackKubeConfigPath)
	}

	return nil
}

func getConfigPath(configDir string) string {
	var configPath string
	configFiles := []string{"Stackfile"}
	for _, file := range configFiles {
		path := filepath.Join(configDir, file)
		log.Debug("Looking for config at ", path)
		if _, err := os.Stat(path); err == nil {
			// ACS config found in the current directory
			// deduct stack from CWD name
			configPath = path
			log.Debug("Found config at ", path)
			return configPath
		}
	}
	log.Debug("Couldn't find config")

	return ""
}

func CreateStackDir(stackName string) (string, error) {
	var stackDir string = GetStackDir(stackName)

	err := os.MkdirAll(stackDir, os.ModePerm)
	CheckErr(err)

	return stackDir, nil
}

func CheckStackDirAvailable(stackDir string) bool {
	_, err := os.Stat(stackDir)
	return err != nil
}

func CheckStackExists(stackName string) bool {
	return CheckStackDirAvailable(stackName)
}

func GetStackDir(stackName string) string {
	acsDir := workspaceConfig.GetString("acsDir")
	var stackDir string = filepath.Join(acsDir, stackName)

	return stackDir
}

func CheckKubeconfigExists(kubeconfig string) bool {
	_, err := os.Stat(kubeconfig)
	return err == nil
}

func DownloadFile(url string, fp string) error {
	// Create the file
	out, err := os.Create(fp)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)

	if err != nil {
		return err
	}

	if resp.StatusCode == 404 {
		//log.Error("Download failed: file not found (404). URL: " + url)
		err = goErrors.New("Download failed: file not found (404). URL: " + url)
		return err
	}

	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	log.Debug("Downloaded file from ", url, " to ", fp)

	return nil
}

func getRemoteFileContent(url string) (string, error) {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	// b, err := ioutil.ReadAll(resp.Body)  Go.1.15 and earlier
	if err != nil {
		log.Fatalln(err)
	}

	return string(b), err
}

func walkBlocksDir(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}

	if !d.IsDir() {
		fileinfo, _ := d.Info()

		if fileinfo.Name() == blockConfigFile {
			blockConfigFileDir := filepath.Dir(path)
			log.Debug("Block detected - found " + blockConfigFile + " in " + blockConfigFileDir)
			blockPaths = append(blockPaths, blockConfigFileDir)
		}
	}
	return nil
}

func loadWorkspace() error {
	log.Debug("Loading Workspace")

	// // Check if Stackfile / acs.yml exists
	// if _, err := os.Stat(workspaceConfigFilePath); os.IsNotExist(err) {
	// 	return goErrors.New(workspaceConfigFilePath, " does not exist")
	// } else {
	// 	runtimeWorkspaceConfigFilePath = workspaceConfigFilePath
	// }

	// Check overrides
	if len(overrides) > 0 {
		for _, override := range overrides {
			// Split string by =
			kv := strings.Split(override, "=")

			// Override property
			log.Debug("Setting " + kv[0] + " to " + kv[1])
			workspaceConfig.Set(kv[0], kv[1])
		}
	}

	workspaceConfig.SetConfigType("yaml")

	// Load plugin configs
	var blockDirContent []fs.FileInfo
	blockDirPath := filepath.Join(workspace.path, blocksRoot)
	if _, err := os.Stat(blockDirPath); !os.IsNotExist(err) {
		if blockDirContent, err = ioutil.ReadDir(blockDirPath); err != nil {
			return err
		}

		// Loop over block folder content
		for _, file := range blockDirContent {
			if file.IsDir() {
				// This is a plugin!
				blockName := file.Name()
				log.Debug("Loading config for block " + blockName)
				blockPath := filepath.Join(blockDirPath, blockName)
				blockConfigFilePath := filepath.Join(blockPath, blockConfigFile)

				// Lookup Pluginfile
				if _, err := os.Stat(blockConfigFilePath); !os.IsNotExist(err) {
					// block.yml exists
					// Merge to config

					blockConfigObject := viper.New()
					blockConfigObject.SetConfigType("yaml")

					// read in environment variables that match
					blockConfigObject.SetEnvPrefix(blockName)
					blockConfigObject.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
					blockConfigObject.AutomaticEnv()

					log.Debug("Loading ", blockConfigFile, " from "+blockPath)
					blockConfigObject.SetConfigFile(blockConfigFilePath)

					if err := blockConfigObject.MergeInConfig(); err != nil {
						return err
					}

					// Update cloudstack config at plugins.PLUGIN_NAME with pluginConfig
					// We construct a copy of the Stackfile format to be able to properly MERGE in the plugin config
					// which we need to do to be able to override it again from the actual Stackfile
					// Rebuilding this:
					// plugins:
					//   PLUGIN_NAME:
					//     ...
					m := make(map[string]interface{})
					p := make(map[string]interface{})
					p[blockName] = blockConfigObject.AllSettings()
					m["blocks"] = p

					//cloudstackConfig.Set(strings.Join([]string{"plugins", pluginName}, "."), pluginConfig.AllSettings())
					workspaceConfig.MergeConfigMap(m)
				}
			}
		}
	}

	// read in environment variables that match
	log.Debug("Loading from environment variables")
	workspaceConfig.SetEnvPrefix("cloudstack")
	workspaceConfig.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	workspaceConfig.AutomaticEnv()

	log.Debug("Trying to load ", workspaceConfigFile, " from ", workspace.path)
	workspaceConfig.SetConfigFile(workspaceConfigFilePath)

	// Config file not found
	if err := workspaceConfig.MergeInConfig(); err != nil {
		// Return a warning
		log.Warn(workspaceConfigFile, " not found in ", workspace.path)

		// Set workspace.Metadata.Name to basename of $PWD
		workspaceConfig.SetDefault("name", filepath.Base(cwd))
		workspaceConfig.SetDefault("display", "Ad-hoc Workspace in "+cwd)
	}

	// https://github.com/spf13/viper/issues/188#issuecomment-399518663
	// for _, key := range viper.AllKeys() {
	// 	val := viper.Get(key)
	// 	viper.Set(key, val)
	// }

	// // Verify config
	// // Check if "plugins" key exists
	// // This key is new in 2.0.0 (starting from 1.10.0)
	// // All plugins are now sorted under this key
	// // If it's missing, exit with notice to migrate config
	// if !workspaceConfig.IsSet("plugins") {
	// 	return goErrors.New("'plugins' key not found in configuration. Please migrate your configuration to the latest schema. For more information see https://docs.cloudstack.one/config-v2")
	// }

	// // Check if any of the legacy options is still configured
	// legacyOptions := []string{
	// 	"hcloud_csi",
	// 	"hcloud_vms",
	// 	"azure_aks",
	// 	"k3s",
	// 	"longhorn",
	// 	"argocd",
	// 	"letsencrypt",
	// 	"cert_manager_crds",
	// 	"eventrouter",
	// 	"external_dns",
	// 	"cilium_cni",
	// 	"loki",
	// 	"promtail",
	// 	"nginx_ingress",
	// 	"prometheus",
	// 	"tempo",
	// 	"portainer",
	// 	"portainer_agent",
	// 	"weave_scope",
	// 	"keycloak",
	// 	"kubeapps",
	// 	"sonarqube",
	// 	"metallb",
	// 	"fission",
	// 	"gitlab",
	// 	"harbor",
	// 	//
	// 	"hcloud",
	// 	"azure",
	// 	"csi", // Dropped
	// 	"ssh",
	// 	"sshd",
	// 	"route53",
	// 	"cloudflare",
	// 	"linkerd", // Dropped
	// 	"slack",
	// 	"alertmanager",
	// 	"grafana",
	// 	"paas", // Dropped
	// 	"bastion",
	// 	"stack.mail",   // moved to plugins.letsencrypt.config.mail
	// 	"stack.flavor", // Dropped
	// 	"stack.sso",    // Dropped
	// }

	// var legacyOptionError bool = false
	// for _, legacyOption := range legacyOptions {
	// 	if cloudstackConfig.IsSet(legacyOption) {
	// 		log.Error("Option '" + legacyOption + "' is depcrecated. Please migrate your configuration to the latest schema. For more information see https://docs.cloudstack.one/configuration/configuration")
	// 		legacyOptionError = true
	// 	}
	// }

	// if legacyOptionError {
	// 	return goErrors.New("Legacy options found")
	// }

	// Bind config to CLI flags

	// Unmarshal
	if err := unmarshalworkspaceConfig(); err != nil {
		return err
	}

	log.Debug("Workspace Config loaded")
	return nil
}

func loadInventory() {
	//loadDefaults()

	// Check if inventory.yml exists

	_inventoryPath := filepath.Join(workspace.path, "inventory.yml")
	if _, err := os.Stat(_inventoryPath); os.IsNotExist(err) {
		log.Fatal("inventory.yml not found. Please add an inventory.")
	} else {
		inventory = _inventoryPath
	}

	log.Debug("Loading inventory from " + inventory)
	inventoryConfigObject.SetConfigFile(inventory)
	inventoryConfigObject.SetConfigType("yaml")

	err := inventoryConfigObject.MergeInConfig()
	CheckErr(err)
}

func saveConfig() string {
	// Get temp path
	file, err := ioutil.TempFile("/tmp", "Stackfile."+workspaceConfig.GetString("stack.name")+"-*.yml")
	CheckErr(err)
	// TODO: make this toggleable
	//defer os.Remove(file.Name())

	workspaceConfig.WriteConfigAs(file.Name())
	return file.Name()
}

func saveRuntimeStackfile() {
	log.Debug("Setting runtime Stackfile: " + blockName)

	// Get temp path
	file, err := ioutil.TempFile("/tmp", "Stackfile."+workspace.Metadata.Name+"-*.yml")
	CheckErr(err)

	fileData, _ := yaml.Marshal(workspace)

	_ = ioutil.WriteFile(file.Name(), fileData, 0644)
	// TODO: make this toggleable
	//defer os.Remove(file.Name())

	//cloudstackConfig.WriteConfigAs(file.Name())
	//runtimeStackfile := file.Name()
}

func configToEnv() []string {
	var env_vars []string
	for _, key := range workspaceConfig.AllKeys() {
		var env_key_prefix string = "CLOUDSTACK"
		var env_key string = strings.ToUpper(strings.Join([]string{env_key_prefix, strings.ReplaceAll(key, ".", "_")}, "_"))
		var env_output string

		// Get value
		value := workspaceConfig.Get(key)

		switch value := value.(type) {
		case []map[string]interface{}:
			log.Debug("Type: []map[string]interface{}")
			for index, listItem := range value {
				// Extend key
				_env_key := strings.Join([]string{env_key, fmt.Sprint(index)}, "_")

				for mapKey, mapValue := range listItem {
					switch mapValue := mapValue.(type) {
					case string:
						// Extend key
						_sub_key := strings.Join([]string{_env_key, fmt.Sprint(strings.ToUpper(mapKey))}, "_")
						env_output = strings.Join([]string{_sub_key, fmt.Sprint(mapValue)}, "=")
						env_vars = append(env_vars, []string{env_output}...)
					case map[string]string:
						_sub_key := strings.Join([]string{_env_key, fmt.Sprint(strings.ToUpper(mapKey))}, "_")
						for subMapKey, subMapValue := range mapValue {
							// Extend key
							_sub_sub_key := strings.Join([]string{_sub_key, fmt.Sprint(strings.ToUpper(subMapKey))}, "_")
							env_output = strings.Join([]string{_sub_sub_key, fmt.Sprint(subMapValue)}, "=")
							env_vars = append(env_vars, []string{env_output}...)
						}
					}
				}
			}
		case []map[string]string:
			log.Debug("Type: []map[string")
			for index, listItem := range value {
				// Extend key
				_env_key := strings.Join([]string{env_key, fmt.Sprint(index)}, "_")

				for mapKey, mapValue := range listItem {
					// Extend key
					_sub_key := strings.Join([]string{_env_key, fmt.Sprint(strings.ToUpper(mapKey))}, "_")
					env_output = strings.Join([]string{_sub_key, fmt.Sprint(mapValue)}, "=")
					env_vars = append(env_vars, []string{env_output}...)
				}
			}
		case []string:
			for index, listItem := range value {
				// Extend key
				_env_key := strings.Join([]string{env_key, fmt.Sprint(index)}, "_")
				env_output = strings.Join([]string{_env_key, fmt.Sprint(listItem)}, "=")
				env_vars = append(env_vars, []string{env_output}...)
			}
			log.Debug("Type: []string")
		default:
			log.Debug("Type: unknown")
			env_output = strings.Join([]string{env_key, fmt.Sprint(value)}, "=")
			env_vars = append(env_vars, []string{env_output}...)
		}

	}
	sort.Strings(env_vars)
	return env_vars
}

func defaultDecoderConfig(output interface{}) *mapstructure.DecoderConfig {
	c := &mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           output,
		WeaklyTypedInput: true,
		ErrorUnused:      false,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		),
	}
	return c
}

func unmarshalworkspaceConfig() error {
	log.Debug("Unmarshalling Workspace config")
	err := workspaceConfig.UnmarshalExact(&workspace)
	if err != nil {
		return err
	}

	err = workspace.Validate()
	return err
}

func printObject(object interface{}) {
	if outputFormat == "json" {
		data, err := json.Marshal(object)
		CheckErr(err)
		fmt.Printf("%s\n", data)
	}
	if outputFormat == "yaml" {
		data, err := yaml.Marshal(object)
		CheckErr(err)
		fmt.Printf("%s\n", data)
	}
	if outputFormat == "env" {
		fmt.Println(workspaceConfig.AllKeys())
	}
}

func showConfig(format string) {
	if format == "json" {
		data, err := json.Marshal(workspace)
		CheckErr(err)
		fmt.Printf("%s\n", data)
	}
	if format == "yaml" {
		data, err := yaml.Marshal(workspace)
		CheckErr(err)
		fmt.Printf("%s\n", data)
	}
	if format == "env" {
		fmt.Println(workspaceConfig.AllKeys())
	}

}

func unmarshalStackfile2(format string) string {
	var err error
	var bs []byte
	var output string

	c := workspaceConfig.AllSettings()

	if format == "yaml" {
		bs, err = yaml.Marshal(c)
		if err != nil {
			log.Fatalf("unable to marshal config to "+format+": %v", err)
		}
		output = string(bs)
	} else if format == "env" {
		// TODO: implement env
		output = strings.Join(configToEnv(), "\n")
	} else {
		bs, err = yaml.Marshal(c)
		if err != nil {
			log.Fatalf("unable to marshal config to "+format+": %v", err)
		}
		output = string(bs)
	}
	return output
}

func callPlugin(block string, action string) (int, error) {
	log.Info("Calling plugin " + block + ", command " + action)
	var localPluginPath string = filepath.Join(workspace.path, blocksRoot, block)
	var containerPluginPath string = filepath.Join("/context", blocksRoot, block)

	// Load plugin and command config
	blockConfig := workspace.Blocks[0]
	actionConfig := blockConfig.Actions[0]

	// Lookup plugin directory
	if _, err := os.Stat(localPluginPath); !os.IsNotExist(err) {
		// plugin path does exist
		log.Info("Found user plugin at " + localPluginPath)

		// Set workdir
		if !local {
			workdirContainer = filepath.Join(containerPluginPath)
		} else {
			workdir = filepath.Join(localPluginPath)
		}
		log.Debug("Changing workdir to " + workdir)
	}

	// Validate that there's a script
	err := actionConfig.ValidateScript()
	if err != nil {
		return 1, err
	}

	// Check providers
	// var providers = map[string]string{}

	// for _, pluginItem := range cloudstack.Stack.Plugins {
	// 	// Add plugin itself as a provider
	// 	providers[pluginItem] = pluginItem

	// 	// Add items from plugin.provides
	// 	provides := cloudstack.Plugins[pluginItem].Provides
	// 	log.Debug("Plugin ", pluginItem, " provides ", provides)

	// 	for _, providerItem := range provides {
	// 		log.Debug("Added " + pluginItem + " as a provider for " + providerItem)
	// 		providers[providerItem] = pluginItem
	// 	}
	// }

	// // Check needs
	// needs := cloudstack.Plugins[plugin].Needs //viper.GetStringSlice("plugins." + plugin + ".needs")

	// for _, need := range needs {
	// 	if providers[need] == "" {
	// 		log.Fatal("Plugin " + plugin + " needs " + need + " which is not provided by any plugin in stack.plugins")
	// 	}
	// }

	err, executionScriptName := saveExecutionScript(actionConfig.Script)
	if err != nil {
		return 1, err
	}

	runCommand := []string{"bash", "-c", executionScriptName}

	// load environment variables
	//extraEnvironments := []string{strings.Join([]string{"CLOUDSTACK_SCRIPT_FILE", executionScriptName}, "=")}
	//env := getEnvironment(extraEnvironments)

	// load mounts
	// extraMounts := []string{strings.Join([]string{executionScriptName, executionScriptName}, ":")}
	// mounts := getMounts(extraMounts)

	// Check ports to open
	//ports := []string{}

	if workspace.Blocks[0].Actions[0].Interactive {
		interactive = true
	}

	exitCode, err := RunContainer(
		workspace.Config.Image.Reference,
		workspace.Config.Image.Version,
		runCommand,
	)

	if err != nil {

		log.Error("Plugin ", block, " failed with exit code ", exitCode, ": ", err.Error())
	} else {
		log.Info("Plugin ", block, " succeeded with exit code ", exitCode, ": OK")
	}
	// var stackHistory []StateHistoryItem

	// if err != nil {

	// 	log.Error("Plugin ", plugin, " failed with exit code ", exitCode, ": ", err.Error())

	// 	// Write state
	// 	stackHistory = append(state.History, StateHistoryItem{
	// 		Date:   time.Now().String(),
	// 		Commit: callUUID,
	// 		Data: map[string]interface{}{
	// 			"command": pluginCommand,
	// 			"plugin":  plugin,
	// 			"success": false,
	// 			"details": err.Error(),
	// 		},
	// 	})
	// 	stateConfig.Set("plugins."+plugin, nil)
	// } else {
	// 	log.Info("Plugin ", plugin, " succeeded with exit code ", exitCode, ": OK")
	// 	// Write state
	// 	stackHistory = append(state.History, StateHistoryItem{
	// 		Date:   time.Now().String(),
	// 		Commit: callUUID,
	// 		Data: map[string]interface{}{
	// 			"command": pluginCommand,
	// 			"plugin":  plugin,
	// 			"success": true,
	// 		},
	// 	})
	// 	stateConfig.Set("plugins."+plugin, cloudstack.Plugins[plugin])
	// }

	// stateConfig.Set("stack", cloudstack.Stack)
	// stateConfig.Set("history", stackHistory)
	// stateConfig.WriteConfigAs(filepath.Join(context, "Statefile"))

	//commitContext("chore(" + cloudstack.Stack.Name + "): cloudstack plugins " + plugin + " " + pluginCommand)

	return exitCode, err
}

func promptGetInput(pc promptContent) string {
	validate := func(input string) error {
		if len(input) <= 0 {
			return goErrors.New(pc.errorMsg)
		}
		return nil
	}

	templates := &promptui.PromptTemplates{
		Prompt:  "{{ . }} ",
		Valid:   "{{ . | green }} ",
		Invalid: "{{ . | red }} ",
		Success: "{{ . | bold }} ",
	}

	prompt := promptui.Prompt{
		Label:     pc.label,
		Templates: templates,
		Validate:  validate,
	}

	result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	return result
}

func promptGetSelect(pc promptContent) string {
	items := []string{"animal", "food", "person", "object"}
	index := -1
	var result string
	var err error

	for index < 0 {
		prompt := promptui.SelectWithAdd{
			Label:    pc.label,
			Items:    items,
			AddLabel: "Other",
		}

		index, result, err = prompt.Run()

		if index == -1 {
			items = append(items, result)
		}
	}

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	return result
}

func promptYesNo(pc promptContent) string {
	items := []string{"Yes", "No"}
	index := -1
	var result string
	var err error

	for index < 0 {
		prompt := promptui.Select{
			Label: pc.label,
			Items: items,
		}

		index, result, err = prompt.Run()

		if index == -1 {
			items = append(items, result)
		}
	}

	if err != nil {
		log.Error("Prompt failed %v\n", err)
		os.Exit(1)
	}

	return result
}

func loadConfigFile(fp string, t string) (*viper.Viper, error) {
	config := viper.New()
	config.SetConfigFile(fp)
	config.SetConfigType(t)

	if err := config.ReadInConfig(); err != nil {
		return nil, err
	}
	return config, nil
}

func getCallUUID() string {
	id := uuid.New()
	return id.String()
}
