package cmd

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"

	// "encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/gosimple/slug"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/crypto/ssh"

	goErrors "errors"

	"os"
	"os/exec"

	validator "github.com/go-playground/validator/v10"
	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type promptContent struct {
	errorMsg string
	label    string
}

// var activateFlag bool

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

func mapDockerTag(tag string) (string, string, string, string) {
	regex := regexp.MustCompile(`([^\/]+\.[^\/.]+)?\/?([^:]+):?(.+)?`)

	rs := regex.FindStringSubmatch(tag)

	fullTag := rs[0]
	registryUrl := rs[1]
	name := rs[2]
	version := rs[3]

	if registryUrl == "" {
		// Set default registry URL when no registry has been given in the tag
		registryUrl = polycrate.Config.Registry.Url
	}

	if version == "" {
		version = "latest"
	}

	return fullTag, registryUrl, name, version
	//return regex.MatchString(tag)
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

func RunCommand(ctx context.Context, env []string, name string, args ...string) (exitCode int, output string, err error) {
	log := log.WithField("command", name)
	log = log.WithField("args", strings.Join(args, " "))

	//log.Debug("Running command: ", name, " ", strings.Join(args, " "))
	log.Trace("Running shell command")

	cmd := exec.CommandContext(ctx, name, args...)

	cmd.Env = os.Environ()

	if len(env) > 0 {
		cmd.Env = append(cmd.Env, env...)
	}

	var stdBuffer bytes.Buffer
	if !interactive {

		mw := io.MultiWriter(os.Stdout, &stdBuffer)

		cmd.Stdout = mw
		cmd.Stderr = mw
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
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
			log.Warnf("Could not get exit code for failed program: %v, %v", name, args)
			log.Trace(err)
			exitCode = defaultFailedCode
		}
	} else {
		// success, exitCode should be 0 if go is ok
		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		exitCode = ws.ExitStatus()
	}
	return exitCode, stdBuffer.String(), err
}

func RunCommandWithOutput(ctx context.Context, env []string, name string, args ...string) (exitCode int, output string, err error) {
	// log := tx.Log.log
	// log = log.WithField("command", name)
	// log = log.WithField("args", strings.Join(args, " "))

	// //log.Debug("Running command: ", name, " ", strings.Join(args, " "))
	// log.Trace("Running shell command")

	var outb, errb bytes.Buffer

	cmd := exec.CommandContext(ctx, name, args...)

	cmd.Env = os.Environ()
	if len(env) > 0 {
		cmd.Env = append(cmd.Env, env...)
	}

	cmd.Stdout = &outb
	cmd.Stderr = &errb

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
			log.Warnf("Could not get exit code for failed program: %v, %v", name, args)
			exitCode = defaultFailedCode
		}
	} else {
		// success, exitCode should be 0 if go is ok
		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		exitCode = ws.ExitStatus()
	}

	return exitCode, string(outb.String()), err
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

func unzipSource(source, destination string) error {
	// 1. Open the zip file
	reader, err := zip.OpenReader(source)
	if err != nil {
		return err
	}
	defer reader.Close()

	// 2. Get the absolute destination path
	destination, err = filepath.Abs(destination)
	log.Debugf("filepath.Abs(f, destionation)")
	if err != nil {
		return err
	}

	// 3. Iterate over zip files inside the archive and unzip each of them
	for _, f := range reader.File {
		err := unzipFile(f, destination)
		log.Debugf("Unzipfile(f, destionation)")
		if err != nil {

			return err
		}
	}

	return nil
}

func unzipFile(f *zip.File, destination string) error {
	// 4. Check if file paths are not vulnerable to Zip Slip
	filePath := filepath.Join(destination, f.Name)
	log.Debugf("File: %s", f.Name)
	log.Debugf(filepath.Clean(destination))
	log.Debugf("Path separator: %s", string(os.PathSeparator))
	log.Debugf("filepath: %s", filePath)
	log.Debugf("destination: %s", destination)

	if f.Name == "./" {
		return nil
	}
	if !strings.HasPrefix(filePath, filepath.Clean(destination)+string(os.PathSeparator)) {
		return fmt.Errorf("invalid file path: %s", filePath)
	}

	// 5. Create directory tree
	if f.FileInfo().IsDir() {
		if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
			return err
		}
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return err
	}

	// 6. Create a destination file for unzipped content
	destinationFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	// 7. Unzip the content of a file and copy it to the destination file
	zippedFile, err := f.Open()
	if err != nil {
		return err
	}
	defer zippedFile.Close()

	if _, err := io.Copy(destinationFile, zippedFile); err != nil {
		return err
	}
	return nil
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

	//log.Debug("Downloaded file from ", url, " to ", fp)

	return nil
}

// func cleanupWorkspace() {
// 	if !workspace.containerStatus.Pruned {
// 		ctx := context.Background()
// 		workspace.PruneContainer(ctx)
// 	}
// }

// func walkWorkspacesDir(path string, d fs.DirEntry, err error) error {
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
// 			if localWorkspaceIndex[w.Name] != "" {
// 				err := fmt.Errorf("Workspace already exists: %s", w.Path)
// 				return err
// 			}
// 			localWorkspaceIndex[w.Name] = w.LocalPath
// 			workspacePaths = append(workspacePaths, workspaceConfigFileDir)
// 		}
// 	}
// 	return nil
// }
// func walkBlocksDir(path string, d fs.DirEntry, err error) error {
// 	if err != nil {
// 		return err
// 	}

// 	if !d.IsDir() {
// 		fileinfo, _ := d.Info()

// 		if fileinfo.Name() == workspace.Config.BlocksConfig {
// 			blockConfigFileDir := filepath.Dir(path)
// 			log.WithFields(log.Fields{
// 				"path": blockConfigFileDir,
// 			}).Debugf("Block detected")
// 			blockPaths = append(blockPaths, blockConfigFileDir)
// 		}
// 	}
// 	return nil
// }

// func loadInventory() {
// 	//loadDefaults()

// 	// Check if inventory.yml exists

// 	_inventoryPath := filepath.Join(workspace.LocalPath, "inventory.yml")
// 	if _, err := os.Stat(_inventoryPath); os.IsNotExist(err) {
// 		log.Fatal("inventory.yml not found. Please add an inventory.")
// 	} else {
// 		inventory = _inventoryPath
// 	}

// 	log.Debug("Loading inventory from " + inventory)
// 	inventoryConfigObject.SetConfigFile(inventory)
// 	inventoryConfigObject.SetConfigType("yaml")

// 	err := inventoryConfigObject.MergeInConfig()
// 	CheckErr(err)
// }

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

// func getRemoteFileContent(url string) (string, error) {
// 	// Get the data
// 	resp, err := http.Get(url)
// 	if err != nil {
// 		return "", err
// 	}
// 	defer resp.Body.Close()

// 	b, err := io.ReadAll(resp.Body)
// 	// b, err := ioutil.ReadAll(resp.Body)  Go.1.15 and earlier
// 	if err != nil {
// 		log.Fatalln(err)
// 	}

// 	return string(b), err
// }

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

func ValidateMetaDataName(name string) bool {
	regex := regexp.MustCompile(ValidateMetaDataNameRegex)
	// (?!.*--.*)^(?!.*__.*)

	return regex.MatchString(name)
}

func ValidateBlockName(name string) bool {
	regex := regexp.MustCompile(ValidateBlockNameRegex)
	//regex := regexp.MustCompile("^[a-zA-Z]+([-/_]?[a-zA-Z0-9_]+)+$")
	// (?!.*--.*)^(?!.*__.*)

	return regex.MatchString(name)
}

func validateMetadataName(fl validator.FieldLevel) bool {
	name := fl.Field().String()

	return ValidateMetaDataName(name)
}

func validateBlockName(fl validator.FieldLevel) bool {
	name := fl.Field().String()

	return ValidateBlockName(name)
}

// func discoverWorkspaces() error {
// 	workspacesDir := polycrateWorkspaceDir

// 	if _, err := os.Stat(workspacesDir); !os.IsNotExist(err) {
// 		log.WithFields(log.Fields{
// 			"path": workspacesDir,
// 		}).Debugf("Discovering local workspaces")

// 		// This function adds all valid Blocks to the list of
// 		err := filepath.WalkDir(workspacesDir, walkWorkspacesDir)
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

// func createZipFile(sourcePath string, filename string) (string, error) {

// 	f, err := ioutil.TempFile("/tmp", filename+".*.zip")
// 	if err != nil {
// 		return "", err
// 	}
// 	defer f.Close()

// 	log.WithFields(log.Fields{
// 		"source":      sourcePath,
// 		"destination": f.Name,
// 	}).Debugf("Creating ZIP file from source folder")

// 	writer := zip.NewWriter(f)
// 	defer writer.Close()

// 	// 2. Go through all the files of the source
// 	err = filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
// 		if err != nil {
// 			return err
// 		}

// 		// 3. Create a local file header
// 		header, err := zip.FileInfoHeader(info)
// 		if err != nil {
// 			return err
// 		}

// 		// set compression
// 		header.Method = zip.Deflate

// 		// 4. Set relative path of a file as the header name
// 		//header.Name, err = filepath.Rel(filepath.Dir(sourcePath), path)
// 		// https://stackoverflow.com/questions/57504246/how-to-compress-a-file-to-zip-without-directory-folder-in-go
// 		// log.Debugf("Zipping file %s", path)
// 		// log.Debugf("Header name %s", header.Name)
// 		log.Debugf("Source path %s", sourcePath)
// 		log.Debugf("Path %s", path)
// 		log.Debugf("path filepath.Dir %s", filepath.Dir(path))
// 		filePathRel, err := filepath.Rel(sourcePath, path)
// 		if err != nil {
// 			log.Error(err)
// 			return err
// 		}
// 		log.Debugf("filePathRel %s", filePathRel)

// 		header.Name = filePathRel
// 		// log.Debugf("filepath rel %s", filePathRel)
// 		if err != nil {
// 			log.Error(err)
// 			return err
// 		}
// 		if info.IsDir() {
// 			header.Name += "/"
// 		}
// 		log.Debugf("Zipping file at path %s", header.Name)

// 		// 5. Create writer for the file header and save content of the file
// 		headerWriter, err := writer.CreateHeader(header)
// 		if err != nil {
// 			return err
// 		}

// 		if info.IsDir() {
// 			return nil
// 		}

// 		f, err := os.Open(path)
// 		if err != nil {
// 			return err
// 		}
// 		defer f.Close()

// 		_, err = io.Copy(headerWriter, f)
// 		return err
// 	})
// 	if err != nil {
// 		return "", err
// 	}
// 	return f.Name(), nil
// }

func slugify(args []string) string {
	preSlug := strings.Join(args, "-")
	slug := slug.Make(preSlug)

	return slug
}

// func NewGitlabSyncProvider() PolycrateProvider {
// 	return config.Gitlab
// }

// func getSyncProvider() PolycrateProvider {
// 	if config.Sync.Provider == "gitlab" {
// 		var pf SyncProviderFactory = NewGitlabSyncProvider
// 		provider := pf()

// 		log.WithFields(log.Fields{
// 			"provider": "gitlab",
// 		}).Debugf("Loading sync provider")
// 		return provider
// 	}
// 	return nil
// }

// func NewSync(path string) (*PolycrateSync, error) {
// 	log.WithFields(log.Fields{
// 		"path": path,
// 	}).Debugf("Initializing Sync")
// 	s := PolycrateSync{}

// 	// Check if workspace.Remote is configured
// 	// if workspace.Remote == "" {
// 	// 	return nil, errors.New("workspace.remote needs to be configured for sync to work")
// 	// }

// 	//s.LoadProvider()

// 	// Upsert-style behaviour - load OR create the repo
// 	// - Checks if a repository exists at workspce.LocalPath
// 	// - Checks if the repository's remote is equal to the workspace's remote
// 	// - Updates the repository remote if not
// 	// - Creates a locl repository if none exists
// 	// - Creates a remote repository at the configured provider if configured and none exists
// 	// - configures remote from created project
// 	// - initializes the repository with the configured remote
// 	//s.LoadRepo().Flush()

// 	return &s, nil
// }

func marshalRSAPrivate(priv *rsa.PrivateKey) string {
	return string(pem.EncodeToMemory(&pem.Block{
		Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv),
	}))
}

func generateKey() (string, string, error) {
	reader := rand.Reader
	bitSize := 2048

	key, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		return "", "", err
	}

	pub, err := ssh.NewPublicKey(key.Public())
	if err != nil {
		return "", "", err
	}
	pubKeyStr := string(ssh.MarshalAuthorizedKey(pub))
	privKeyStr := marshalRSAPrivate(key)

	return pubKeyStr, privKeyStr, nil
}

func CreateSSHKeys(ctx context.Context, path string, SshPrivateKey string, SshPublicKey string) error {

	privKeyPath := filepath.Join(path, SshPrivateKey)
	pubKeyPath := filepath.Join(path, SshPublicKey)

	log.Trace("Asserting private ssh key at ", privKeyPath)
	log.Trace("Asserting public ssh key at ", pubKeyPath)

	_, privKeyErr := os.Stat(privKeyPath)
	_, pubKeyErr := os.Stat(pubKeyPath)

	// Check if keys do already exist
	if os.IsNotExist(privKeyErr) && os.IsNotExist(pubKeyErr) {
		// No keys found
		// Generate new ones
		pubKeyStr, privKeyStr, err := generateKey()
		if err != nil {
			return err
		}

		// Save private key
		privKeyFile, err := os.Create(privKeyPath)
		if err != nil {
			return err
		}

		defer privKeyFile.Close()

		_, errPrivKey := privKeyFile.WriteString(privKeyStr)
		if errPrivKey != nil {
			return errPrivKey
		}

		err = os.Chmod(privKeyPath, 0600)
		if err != nil {
			return err
		}

		// Save public key
		pubKeyFile, err := os.Create(pubKeyPath)
		if err != nil {
			return err
		}

		defer pubKeyFile.Close()

		_, errPubKey := pubKeyFile.WriteString(pubKeyStr)
		if errPubKey != nil {
			return errPubKey
		}

		err = os.Chmod(pubKeyPath, 0644)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("SSH Keys already exist")
	}

	return nil
}

func FormatCommand(cmd *cobra.Command) string {

	commandPath := cmd.CommandPath()
	localArgs := cmd.Flags().Args()

	localFlags := []string{}

	cmd.Flags().Visit(func(flag *pflag.Flag) {
		//fmt.Printf("--%s=%s\n", flag.Name, flag.Value)
		localFlags = append(localFlags, fmt.Sprintf("--%s=%s", flag.Name, flag.Value))
	})

	command := strings.Join([]string{
		commandPath,
		strings.Join(localArgs, " "),
		strings.Join(localFlags, " "),
	}, " ")

	return command
}

func mergeMaps(a, b map[interface{}]interface{}) map[interface{}]interface{} {
	out := make(map[interface{}]interface{}, len(a))
	for k, v := range a {
		out[k] = v
	}

	for k, v := range b {
		// If you use map[string]interface{}, ok is always false here.
		// Because yaml.Unmarshal will give you map[interface{}]interface{}.

		if v, ok := v.(map[interface{}]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[interface{}]interface{}); ok {
					out[k] = mergeMaps(bv, v)
					continue
				}

				if bv, ok := bv.(map[string]interface{}); ok {

					_bv := map[interface{}]interface{}{}

					inter, err := yaml.Marshal(bv)
					if err != nil {
						panic(err)
					}

					err = yaml.Unmarshal(inter, _bv)
					if err != nil {
						panic(err)
					}

					out[k] = mergeMaps(_bv, v)

					continue
				}

			}
		}
		if v, ok := v.(map[string]interface{}); ok {
			mapZ := map[interface{}]interface{}{}

			inter, err := yaml.Marshal(v)
			if err != nil {
				panic(err)
			}

			err = yaml.Unmarshal(inter, mapZ)
			if err != nil {
				panic(err)
			}

			if bv, ok := out[k]; ok {

				if bv, ok := bv.(map[interface{}]interface{}); ok {

					out[k] = mergeMaps(bv, mapZ)
					continue
				}
			}
			continue

		}

		out[k] = v
	}
	return out
}
