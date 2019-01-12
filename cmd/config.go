// Copyright © 2019 Jeff Wu <jeff.wu.junfei@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"

	"github.com/jeffwubj/kubev/pkg/kubev/model"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/AlecAivazis/survey.v1"
)

var descriptions = map[string]string{
	"serverurl":         "vCenter/ESX URL Ex: 10.192.10.30 or myvcenter.io",
	"port":              "vCenter/ESX port",
	"username":          "vCenter/ESX username",
	"password":          "vCenter/ESX password",
	"datacenter":        "Datacenter",
	"datastore":         "Datastore",
	"cluster":           "Cluster",
	"resourcepool":      "Resource pool to hold Kubernetese nodes, input none to not use resource pool",
	"folder":            "VM Folder name",
	"cpu":               "Number of vCPUs for each VM, at least 2",
	"memory":            "Memory for each VM (MB)",
	"network":           "Network for each VM, default [VM Network]",
	"kubernetesversion": "Kubernetes version, e.g. [v1.13.0]",
}

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure a Kubernetes installation on vSphere or ESX.",
	Long:  `The whole procedure includes two phases: 1st config, 2nd deploy`,
	Run:   runConfig,
}

var qs = []*survey.Question{
	{
		Name:     "serverurl",
		Prompt:   &survey.Input{Message: descriptions["serverurl"]},
		Validate: survey.Required,
	},
	{
		Name:   "port",
		Prompt: &survey.Input{Message: descriptions["port"]},
	},
	{
		Name:     "username",
		Prompt:   &survey.Input{Message: descriptions["username"]},
		Validate: survey.Required,
	},
	{
		Name:     "password",
		Prompt:   &survey.Input{Message: descriptions["password"]},
		Validate: survey.Required,
	},
	{
		Name:     "datacenter",
		Prompt:   &survey.Input{Message: descriptions["datacenter"]},
		Validate: survey.Required,
	},
	{
		Name:     "datastore",
		Prompt:   &survey.Input{Message: descriptions["datastore"]},
		Validate: survey.Required,
	},
	{
		Name:     "cluster",
		Prompt:   &survey.Input{Message: descriptions["cluster"]},
		Validate: survey.Required,
	},
	{
		Name:     "resourcepool",
		Prompt:   &survey.Input{Message: descriptions["resourcepool"]},
		Validate: survey.Required,
	},
	{
		Name:     "folder",
		Prompt:   &survey.Input{Message: descriptions["folder"]},
		Validate: survey.Required,
	},
	{
		Name:     "cpu",
		Prompt:   &survey.Input{Message: descriptions["cpu"]},
		Validate: survey.Required,
	},
	{
		Name:     "memory",
		Prompt:   &survey.Input{Message: descriptions["memory"]},
		Validate: survey.Required,
	},
	{
		Name:     "network",
		Prompt:   &survey.Input{Message: descriptions["network"]},
		Validate: survey.Required,
	},
	{
		Name:     "kubernetesversion",
		Prompt:   &survey.Input{Message: descriptions["kubernetesversion"]},
		Validate: survey.Required,
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.Flags().String("serverurl", "", descriptions["serverurl"])
	configCmd.Flags().Int("port", 443, descriptions["port"])
	configCmd.Flags().String("username", "", descriptions["username"])
	configCmd.Flags().String("password", "", descriptions["password"])
	configCmd.Flags().String("datacenter", "", descriptions["datacenter"])
	configCmd.Flags().String("datastore", "", descriptions["datastore"])
	configCmd.Flags().String("cluster", "", descriptions["cluster"])
	configCmd.Flags().String("resourcepool", "", descriptions["resourcepool"])
	configCmd.Flags().String("folder", "", descriptions["folder"])
	configCmd.Flags().Int("cpu", 2, descriptions["cpu"])
	configCmd.Flags().Int("memory", 2048, descriptions["memory"])
	configCmd.Flags().String("network", "", descriptions["network"])
	configCmd.Flags().String("kubernetesversion", "", descriptions["kubernetesversion"])
	viper.BindPFlags(configCmd.Flags())
}

func runConfig(cmd *cobra.Command, args []string) {
	interactiveSetConfig()
}

func interactiveSetConfig() model.Answers {
	// perform the questions
	answers := model.Answers{}
	err := survey.Ask(qs, &answers)
	if err != nil {
		fmt.Println(err.Error())
		return answers
	}

	viper.Set("serverurl", answers.Serverurl)
	viper.Set("port", answers.Port)
	viper.Set("username", answers.Username)
	viper.Set("password", answers.Password)
	viper.Set("datacenter", answers.Datacenter)
	viper.Set("datastore", answers.Datastore)
	viper.Set("cluster", answers.Cluster)
	viper.Set("resourcepool", answers.Resourcepool)
	viper.Set("folder", answers.Folder)
	viper.Set("cpu", answers.Cpu)
	viper.Set("memory", answers.Memory)
	viper.Set("network", answers.Network)
	viper.Set("kubernetesVersion", answers.KubernetesVersion)

	viper.WriteConfigAs("kubevconfig")
	viper.WriteConfigAs(viper.ConfigFileUsed())
	return answers
}
