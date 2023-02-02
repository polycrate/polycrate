// curl -X 'GET' \
//   'https://cargo.ayedo.cloud/api/v2.0/projects/ayedo/repositories?page=1&page_size=10' \
//   -H 'accept: application/json'
package cmd

import (
	"context"
	_ "embed"

	// "encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	semver "github.com/hashicorp/go-version"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

/*

"extra_attrs": {
		"architecture": "amd64",
		"author": "",
		"config": {
			"Env": [
				"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
			],
			"Labels": {
				"ayedo.cruise.app": "custom-app",
				"ayedo.product.name": "cruise",
				"ayedo.product.url": "https://www.ayedo.de/products/cruise/managed-custom-app/"
			},
			"WorkingDir": "/"
		},
		"created": "0001-01-01T00:00:00Z",
		"os": "linux"
	},
*/

type HarborRepository struct {
	Name          string  `yaml:"name,omitempty" mapstructure:"name,omitempty" json:"name,omitempty" validate:"required"`
	ArtifactCount float64 `yaml:"artifact_count,omitempty" mapstructure:"artifact_count,omitempty" json:"artifact_count,omitempty"`
	CreationTime  string  `yaml:"creation_time,omitempty" mapstructure:"creation_time,omitempty" json:"creation_time,omitempty"`
	ID            float64 `yaml:"id,omitempty" mapstructure:"id,omitempty" json:"id,omitempty"`
	ProjectID     float64 `yaml:"project_id,omitempty" mapstructure:"project_id,omitempty" json:"project_id,omitempty"`
	PullCount     float64 `yaml:"pull_count,omitempty" mapstructure:"pull_count,omitempty" json:"pull_count,omitempty"`
	UpdateTime    string  `yaml:"update_time,omitempty" mapstructure:"update_time,omitempty" json:"update_time,omitempty"`
}

type HarborRepositoryArtifactTag struct {
	ID           float64                  `yaml:"id,omitempty" mapstructure:"id,omitempty" json:"id,omitempty"`
	ArtifactID   float64                  `yaml:"artifact_id,omitempty" mapstructure:"artifact_id,omitempty" json:"artifact_id,omitempty"`
	Artifact     HarborRepositoryArtifact `yaml:"artifact,omitempty" mapstructure:"artifact,omitempty" json:"artifact,omitempty"`
	Name         string                   `yaml:"name,omitempty" mapstructure:"name,omitempty" json:"name,omitempty" validate:"required"`
	Version      string                   `yaml:"version,omitempty" mapstructure:"version,omitempty" json:"version,omitempty" validate:"required"`
	PullTime     string                   `yaml:"pull_time,omitempty" mapstructure:"pull_time,omitempty" json:"pull_time,omitempty"`
	PushTime     string                   `yaml:"push_time,omitempty" mapstructure:"push_time,omitempty" json:"push_time,omitempty"`
	PullCount    float64                  `yaml:"pull_count,omitempty" mapstructure:"pull_count,omitempty" json:"pull_count,omitempty"`
	Signed       bool                     `yaml:"signed,omitempty" mapstructure:"signed,omitempty" json:"signed,omitempty"`
	RepositoryID float64                  `yaml:"repository_id,omitempty" mapstructure:"repository_id,omitempty" json:"repository_id,omitempty"`
	Repository   HarborRepository         `yaml:"repository,omitempty" mapstructure:"repository,omitempty" json:"repository,omitempty"`
	Readme       string
	Registry     string `yaml:"registry,omitempty" mapstructure:"registry,omitempty" json:"registry,omitempty"`
	Project      string `yaml:"project,omitempty" mapstructure:"project,omitempty" json:"project,omitempty"`
	HasReadme    bool   `yaml:"has_readme,omitempty" mapstructure:"has_readme,omitempty" json:"has_readme,omitempty"`
	Digest       string `string:"digest,omitempty" mapstructure:"digest,omitempty" json:"digest,omitempty"`
	//ProjectID     float64 `yaml:"project_id,omitempty" mapstructure:"project_id,omitempty" json:"project_id,omitempty"`
}

type HarbarRepositoryArtifactExtraAttrsConfig struct {
	Env        []string          `yaml:"Env,omitempty" mapstructure:"Env,omitempty" json:"Env,omitempty"`
	Labels     map[string]string `yaml:"Labels,omitempty" mapstructure:"Labels,omitempty" json:"Labels,omitempty"`
	WorkingDir string            `yaml:"WorkingDir,omitempty" mapstructure:"WorkingDir,omitempty" json:"WorkingDir,omitempty"`
}

type HarbarRepositoryArtifactExtraAttrs struct {
	Architecture string                                   `yaml:"architecture,omitempty" mapstructure:"architecture,omitempty" json:"architecture,omitempty"`
	Author       string                                   `yaml:"author,omitempty" mapstructure:"author,omitempty" json:"author,omitempty"`
	Os           string                                   `yaml:"os,omitempty" mapstructure:"os,omitempty" json:"os,omitempty"`
	Config       HarbarRepositoryArtifactExtraAttrsConfig `yaml:"config,omitempty" mapstructure:"config,omitempty" json:"config,omitempty"`
}

type HarborRepositoryArtifact struct {
	ID           float64                            `yaml:"id,omitempty" mapstructure:"id,omitempty" json:"id,omitempty"`
	Name         string                             `yaml:"name,omitempty" mapstructure:"name,omitempty" json:"name,omitempty" validate:"required"`
	Digest       string                             `string:"digest,omitempty" mapstructure:"digest,omitempty" json:"digest,omitempty"`
	CreationTime string                             `yaml:"creation_time,omitempty" mapstructure:"creation_time,omitempty" json:"creation_time,omitempty"`
	ProjectID    float64                            `yaml:"project_id,omitempty" mapstructure:"project_id,omitempty" json:"project_id,omitempty"`
	PullTime     string                             `yaml:"pull_time,omitempty" mapstructure:"pull_time,omitempty" json:"pull_time,omitempty"`
	PushTime     string                             `yaml:"push_time,omitempty" mapstructure:"push_time,omitempty" json:"push_time,omitempty"`
	Icon         string                             `yaml:"icon,omitempty" mapstructure:"icon,omitempty" json:"icon,omitempty"`
	Labels       map[string]string                  `yaml:"labels,omitempty" mapstructure:"labels,omitempty" json:"labels,omitempty"`
	PullCount    float64                            `yaml:"pull_count,omitempty" mapstructure:"pull_count,omitempty" json:"pull_count,omitempty"`
	UpdateTime   string                             `yaml:"update_time,omitempty" mapstructure:"update_time,omitempty" json:"update_time,omitempty"`
	Type         string                             `yaml:"type,omitempty" mapstructure:"type,omitempty" json:"type,omitempty"`
	Size         float64                            `yaml:"size,omitempty" mapstructure:"size,omitempty" json:"size,omitempty"`
	RepositoryID float64                            `yaml:"repository_id,omitempty" mapstructure:"repository_id,omitempty" json:"repository_id,omitempty"`
	Repository   HarborRepository                   `yaml:"repository,omitempty" mapstructure:"repository,omitempty" json:"repository,omitempty"`
	Tags         []HarborRepositoryArtifactTag      `yaml:"tags,omitempty" mapstructure:"tags,omitempty" json:"tags,omitempty"`
	ExtraAttrs   HarbarRepositoryArtifactExtraAttrs `yaml:"extra_attrs,omitempty" mapstructure:"extra_attrs,omitempty" json:"extra_attrs,omitempty"`
}

type DocsBlocksCatalog struct {
	Count  int                       `yaml:"count,omitempty" mapstructure:"count,omitempty" json:"count,omitempty"`
	Blocks []*DocsBlocksCatalogBlock `yaml:"blocks,omitempty" mapstructure:"blocks,omitempty" json:"blocks,omitempty"`
}
type DocsBlocksCatalogBlock struct {
	Name                  string `yaml:"name,omitempty" mapstructure:"name,omitempty" json:"name,omitempty"`
	Registry              string `yaml:"registry,omitempty" mapstructure:"registry,omitempty" json:"registry,omitempty"`
	Project               string `yaml:"project,omitempty" mapstructure:"project,omitempty" json:"project,omitempty"`
	LatestVersion         string `yaml:"latest_version,omitempty" mapstructure:"latest_version,omitempty" json:"latest_version,omitempty"`
	HasReadme             bool   `yaml:"has_readme,omitempty" mapstructure:"has_readme,omitempty" json:"has_readme,omitempty"`
	Repository            *HarborRepository
	Versions              []*HarborRepositoryArtifactTag `yaml:"versions,omitempty" mapstructure:"versions,omitempty" json:"versions,omitempty"`
	docsBlockPath         string
	docsBlockReleasesPath string
}

var cargoDocsCmd = &cobra.Command{
	Hidden: true,
	Use:    "docs",
	Short:  "Generate docs from the available registry blocks",
	Long:   ``,
	Args:   cobra.ExactArgs(0), // https://github.com/spf13/cobra/blob/master/user_guide.md
	RunE: func(cmd *cobra.Command, args []string) error {
		catalog := DocsBlocksCatalog{}
		registry := "cargo.ayedo.cloud"
		project := "ayedo"

		log.Info("1.0 Get blocks from registry")

		// Get repositories from registry
		// A repository equals a block, e.g. `ayedo/k8s/hairpin-proxy`
		regpositoryListURL := fmt.Sprintf("https://%s/api/v2.0/projects/%s/repositories?page=1&page_size=100", registry, project)
		resp, err := http.Get(regpositoryListURL)
		if err != nil {
			log.Fatal(err)
		}

		// unmarshal json to struct
		var repositoryList []HarborRepository
		if err := json.NewDecoder(resp.Body).Decode(&repositoryList); err != nil {
			log.Fatal(err)
		}

		log.Infof("Found %d blocks", len(repositoryList))

		// the docs/blocks directory should be created manually for awareness
		basePath := "docs/blocks"
		if _, err := os.Stat(basePath); os.IsNotExist(err) {
			log.Fatal("'docs/blocks' does not exist")
		}

		ctx := context.Background()
		ctx, cancel, err := polycrate.NewTransaction(ctx, cmd)
		defer polycrate.StopTransaction(ctx, cancel)
		if err != nil {
			log.Fatal(err)
		}

		// Loop over repositories
		for _, repository := range repositoryList {
			// Initialize block for catalog
			catalogBlock := DocsBlocksCatalogBlock{}

			// Add pointer to block to catalog
			catalog.Blocks = append(catalog.Blocks, &catalogBlock)

			catalogBlock.Registry = registry
			catalogBlock.Project = project
			catalogBlock.Name = repository.Name
			catalogBlock.Repository = &repository
			//blockName := repository.Name

			catalogBlock.docsBlockPath = filepath.Join([]string{basePath, catalogBlock.Name}...)
			catalogBlock.docsBlockReleasesPath = filepath.Join([]string{catalogBlock.docsBlockPath, "releases"}...)

			// create block path
			if err := os.MkdirAll(catalogBlock.docsBlockReleasesPath, os.ModePerm); err != nil {
				log.Fatal(err)
			}

			// // Create releases subdir
			// if err := os.MkdirAll(catalogBlock.docsBlockReleasesPath, os.ModePerm); err != nil {
			// 	log.Fatal(err)
			// }

			// Get repository artifacts
			// 1 artifact equals to 1 release of a block (e.g. 0.0.1 is an artifact and 0.0.2 is another)
			// 1 artifact can have multiple labels though; with regard to Polycrate, artifacts usually have 1 tag
			// Exception: the latest release always also has the "latest" tag
			repositoryNameSplit := strings.Split(repository.Name, "/")[1:]
			repositoryReference := strings.Join(repositoryNameSplit, "/")
			url := fmt.Sprintf("https://%s/api/v2.0/projects/%s/repositories/%s/artifacts?page=1&page_size=100", catalogBlock.Registry, catalogBlock.Project, url.QueryEscape(repositoryReference))
			resp, err := http.Get(url)
			if err != nil {
				log.Fatalln(err)
			}

			// Unmarshal artifacts
			var repositoryArtifactList []HarborRepositoryArtifact
			err = json.NewDecoder(resp.Body).Decode(&repositoryArtifactList)
			if err != nil {
				log.Fatalln(err)
			}

			// Loop over artifacts
			for _, artifact := range repositoryArtifactList {

				downloadPath := "dist/docs/blocks"

				if err := os.MkdirAll(downloadPath, os.ModePerm); err != nil {
					log.Fatal(err)
				}

				for _, tag := range artifact.Tags {
					tag.Artifact = artifact
					tag.Repository = repository
					tag.Digest = artifact.Digest
					tag.PullCount = artifact.PullCount
					tag.Version = tag.Name
					tag.Registry = catalogBlock.Registry
					tag.Project = catalogBlock.Project
					//blockVersion := tag.Name

					if tag.Version != "latest" {
						// Append to block catalog
						catalogBlock.Versions = append(catalogBlock.Versions, &tag)

						// Set latest version if not yet given
						if catalogBlock.LatestVersion == "" {
							catalogBlock.LatestVersion = tag.Version
						}

						// Define path to download the block to
						blockDownloadPath := filepath.Join([]string{downloadPath, catalogBlock.Name, tag.Version}...)

						// Download the block from the registry if it doesn't exist already
						if _, err := os.Stat(blockDownloadPath); os.IsNotExist(err) {
							if err := os.MkdirAll(blockDownloadPath, os.ModePerm); err != nil {
								log.Fatal(err)
							}

							log.Debugf("Downloading %s/%s:%s to %s", catalogBlock.Registry, catalogBlock.Name, tag.Version, blockDownloadPath)

							if err := UnwrapOCIImage(ctx, blockDownloadPath, catalogBlock.Registry, catalogBlock.Name, tag.Version); err != nil {
								log.Fatal(err)
							}
						}

						// if a README.md exists, copy it over
						downloadsReadmeFile := filepath.Join([]string{blockDownloadPath, "README.md"}...)
						if _, err := os.Stat(downloadsReadmeFile); !os.IsNotExist(err) {
							// Extra confditions for old blocks with bad readmes
							buildReadme := true
							fromVersion, _ := semver.NewVersion("0.1.5")
							actualVersion, _ := semver.NewVersion(tag.Version)

							if catalogBlock.Name == "ayedo/hcloud/k8s" {
								if actualVersion.LessThan(fromVersion) {
									log.Warnf("Not adding block README for %s", catalogBlock.Name)
									buildReadme = false
								}
							}

							// The file exists
							// We now load the file as string and feed it to the block readme template
							if buildReadme {
								sourceBuffer, err := os.ReadFile(downloadsReadmeFile)
								if err != nil {
									log.Fatal(err)
								}
								tag.Readme = string(sourceBuffer)
							}
						}

						// Get the path for the target file in docs/
						docsReadmeFile := filepath.Join([]string{catalogBlock.docsBlockReleasesPath, strings.Join([]string{tag.Version, "md"}, ".")}...)

						// Load release template
						t, err := template.ParseFS(templateFiles, "templates/docs_block_release.md")
						if err != nil {
							log.Fatal(err)
						}

						// Create / Truncate target file
						f, err := os.Create(docsReadmeFile)
						if err != nil {
							log.Fatalf("create file: ", err)
						}
						defer f.Close()

						err = t.Execute(f, tag)
						if err != nil {
							log.Fatalf("execute: ", err)
						}
						f.Close()

						tag.HasReadme = true

					} else {
						log.Warn("Skipping 'latest' tag")
					}

				}

				// End of tags loop
				// Create README for block
				log.Infof("Generate README.md for docs/blocks/%s", catalogBlock.Name)
				docsBlockREADME := filepath.Join(basePath, catalogBlock.Name, "README.md")

				t, err := template.ParseFS(templateFiles, "templates/docs_block_readme.md")
				if err != nil {
					log.Fatal(err)
				}

				f, err := os.Create(docsBlockREADME)
				if err != nil {
					log.Fatalf("create file: ", err)
				}
				err = t.Execute(f, catalogBlock)
				if err != nil {
					log.Fatalf("execute: ", err)
				}
				f.Close()
				log.Infof("Created file at %s", f.Name())

			}
		}

		log.Info("Generate README.md for docs/blocks")
		docsBlocksMainREADME := filepath.Join(basePath, "README.md")

		t, err := template.ParseFS(templateFiles, "templates/docs_base_readme.md")
		if err != nil {
			log.Fatal(err)
		}

		f, err := os.Create(docsBlocksMainREADME)
		if err != nil {
			log.Fatalf("create file: ", err)
		}
		err = t.Execute(f, catalog)
		if err != nil {
			log.Fatalf("execute: ", err)
		}
		f.Close()
		log.Infof("Created file at %s", f.Name())

		//printObject(catalog)

		return nil
	},
}

func init() {
	cargoCmd.AddCommand(cargoDocsCmd)
}
