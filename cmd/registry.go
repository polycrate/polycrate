package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"

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
	Releases  []RegistryRelease `yaml:"releases,omitempty" mapstructure:",omitempty" json:",omitempty"`
	BlockName string            `yaml:"block_name" mapstructure:"block_name" json:"block_name" validate:"required"`
}

type Registry struct {
	Url      string `yaml:"url" mapstructure:"url" json:"url" validate:"required"`
	ApiBase  string `yaml:"api_base" mapstructure:"api_base" json:"api_base" validate:"required"`
	Username string `yaml:"username,omitempty" mapstructure:"username,omitempty" json:"username,omitempty"`
	Password string `yaml:"password,omitempty" mapstructure:"password,omitempty" json:"password,omitempty"`
}

func (o *Registry) SearchBlock(blockName string) ([]RegistryBlock, error) {
	url := fmt.Sprintf("%s/%s/block?search=%s", o.Url, o.ApiBase, blockName)

	response, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var blocks []RegistryBlock
	err = json.Unmarshal(responseData, &blocks)
	if err != nil {
		return nil, err
	}
	log.Debug(string(responseData))

	if len(blocks) > 0 {
		// Sort the Releases inside of a Block by date
		for _, block := range blocks {
			sort.Slice(block.Releases, func(i, j int) bool {
				return block.Releases[i].PostDate > block.Releases[j].PostDate
			})
		}

		return blocks, nil
	} else {
		err := fmt.Errorf("Block not found: %s", blockName)
		return nil, err
	}

}

func (o *Registry) GetBlock(blockName string) (*RegistryBlock, error) {
	registryBlocks, err := o.SearchBlock(blockName)
	if err != nil {
		return nil, err
	}

	blockIndex := slices.IndexFunc(registryBlocks, func(o RegistryBlock) bool { return o.BlockName == blockName })
	if blockIndex != -1 {
		return &registryBlocks[blockIndex], nil
	} else {
		err := fmt.Errorf("Block not found: %s", blockName)
		return nil, err
	}

}

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

func (o *Registry) resolveArg(arg string) (string, string, error) {
	var blockName string
	var blockVersion string

	// Split arg by :
	splitArgs := strings.Split(arg, ":")

	// length can be either 1 (BLOCK) or 2 (BLOCK:VERSION)
	// Everything else is an error
	switch len(splitArgs) {
	case 1:
		// it's only the block
		blockName = splitArgs[0]
		blockVersion = "latest"
	case 2:
		blockName = splitArgs[0]
		blockVersion = splitArgs[1]
	default:
		err := fmt.Errorf("Wrong format: %s - should be 'block:version' as in 'test:0.0.1'", arg)
		return "", "", err
	}
	return blockName, blockVersion, nil
}

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

func (o *RegistryBlock) Install(blockDir string, version string) error {
	log.WithFields(log.Fields{
		"workspace": workspace.Name,
		"block":     o.BlockName,
		"version":   version,
	}).Debugf("Installing block from registry")

	// Get the correct release
	release, err := o.GetRelease(version)
	if err != nil {
		return err
	}

	// Get download url
	downloadUrl := release.ReleaseBundle

	log.WithFields(log.Fields{
		"workspace": workspace.Name,
		"block":     o.BlockName,
		"path":      blockDir,
		"version":   release.Version,
		"url":       downloadUrl,
	}).Debugf("Installing block from registry")

	// Create temp file
	releaseBundle, err := ioutil.TempFile("/tmp", "polycrate-block-"+o.Slug+"-"+release.Version+"-*.zip")
	if err != nil {
		return err
	}

	// Download to tempfile
	err = DownloadFile(downloadUrl, releaseBundle.Name())
	if err != nil {
		return err
	}
	defer os.Remove(releaseBundle.Name())

	log.WithFields(log.Fields{
		"workspace": workspace.Name,
		"block":     o.BlockName,
		"path":      releaseBundle.Name(),
		"version":   release.Version,
		"url":       downloadUrl,
	}).Debugf("Downloaded release bundle")

	// Unpack
	err = unzipSource(releaseBundle.Name(), blockDir)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"workspace": workspace.Name,
		"block":     o.BlockName,
		"path":      blockDir,
		"version":   release.Version,
		"url":       downloadUrl,
	}).Debugf("Unpacked release bundle")

	log.WithFields(log.Fields{
		"workspace": workspace.Name,
		"block":     o.BlockName,
		"version":   version,
	}).Infof("Successfully installed block to workspace")

	return nil
}
