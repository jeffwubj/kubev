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

	"github.com/jeffwubj/kubev/pkg/kubev/constants"
	"github.com/jeffwubj/kubev/pkg/kubev/deployer"
	"github.com/jeffwubj/kubev/pkg/kubev/model"
	"github.com/jeffwubj/kubev/pkg/kubev/utils"
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
	"resourcepool":      "Resource pool to hold Kubernetese nodes, input none to not use resource pool",
	"folder":            "VM Folder name",
	"cpu":               "Number of vCPUs for each VM, at least 2",
	"memory":            "Memory for each VM (MB)",
	"network":           "Network for each VM, default [VM Network]",
	"kubernetesversion": "Kubernetes version, e.g. [v1.13.0]",
	"workernodes":       "Worker nodes number",
}

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure a Kubernetes installation on vSphere or ESX.",
	Long:  `The whole procedure includes two phases: 1st config, 2nd deploy`,
	Run:   runConfig,
}

var basicqs = []*survey.Question{
	{
		Name:     "serverurl",
		Prompt:   &survey.Input{Message: descriptions["serverurl"]},
		Validate: survey.Required,
	},
	{
		Name:   "port",
		Prompt: &survey.Input{Message: descriptions["port"], Default: constants.DefaultRemotePort},
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
}

var vcenterqs = []*survey.Question{
	{
		Name:     "datacenter",
		Prompt:   &survey.Input{Message: descriptions["datacenter"], Default: constants.DefaultRemoteDatacenter},
		Validate: survey.Required,
	},
	{
		Name:     "datastore",
		Prompt:   &survey.Input{Message: descriptions["datastore"]},
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
		Prompt:   &survey.Input{Message: descriptions["cpu"], Default: constants.DefaultRemoteCPU},
		Validate: survey.Required,
	},
	{
		Name:     "memory",
		Prompt:   &survey.Input{Message: descriptions["memory"], Default: constants.DefaultRemoteMemory},
		Validate: survey.Required,
	},
	{
		Name:     "network",
		Prompt:   &survey.Input{Message: descriptions["network"], Default: constants.DefaultRemoteNetwork},
		Validate: survey.Required,
	},
	{
		Name:     "kubernetesversion",
		Prompt:   &survey.Input{Message: descriptions["kubernetesversion"], Default: constants.DefaultKubernetesVersion},
		Validate: survey.Required,
	},
	{
		Name:     "workernodes",
		Prompt:   &survey.Input{Message: descriptions["workernodes"], Default: constants.DefaultKubernetesWorkderNodeNum},
		Validate: survey.Required,
	},
}

var esxqs = []*survey.Question{
	{
		Name:     "datastore",
		Prompt:   &survey.Input{Message: descriptions["datastore"]},
		Validate: survey.Required,
	},
	{
		Name:     "cpu",
		Prompt:   &survey.Input{Message: descriptions["cpu"], Default: constants.DefaultRemoteCPU},
		Validate: survey.Required,
	},
	{
		Name:     "memory",
		Prompt:   &survey.Input{Message: descriptions["memory"], Default: constants.DefaultRemoteMemory},
		Validate: survey.Required,
	},
	{
		Name:     "network",
		Prompt:   &survey.Input{Message: descriptions["network"], Default: constants.DefaultRemoteNetwork},
		Validate: survey.Required,
	},
	{
		Name:     "kubernetesversion",
		Prompt:   &survey.Input{Message: descriptions["kubernetesversion"], Default: constants.DefaultKubernetesVersion},
		Validate: survey.Required,
	},
	{
		Name:     "workernodes",
		Prompt:   &survey.Input{Message: descriptions["workernodes"], Default: constants.DefaultKubernetesWorkderNodeNum},
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
	configCmd.Flags().String("resourcepool", "", descriptions["resourcepool"])
	configCmd.Flags().String("folder", "", descriptions["folder"])
	configCmd.Flags().Int("cpu", 2, descriptions["cpu"])
	configCmd.Flags().Int("memory", 2048, descriptions["memory"])
	configCmd.Flags().String("network", "", descriptions["network"])
	configCmd.Flags().String("kubernetesversion", "", descriptions["kubernetesversion"])
	configCmd.Flags().Int("workernodes", 5, descriptions["workernodes"])
	viper.BindPFlags(configCmd.Flags())
}

func runConfig(cmd *cobra.Command, args []string) {
	_, err := interactiveSetConfig()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

func interactiveSetConfig() (*model.Answers, error) {
	// perform the questions
	answers := &model.Answers{}
	err := survey.Ask(basicqs, answers)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	if err := deployer.ValidatevSphereAccount(answers); err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	if answers.IsVCenter {
		err := survey.Ask(vcenterqs, answers)
		if err != nil {
			fmt.Println(err.Error())
			return nil, err
		}
	} else {
		err := survey.Ask(esxqs, answers)
		if err != nil {
			fmt.Println(err.Error())
			return nil, err
		}
		answers.Datacenter = "ha-datacenter"
	}

	if utils.FileExists(constants.GetK8sNodesConfigFilePath()) {
		save := false
		survey.AskOne(&survey.Confirm{
			Message: "If there is local cluster information, it will be deleted, continue?",
			Default: true,
		}, &save, nil)
		if !save {
			fmt.Println("Bye")
			return nil, fmt.Errorf("Canceled")
		}
		utils.DeleteFile(constants.GetK8sNodesConfigFilePath())
	}

	SaveAnswers(answers)

	return answers, nil
}

func SaveAnswers(answers *model.Answers) {
	viper.Set("serverurl", answers.Serverurl)
	viper.Set("port", answers.Port)
	viper.Set("username", answers.Username)
	viper.Set("password", answers.Password)
	viper.Set("datacenter", answers.Datacenter)
	viper.Set("datastore", answers.Datastore)
	viper.Set("resourcepool", answers.Resourcepool)
	viper.Set("folder", answers.Folder)
	viper.Set("cpu", answers.Cpu)
	viper.Set("memory", answers.Memory)
	viper.Set("network", answers.Network)
	viper.Set("kubernetesVersion", answers.KubernetesVersion)
	viper.Set("workernodes", answers.WorkerNodes)
	viper.Set("isvcenter", answers.IsVCenter)
	viper.WriteConfigAs(viper.ConfigFileUsed())
}
