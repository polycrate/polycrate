package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	goErrors "errors"

	"os"
	"os/exec"

	validator "github.com/go-playground/validator/v10"
	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		if v.IsNil() {
			return true
		}
		return isEmptyValue(v.Elem())
	case reflect.Func:
		return v.IsNil()
	case reflect.Invalid:
		return true
	}
	return false
}

func CreateDir(path string) error {
	err := os.MkdirAll(path, os.ModePerm)
	return err
}

func CreateFile(path string) error {
	file, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	return file.Close()
}

func CheckErr(msg interface{}) {
	if msg != nil {
		log.Fatal(msg)
		os.Exit(1)
	}
}

func RunCommand(name string, args ...string) (exitCode int, err error) {
	log.Debug("Running command: ", name, " ", strings.Join(args, " "))

	cmd := exec.Command(name, args...)

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

func ToPathSlice(t reflect.Value, name string, dst []string) []string {
	typeName := t.Type().Name()
	if typeName != "" {
		fmt.Println(typeName)
		if typeName == "WorkspaceIndex" {
			return dst
		}
	}

	switch t.Kind() {
	case reflect.Ptr, reflect.Interface:
		return ToPathSlice(t.Elem(), strings.ToUpper(name), dst)

	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			fname := t.Type().Field(i).Name
			dst = ToPathSlice(t.Field(i), strings.ToUpper(name+"_"+fname), dst)
		}

	case reflect.Slice, reflect.Array:
		for i := 0; i < t.Len(); i++ {
			dst = ToPathSlice(t.Index(i), strings.ToUpper(name+"_"+strconv.Itoa(i)), dst)
		}

	case reflect.Map:
		for _, e := range t.MapKeys() {
			//v := t.MapIndex(e)
			dst = ToPathSlice(t.MapIndex(e), strings.ToUpper(name), dst)
		}

	default:
		var value string
		switch t.Kind() {
		case reflect.Bool:
			value = t.String()
		case reflect.Struct:
			value = t.String()
		default:
			fmt.Println(value)
		}
		fmt.Println(value)
		return append(dst, name+"="+fmt.Sprintf("%s", t))
	}
	return dst
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

func cleanupWorkspace() {
	if workspace.containerID != "" {
		log.Debugf("Pruning container with id '%s'", workspace.containerID)
		if cli, err := getDockerCLI(); err == nil {
			if err := pruneContainer(cli, workspace.containerID); err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatal(err)
		}
	}
}

func walkBlocksDir(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}

	if !d.IsDir() {
		fileinfo, _ := d.Info()

		if fileinfo.Name() == workspace.Config.BlocksConfig {
			blockConfigFileDir := filepath.Dir(path)
			log.WithFields(log.Fields{
				"path": blockConfigFileDir,
			}).Debugf("Block detected")
			blockPaths = append(blockPaths, blockConfigFileDir)
		}
	}
	return nil
}

func loadInventory() {
	//loadDefaults()

	// Check if inventory.yml exists

	_inventoryPath := filepath.Join(workspace.Path, "inventory.yml")
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

func validateMetadataName(fl validator.FieldLevel) bool {
	name := fl.Field().String()

	regex := regexp.MustCompile("^[a-zA-Z]+([-/_]?[a-zA-Z0-9_]+)+$")
	// (?!.*--.*)^(?!.*__.*)

	if regex.MatchString(name) {
		// check if there's any __ or -- or //
		//r2 := regexp.MustCompile(string("(--|\\/\\/|__)+"))
		log.WithFields(log.Fields{
			"validated": name,
			"regex":     regex.String(),
		}).Debugf("Name validation successful")
	} else {
		log.WithFields(log.Fields{
			"validated": name,
			"regex":     regex.String(),
		}).Warnf("Name validation failed")
		return false
	}
	return true
}
