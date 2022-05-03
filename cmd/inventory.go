/*
Copyright © 2021 Fabian Peter <fp@ayedo.de>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// installCmd represents the install command
var inventoryCmd = &cobra.Command{
	Hidden: true,
	Use:    "inventory",
	Short:  "Inventory",
	Long:   `Inventory`,
	Run: func(cmd *cobra.Command, args []string) {
		showInventory()
	},
}

func init() {
	rootCmd.AddCommand(inventoryCmd)
	//initCmd.Flags().StringVar(&fileUrl, "template", "https://gitlab.com/ayedocloudsolutions/cloudstack/cloudstack/-/raw/main/examples/hcloud-single-node/Stackfile", "Stackfile to init from")
}

func showInventory() {
	inv := Inventory{
		All: Group{
			Vars: map[string]interface{}{},
		},
	}

	var inv2 Inventory
	mainInventory := viper.New()
	mainInventory.SetConfigType("yaml")
	mainInventory.SetConfigFile(filepath.Join(workspace.Path, "inventory.yml"))
	if err := mainInventory.MergeInConfig(); err != nil {
		log.Error(err)
	}

	if err := mainInventory.UnmarshalExact(&inv2); err != nil {
		log.Error(err)
	}
	printObject(inv)
	printObject(inv2)
}

// Group represents ansible group
type Group struct {
	Vars     map[string]interface{} `mapstructure:"vars" json:"vars"`
	Hosts    map[string]interface{} `mapstructure:"hosts" json:"hosts"`
	Children map[string]*Group      `mapstructure:"children" json:"children"`
}

// Host represents ansible host

type Inventory struct {
	All Group `mapstructure:"all" json:"all" validate:"required"`
}
