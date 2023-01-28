// curl -X 'GET' \
//   'https://cargo.ayedo.cloud/api/v2.0/projects/ayedo/repositories?page=1&page_size=10' \
//   -H 'accept: application/json'
package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type HarborRepository struct {
	Name          string  `yaml:"name,omitempty" mapstructure:"name,omitempty" json:"name,omitempty" validate:"required"`
	ArtifactCount float64 `yaml:"artifact_count,omitempty" mapstructure:"artifact_count,omitempty" json:"artifact_count,omitempty"`
	CreationTime  string  `yaml:"creation_time,omitempty" mapstructure:"creation_time,omitempty" json:"creation_time,omitempty"`
	ID            float64 `yaml:"id,omitempty" mapstructure:"id,omitempty" json:"id,omitempty"`
	ProjectID     float64 `yaml:"project_id,omitempty" mapstructure:"project_id,omitempty" json:"project_id,omitempty"`
	PullCount     float64 `yaml:"pull_count,omitempty" mapstructure:"pull_count,omitempty" json:"pull_count,omitempty"`
	UpdateTime    string  `yaml:"update_time,omitempty" mapstructure:"update_time,omitempty" json:"update_time,omitempty"`
}

var cargoDocsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Generate docs from the available registry blocks",
	Long:  ``,
	Args:  cobra.ExactArgs(0), // https://github.com/spf13/cobra/blob/master/user_guide.md
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := http.Get("https://cargo.ayedo.cloud/api/v2.0/projects/ayedo/repositories?page=1&page_size=10")
		if err != nil {
			log.Fatalln(err)
		}

		var repositoryList []HarborRepository
		err = json.NewDecoder(resp.Body).Decode(&repositoryList)
		if err != nil {
			log.Fatalln(err)
		}

		for _, repository := range repositoryList {
			fmt.Printf("name: %s, artifacts: %.0f, pulls: %.0f\n", repository.Name, repository.ArtifactCount, repository.PullCount)
		}

		return nil
	},
}

func init() {
	cargoCmd.AddCommand(cargoDocsCmd)
}
