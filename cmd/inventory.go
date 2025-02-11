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
	"fmt"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// installCmd represents the install command
var inventoryCmd = &cobra.Command{
	// Hidden: true,
	Use:    "inventory",
	Short:  "Inventory",
	Long:   `Inventory`,
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		_w := cmd.Flags().Lookup("workspace").Value.String()

		tx := polycrate.Transaction().SetCommand(cmd)
		defer tx.Stop()

		workspace, err := polycrate.LoadWorkspace(tx, _w, true)
		if err != nil {
			tx.Log.Fatal(err)
		}

		for path, _ := range workspace.FindInventories(tx) {
			fmt.Printf("%s\n", path)
		}

		showInventory(workspace)
	},
}

func init() {
	rootCmd.AddCommand(inventoryCmd)
	//initCmd.Flags().StringVar(&fileUrl, "template", "https://gitlab.com/ayedocloudsolutions/cloudstack/cloudstack/-/raw/main/examples/hcloud-single-node/Stackfile", "Stackfile to init from")
}

func showInventory(workspace *Workspace) {
	inv := AnsibleInventory{
		All: AnsibleGroup{
			Vars: map[string]interface{}{},
		},
	}

	var inv2 AnsibleInventory
	err := inv2.Load(filepath.Join(workspace.LocalPath, "inventory.yml"))
	if err != nil {
		log.Fatal(err)
	}
	printObject(inv)
	printObject(inv2)
	printObject(inv2.All.Hosts)
}

type AnsibleHost struct {
	AnsibleHost    string `mapstructure:"ansible_host" json:"ansible_host" yaml:"ansible_host"`
	AnsibleSshPort string `mapstructure:"ansible_ssh_port" json:"ansible_ssh_port" yaml:"ansible_ssh_port"`
	AnsibleUser    string `mapstructure:"ansible_user" json:"ansible_user" yaml:"ansible_user"`
	AnsibleBecome  string `mapstructure:"ansible_become" json:"ansible_become" yaml:"ansible_become"`
}

type AnsibleGroup struct {
	Vars     map[string]interface{}   `mapstructure:"vars" json:"vars" yaml:"vars"`
	Hosts    map[string]*AnsibleHost  `mapstructure:"hosts" json:"hosts" yaml:"hosts"`
	Children map[string]*AnsibleGroup `mapstructure:"children" json:"children" yaml:"children"`
}

type AnsibleInventory struct {
	All AnsibleGroup `mapstructure:"all" json:"all" yaml:"all" validate:"required"`
}

type PolycrateInventory struct {
	Hosts  map[string]*PolycrateInventoryHost
	Groups map[string][]*PolycrateInventoryHost
}

type PolycrateInventoryHost struct {
	Name string
	Host string
	Port string
	User string
}

func (i *AnsibleInventory) Load(path string) error {
	inventory := viper.New()
	inventory.SetConfigType("yaml")
	inventory.SetConfigFile(path)
	if err := inventory.MergeInConfig(); err != nil {
		log.Error(err)
	}

	if err := inventory.Unmarshal(i); err != nil {
		log.Error(err)
	}
	return nil
}

func (i *AnsibleInventory) GetHost(hostname string) *PolycrateInventoryHost {
	for key, host := range i.All.Hosts {
		if key == hostname {
			ph := PolycrateInventoryHost{
				Name: key,
				Host: host.AnsibleHost,
				Port: host.AnsibleSshPort,
				User: host.AnsibleUser,
			}
			return &ph
		}
	}
	return nil
}

func (ag *AnsibleGroup) GetHosts() []*PolycrateInventoryHost {
	hosts := []*PolycrateInventoryHost{}
	for key, host := range ag.Hosts {
		ph := PolycrateInventoryHost{
			Name: key,
			Host: host.AnsibleHost,
			Port: host.AnsibleSshPort,
			User: host.AnsibleUser,
		}
		hosts = append(hosts, &ph)
	}

	for _, group := range ag.Children {
		hosts = append(hosts, group.GetHosts()...)
	}
	return hosts
}

func (pi *PolycrateInventory) Load(ai *AnsibleInventory) error {
	for _, host := range ai.All.GetHosts() {
		pi.Hosts[host.Name] = host
	}
	return nil
}
