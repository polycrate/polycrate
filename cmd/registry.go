package cmd

import (
	"fmt"

	"golang.org/x/exp/slices"

	log "github.com/sirupsen/logrus"
	//"github.com/apex/log"
)

type PostRaw struct {
	Id          int    `yaml:"id" mapstructure:"id" json:"id" validate:"required"`
	PostAuthor  string `yaml:"post_author" mapstructure:"post_author" json:"post_author" validate:"required"`
	PostTitle   string `yaml:"post_title" mapstructure:"post_title" json:"post_title" validate:"required"`
	PostExcerpt string `yaml:"post_excerpt,omitempty" mapstructure:"post_excerpt,omitempty" json:"post_excerpt,omitempty"`
	PostContent string `yaml:"post_content,omitempty" mapstructure:"post_content,omitempty" json:"post_content,omitempty"`
	PostDate    string `yaml:"post_date" mapstructure:"post_date" json:"post_date" validate:"required"`
	PostName    string `yaml:"post_name" mapstructure:"post_name" json:"post_name" validate:"required"`
	PostType    string `yaml:"post_type" mapstructure:"post_type" json:"post_type" validate:"required"`
}

type Post struct {
	Id          int                                 `yaml:"id" mapstructure:"id" json:"id" validate:"required"`
	Date        string                              `yaml:"date" mapstructure:"date" json:"date" validate:"required"`
	DateGmt     string                              `yaml:"date_gmt,omitempty" mapstructure:"date_gmt,omitempty" json:"date_gmt,omitempty"`
	Guid        map[string]string                   `yaml:"guid" mapstructure:"guid" json:"guid" validate:"required"`
	Modified    string                              `yaml:"modified" mapstructure:"modified" json:"modified" validate:"required"`
	ModifiedGmt string                              `yaml:"modified_gmt,omitempty" mapstructure:"modified_gmt,omitempty" json:"modified_gmt,omitempty"`
	Slug        string                              `yaml:"slug" mapstructure:"slug" json:"slug" validate:"required"`
	Status      string                              `yaml:"publish" mapstructure:"publish" json:"publish" validate:"required"`
	Type        string                              `yaml:"type" mapstructure:"type" json:"type" validate:"required"`
	Link        string                              `yaml:"link" mapstructure:"link" json:"link" validate:"required"`
	Title       map[string]string                   `yaml:"title" mapstructure:"title" json:"title" validate:"required"`
	Content     map[string]interface{}              `yaml:"content" mapstructure:"content" json:"content" validate:"required"`
	Template    string                              `yaml:"template,omitempty" mapstructure:"template,omitempty" json:"template,omitempty"`
	Links       map[string][]map[string]interface{} `yaml:"_links" mapstructure:"_links" json:"_links" validate:"required"`
}

type RegistryRelease struct {
	PostRaw
	Version       string `yaml:"version" mapstructure:"version" json:"version" validate:"required"`
	ReleaseBundle string `yaml:"release_bundle" mapstructure:"release_bundle" json:"release_bundle" validate:"required"`
}

type RegistryBlock struct {
	Post
	Releases  []RegistryRelease `yaml:"releases,omitempty" mapstructure:"releases,omitempty" json:"releases,omitempty"`
	BlockName string            `yaml:"block_name" mapstructure:"block_name" json:"block_name" validate:"required"`
}
type RegistryWorkspace struct {
	Post
	Releases      []RegistryRelease `yaml:"releases,omitempty" mapstructure:",omitempty" json:",omitempty"`
	WorkspaceName string            `yaml:"block_name" mapstructure:"block_name" json:"block_name" validate:"required"`
}

type Registry struct {
	Url       string `yaml:"url" mapstructure:"url" json:"url" validate:"required"`
	BaseImage string `yaml:"base_image" mapstructure:"base_image" json:"base_image" validate:"required"`
	//BlockNamespace string `yaml:"block_namespace" mapstructure:"block_namespace" json:"block_namespace" validate:"required"`
	//ApiBase   string `yaml:"api_base" mapstructure:"api_base" json:"api_base" validate:"required"`
	//Username  string `yaml:"username,omitempty" mapstructure:"username,omitempty" json:"username,omitempty"`
	//Password  string `yaml:"password,omitempty" mapstructure:"password,omitempty" json:"password,omitempty"`
}

// func (rb *RegistryBlock) UnmarshalJSON(b []byte) error {
// 	type TmpJson RegistryBlock

// 	//var tmpJson map[string]interface{}
// 	var tmpJson TmpJson
// 	err := json.Unmarshal(b, &tmpJson)
// 	if err != nil {
// 		return err
// 	}

// 	// Wordpress returns "releases: false" instead of "releases: []"
// 	// So we're overwriting a boolean value with an empty list during unmarshalling
// 	if reflect.TypeOf(tmpJson.Releases).String() == "bool" {
// 		rb.Releases = []RegistryRelease{}
// 		rb.Id = tmpJson.Id
// 		rb.BlockName = tmpJson.BlockName
// 		return nil
// 	}

// 	//return json.Unmarshal(b, &rb)
// 	*rb = RegistryBlock(tmpJson)
// 	return nil

// }

// func (o *Registry) SearchBlock(blockName string) ([]RegistryBlock, error) {
// 	url := fmt.Sprintf("%s/%s/block?search=%s", config.Registry.Url, config.Registry.ApiBase, blockName)

// 	response, err := http.Get(url)

// 	if err != nil {
// 		return nil, err
// 	}

// 	responseData, err := ioutil.ReadAll(response.Body)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var blocks []RegistryBlock
// 	err = json.Unmarshal(responseData, &blocks)
// 	if err != nil {
// 		return nil, err
// 	}
// 	//log.Trace(string(responseData))

// 	if len(blocks) > 0 {
// 		// Sort the Releases inside of a Block by date
// 		for _, block := range blocks {
// 			sort.Slice(block.Releases, func(i, j int) bool {
// 				return block.Releases[i].PostDate > block.Releases[j].PostDate
// 			})
// 		}

// 		return blocks, nil
// 	} else {
// 		err := fmt.Errorf("Block not found in registry: %s", blockName)
// 		return nil, err
// 	}

// }

// func (o *Registry) SearchWorkspace(workspaceName string) ([]RegistryWorkspace, error) {
// 	url := fmt.Sprintf("%s/%s/workspace?search=%s", o.Url, o.ApiBase, workspaceName)

// 	response, err := http.Get(url)

// 	if err != nil {
// 		return nil, err
// 	}

// 	responseData, err := ioutil.ReadAll(response.Body)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var workspaces []RegistryWorkspace
// 	err = json.Unmarshal(responseData, &workspaces)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if len(workspaces) > 0 {
// 		// Sort the Releases inside of a Block by date
// 		for _, workspace := range workspaces {
// 			sort.Slice(workspace.Releases, func(i, j int) bool {
// 				return workspace.Releases[i].PostDate > workspace.Releases[j].PostDate
// 			})
// 		}

// 		return workspaces, nil
// 	} else {
// 		err := fmt.Errorf("Workspace not found: %s", workspaceName)
// 		return nil, err
// 	}

// }

// func (o *Registry) GetBlock(blockName string) (*RegistryBlock, error) {
// 	registryBlocks, err := o.SearchBlock(blockName)
// 	if err != nil {
// 		return nil, err
// 	}

// 	blockIndex := slices.IndexFunc(registryBlocks, func(o RegistryBlock) bool { return o.BlockName == blockName })
// 	if blockIndex != -1 {
// 		return &registryBlocks[blockIndex], nil
// 	} else {
// 		err := fmt.Errorf("Block not found in registry: %s", blockName)
// 		return nil, err
// 	}

// }

// func (o *Registry) AddBlock(blockName string) (*RegistryBlock, error) {
// 	log.WithFields(log.Fields{
// 		"workspace": workspace.Name,
// 		"block":     blockName,
// 	}).Debugf("Adding block to registry")

// 	var registryBlock RegistryBlock

// 	client := &http.Client{}
// 	postUrl := fmt.Sprintf("%s/%s/block", config.Registry.Url, config.Registry.ApiBase)

// 	// Data
// 	values := map[string]interface{}{
// 		"block_name": blockName,
// 		"title":      blockName,
// 		"status":     "publish",
// 	}
// 	json_data, err := json.Marshal(values)
// 	if err != nil {
// 		return nil, err
// 	}

// 	req, err := http.NewRequest("POST", postUrl, bytes.NewBuffer(json_data))
// 	if err != nil {
// 		return nil, err
// 	}
// 	req.SetBasicAuth(config.Registry.Username, config.Registry.Password)
// 	req.Header.Set("Content-Type", "application/json")

// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return nil, err
// 	}

// 	err = json.Unmarshal(body, &registryBlock)
// 	if err != nil {
// 		return nil, err
// 	}

// 	log.WithFields(log.Fields{
// 		"workspace": workspace.Name,
// 		"block":     blockName,
// 	}).Debugf("New registry block created")

// 	return &registryBlock, nil
// }

// func (o *Registry) GetWorkspace(workspaceName string) (*RegistryWorkspace, error) {
// 	registryWorkspaces, err := o.SearchWorkspace(workspaceName)
// 	if err != nil {
// 		return nil, err
// 	}

// 	workspaceIndex := slices.IndexFunc(registryWorkspaces, func(o RegistryWorkspace) bool { return o.WorkspaceName == workspaceName })
// 	if workspaceIndex != -1 {
// 		return &registryWorkspaces[workspaceIndex], nil
// 	} else {
// 		err := fmt.Errorf("Block not found: %s", workspaceName)
// 		return nil, err
// 	}

// }

// func (o *RegistryBlock) AddRelease(version string, bundle string, filename string) (*RegistryRelease, error) {
// 	// 1. Create attachment
// 	// 2. Create post & link attachment
// 	// credentialString := strings.Join([]string{config.Registry.Username, config.Registry.Password}, ":")
// 	// credentials := base64.StdEncoding.EncodeToString([]byte(credentialString))

// 	// attachmentData := url.Values{
// 	// 	"title":          {filename},
// 	// 	"status":         {"publish"},
// 	// 	"content":        {""},
// 	// 	"slug":           {"---"},
// 	// 	"version":        {version},
// 	// 	"release_bundle": {version},
// 	// }
// 	// data := url.Values{
// 	// 	"title":          {"John Doe"},
// 	// 	"status":         {"publish"},
// 	// 	"content":        {""},
// 	// 	"slug":           {"---"},
// 	// 	"version":        {version},
// 	// 	"release_bundle": {version},
// 	// }

// 	attachmentData, err := os.Open(bundle)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	mediaUrl := fmt.Sprintf("%s/%s/media", config.Registry.Url, config.Registry.ApiBase)
// 	client := &http.Client{}
// 	req, err := http.NewRequest("POST", mediaUrl, attachmentData)
// 	if err != nil {
// 		return nil, err
// 	}
// 	req.SetBasicAuth(config.Registry.Username, config.Registry.Password)
// 	req.Header.Set("Content-Type", "application/zip")
// 	req.Header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", filename))

// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	responseData, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var attachmentPost Post
// 	err = json.Unmarshal(responseData, &attachmentPost)
// 	if err != nil {
// 		return nil, err
// 	}

// 	printObject(attachmentPost)

// 	// Create the release post
// 	releaseUrl := fmt.Sprintf("%s/%s/block_release", config.Registry.Url, config.Registry.ApiBase)

// 	// Data
// 	slug := slugify([]string{o.BlockName, version})
// 	title := strings.Join([]string{o.BlockName, version}, ":")
// 	values := map[string]interface{}{
// 		"release_name":   slug,
// 		"title":          title,
// 		"status":         "publish",
// 		"block":          o.Id,
// 		"release_bundle": attachmentPost.Id,
// 		"version":        version,
// 	}
// 	json_data, err := json.Marshal(values)
// 	if err != nil {
// 		return nil, err
// 	}

// 	req, err = http.NewRequest("POST", releaseUrl, bytes.NewBuffer(json_data))
// 	if err != nil {
// 		return nil, err
// 	}
// 	req.SetBasicAuth(config.Registry.Username, config.Registry.Password)
// 	req.Header.Set("Content-Type", "application/json")

// 	resp, err = client.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	responseData, err = ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var registryRelease Post
// 	err = json.Unmarshal(responseData, &registryRelease)
// 	if err != nil {
// 		return nil, err
// 	}
// 	printObject(registryRelease)

// 	// apiUrl := fmt.Sprintf("%s/%s/block_release", registry.Url, registry.ApiBase)
// 	// resp, err := http.PostForm(apiUrl, data)
// 	// var res map[string]interface{}

// 	// json.NewDecoder(resp.Body).Decode(&res)

// 	// fmt.Println(res["form"])
// 	// //response, err := http.Post(url)

// 	// if err != nil {
// 	// 	return nil, err
// 	// }
// 	return nil, nil

// 	// responseData, err := ioutil.ReadAll(response.Body)
// 	// if err != nil {
// 	// 	return nil, err
// 	// }
// }

func (o *RegistryBlock) GetRelease(version string) (*RegistryRelease, error) {
	// Get the correct release
	var releaseIndex int
	if version == "latest" {
		// Pick the latest release
		releaseIndex = 0
	} else {
		// Search release with given version
		releaseIndex = slices.IndexFunc(o.Releases, func(r RegistryRelease) bool { return r.Version == version })
	}

	if len(o.Releases) > 0 && releaseIndex != -1 {
		release := &o.Releases[releaseIndex]
		if release != nil {
			log.WithFields(log.Fields{
				"workspace":         workspace.Name,
				"block":             o.BlockName,
				"resolved_version":  release.Version,
				"requested_version": version,
			}).Debugf("Found release in registry")
			return release, nil
		} else {
			err := fmt.Errorf("Release not found: %s:%s", o.BlockName, version)
			return nil, err
		}
	} else {
		err := fmt.Errorf("Release not found: %s:%s", o.BlockName, version)
		return nil, err
	}
}

func (o *RegistryWorkspace) GetRelease(version string) (*RegistryRelease, error) {
	// Get the correct release
	var releaseIndex int
	if version == "latest" {
		// Pick the latest release
		releaseIndex = 0
	} else {
		// Search release with given version
		releaseIndex = slices.IndexFunc(o.Releases, func(r RegistryRelease) bool { return r.Version == version })
	}

	if len(o.Releases) > 0 && releaseIndex != -1 {
		release := &o.Releases[releaseIndex]
		if release != nil {
			log.WithFields(log.Fields{
				"workspace":         workspace.Name,
				"block":             o.WorkspaceName,
				"resolved_version":  release.Version,
				"requested_version": version,
			}).Debugf("Found release in registry")
			return release, nil
		} else {
			err := fmt.Errorf("Release not found: %s:%s", o.WorkspaceName, version)
			return nil, err
		}
	} else {
		err := fmt.Errorf("Release not found: %s:%s", o.WorkspaceName, version)
		return nil, err
	}
}

// func (o *Registry) resolveArg(arg string) (string, string, error) {
// 	var name string
// 	var version string

// 	// Split arg by :
// 	splitArgs := strings.Split(arg, ":")

// 	// length can be either 1 (BLOCK) or 2 (BLOCK:VERSION)
// 	// Everything else is an error
// 	switch len(splitArgs) {
// 	case 1:
// 		// it's only the block
// 		name = splitArgs[0]
// 		version = "latest"
// 	case 2:
// 		name = splitArgs[0]
// 		version = splitArgs[1]
// 	default:
// 		err := fmt.Errorf("Wrong format: %s - should be '$name:$version' as in 'test:0.0.1'", arg)
// 		return "", "", err
// 	}
// 	return name, version, nil
// }

// func (o *Registry) UpdateBlocks(args []string) error {
// 	for _, arg := range args {
// 		blockName, blockVersion, err := o.resolveArg(arg)
// 		if err != nil {
// 			return err
// 		}

// 		block := workspace.GetBlockFromIndex(blockName)
// 		if block != nil {
// 			log.WithFields(log.Fields{
// 				"workspace": workspace.Name,
// 				"block":     block.Name,
// 			}).Warnf("Updating block")

// 			// Search block in registry
// 			registryBlock, err := registry.GetBlock(blockName)
// 			if err != nil {
// 				if err != nil {
// 					log.WithFields(log.Fields{
// 						"workspace": workspace.Name,
// 						"block":     blockName,
// 					}).Debug(err)
// 					return err
// 				}
// 			}

// 			// Check if release exists
// 			_, err = registryBlock.GetRelease(blockVersion)
// 			if err != nil {
// 				return err
// 			}

// 			// Uninstall block
// 			// pruneBlock constains a boolean triggered by --prune
// 			// Now install the wanted version
// 			log.WithFields(log.Fields{
// 				"workspace": workspace.Name,
// 				"block":     blockName,
// 			}).Debugf("Uninstalling block from workspace")
// 			err = block.Uninstall(pruneBlock)
// 			if err != nil {
// 				return err
// 			}

// 			// Now install the wanted version
// 			log.WithFields(log.Fields{
// 				"workspace": workspace.Name,
// 				"block":     blockName,
// 			}).Infof("Successfully uninstalled block from workspace")

// 			// Now install the wanted version
// 			log.WithFields(log.Fields{
// 				"workspace": workspace.Name,
// 				"block":     blockName,
// 			}).Infof("Installing block from registry")

// 			registryBlockDir := filepath.Join(workspace.Path, workspace.Config.BlocksRoot, blockName)
// 			registryBlockVersion := blockVersion

// 			err = registryBlock.Install(registryBlockDir, registryBlockVersion)
// 			if err != nil {
// 				return err
// 			}

// 			// Now install the wanted version
// 			log.WithFields(log.Fields{
// 				"workspace": workspace.Name,
// 				"block":     blockName,
// 			}).Infof("Successfully installed block to workspace")
// 		}
// 	}
// 	return nil
// }

// func (o *Registry) InstallBlocks(args []string) error {
// 	for _, arg := range args {
// 		blockName, blockVersion, err := o.resolveArg(arg)
// 		if err != nil {
// 			return err
// 		}

// 		block := workspace.GetBlockFromIndex(blockName)
// 		if block != nil {
// 			// A block exists already
// 			// Let's check if it has a workdir
// 			if block.Workdir.LocalPath != "" {
// 				// The block has a workdir
// 				// Let's check if it exists
// 				if _, err := os.Stat(block.Workdir.LocalPath); !os.IsNotExist(err) {
// 					// The workdir exists
// 					// We're done here
// 					log.WithFields(log.Fields{
// 						"workspace": workspace.Name,
// 						"block":     block.Name,
// 						"path":      block.Workdir.LocalPath,
// 					}).Infof("Block is already installed. Use 'polycrate block update %s'", block.Name)
// 				} else {
// 					// The workdir does not exist
// 					// We can download the block
// 					download = true
// 				}
// 			} else {
// 				download = true
// 			}
// 			block.Inspect()
// 		} else {
// 			download = true
// 		}

// 		// Search block in registry
// 		if download {
// 			log.WithFields(log.Fields{
// 				"workspace": workspace.Name,
// 				"block":     blockName,
// 			}).Infof("Installing block from registry")

// 			registryBlock, err := registry.GetBlock(blockName)
// 			if err != nil {
// 				log.WithFields(log.Fields{
// 					"workspace": workspace.Name,
// 					"block":     blockName,
// 				}).Debug(err)
// 				return err
// 			}

// 			registryBlockDir := filepath.Join(workspace.Path, workspace.Config.BlocksRoot, blockName)
// 			registryBlockVersion := blockVersion

// 			err = registryBlock.Install(registryBlockDir, registryBlockVersion)
// 			if err != nil {
// 				log.WithFields(log.Fields{
// 					"workspace": workspace.Name,
// 					"block":     blockName,
// 					"version":   registryBlockVersion,
// 					"path":      registryBlockDir,
// 				}).Debug(err)
// 				return err
// 			}

// 			log.WithFields(log.Fields{
// 				"workspace": workspace.Name,
// 				"block":     blockName,
// 				"version":   registryBlockVersion,
// 				"path":      registryBlockDir,
// 			}).Infof("Block installed")
// 		}

// 		// - Accepts 1 arg: the block name/slug as it is in the registry
// 		// - Accepts 1 flag: version; defaults to latest
// 		// - Checks if a block with that name exists already AND has a block dir
// 		// - If the block exists, the command fails with a warning and shows a hint to the update command
// 		// - If no block exists, looks up the name of the block via Wordpress API at polycrate.io
// 		// - If a block is found, gets the list of releases
// 		// - Marks the youngest release as "latest"
// 		// - Downloads the release bundle
// 		// - If download succeeds, creates a block dir for the block
// 		// - unpacks the release bundle to the block dir

// 	}
// 	return nil
// }

// func (o *RegistryBlock) Install(blockDir string, version string) error {
// 	log.WithFields(log.Fields{
// 		"workspace": workspace.Name,
// 		"block":     o.BlockName,
// 		"version":   version,
// 	}).Debugf("Installing block from registry")

// 	// Get the correct release
// 	release, err := o.GetRelease(version)
// 	if err != nil {
// 		return err
// 	}

// 	// Get download url
// 	downloadUrl := release.ReleaseBundle

// 	// Create temp file
// 	releaseBundle, err := ioutil.TempFile("/tmp", "polycrate-block-"+o.Slug+"-"+release.Version+"-*.zip")
// 	if err != nil {
// 		return err
// 	}

// 	// Download to tempfile
// 	log.WithFields(log.Fields{
// 		"workspace": workspace.Name,
// 		"block":     o.BlockName,
// 		"path":      releaseBundle.Name(),
// 		"version":   release.Version,
// 		"url":       downloadUrl,
// 	}).Debugf("Downloading release bundle")

// 	// Download to tempfile
// 	err = DownloadFile(downloadUrl, releaseBundle.Name())
// 	if err != nil {
// 		return err
// 	}
// 	defer os.Remove(releaseBundle.Name())

// 	// Unpack
// 	log.WithFields(log.Fields{
// 		"workspace": workspace.Name,
// 		"block":     o.BlockName,
// 		"dst":       blockDir,
// 		"src":       releaseBundle.Name(),
// 	}).Debugf("Unpacking release bundle")

// 	// Unpack
// 	err = unzipSource(releaseBundle.Name(), blockDir)
// 	if err != nil {
// 		return err
// 	}

// 	log.WithFields(log.Fields{
// 		"workspace": workspace.Name,
// 		"block":     o.BlockName,
// 		"version":   release.Version,
// 	}).Debugf("Successfully installed block to workspace")

// 	return nil
// }

// func (o *RegistryWorkspace) Install(workspaceDir string, version string) error {
// 	log.WithFields(log.Fields{
// 		"workspace": o.WorkspaceName,
// 		"version":   version,
// 	}).Debugf("Installing workspace from registry")

// 	// Get the correct release
// 	release, err := o.GetRelease(version)
// 	if err != nil {
// 		return err
// 	}

// 	// Get download url
// 	downloadUrl := release.ReleaseBundle

// 	// Create temp file
// 	releaseBundle, err := ioutil.TempFile("/tmp", "polycrate-workspace-"+o.Slug+"-"+release.Version+"-*.zip")
// 	if err != nil {
// 		return err
// 	}

// 	// Download to tempfile
// 	log.WithFields(log.Fields{
// 		"workspace": o.WorkspaceName,
// 		"path":      releaseBundle.Name(),
// 		"version":   release.Version,
// 		"url":       downloadUrl,
// 	}).Debugf("Downloading release bundle")
// 	err = DownloadFile(downloadUrl, releaseBundle.Name())
// 	if err != nil {
// 		return err
// 	}
// 	defer os.Remove(releaseBundle.Name())

// 	// Unpack
// 	log.WithFields(log.Fields{
// 		"workspace": o.WorkspaceName,
// 		"dst":       workspaceDir,
// 		"src":       releaseBundle.Name(),
// 	}).Debugf("Unpacking release bundle")
// 	err = unzipSource(releaseBundle.Name(), workspaceDir)
// 	if err != nil {
// 		return err
// 	}

// 	log.WithFields(log.Fields{
// 		"workspace": o.WorkspaceName,
// 		"version":   release.Version,
// 	}).Debugf("Successfully installed workspace")

// 	return nil
// }
