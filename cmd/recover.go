// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
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

	"github.com/jeffwubj/kubev/pkg/kubev/cacher"
	"github.com/jeffwubj/kubev/pkg/kubev/constants"
	"github.com/jeffwubj/kubev/pkg/kubev/deployer"
	"github.com/jeffwubj/kubev/pkg/kubev/model"
	"github.com/jeffwubj/kubev/pkg/kubev/utils"
	"github.com/spf13/cobra"
	survey "gopkg.in/AlecAivazis/survey.v1"
)

// recoverCmd represents the recover command
var recoverCmd = &cobra.Command{
	Use:   "recover",
	Short: "Find Kubernetes cluster deployed by kubev in your vSphere/ESX",
	Long:  "Find Kubernetes cluster deployed by kubev in your vSphere/ESX, by recovering a configuration from cluster, your local 'kubev config' result will be overwritten",
	Run:   runRecover,
}

func init() {
	rootCmd.AddCommand(recoverCmd)
}

func runRecover(cmd *cobra.Command, args []string) {
	answers := &model.Answers{}
	err := survey.Ask(basicqs, answers)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := deployer.ValidatevSphereAccount(answers); err != nil {
		fmt.Println(err.Error())
		return
	}

	if answers.IsVCenter {
		survey.AskOne(&survey.Input{
			Message: descriptions["datacenter"],
		}, &answers.Datacenter, nil)
	} else {
		answers.Datacenter = "ha-datacenter"
	}

	fmt.Println("Searching...")
	vmconfig, err := deployer.FindMasterNode(answers)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if vmconfig == nil {
		fmt.Println("Cannot find a master node")
		return
	}

	fmt.Printf("Found master node at %s\n", vmconfig.IP)

	answers, _, err = RecoverConfigFilesFromMaster(answers, vmconfig)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Configuration files recovered")

	fmt.Printf("Cache %s kits...\n", answers.KubernetesVersion)
	cacher.CacheAll(answers.KubernetesVersion)

	fmt.Println("All recovered")

	runInfo(cmd, args)

}

func RecoverConfigFilesFromMaster(answers *model.Answers, vmconfig *model.K8sNode) (*model.Answers, *model.K8sNodes, error) {
	if err := deployer.CopyRemoteFileToLocal(vmconfig, constants.GetRemoteK8sNodesConfigFilePath(), constants.GetK8sNodesConfigFilePath()); err != nil {
		fmt.Println("Failed to download meta data")
		return nil, nil, err
	}

	vms, err := utils.ReadK8sNodes()
	if err != nil {
		return nil, nil, err
	}

	if err := deployer.CopyRemoteFileToLocal(vmconfig, constants.GetRemoteKubeVConfigFilePath(), constants.GetKubeVConfigFilePath()); err != nil {
		fmt.Println("Failed to download meta data")
		return nil, nil, err
	}
	downloadanswers, err := readConfig()
	if err != nil {
		return nil, nil, err
	}
	downloadanswers.Username = answers.Username
	downloadanswers.Password = answers.Password
	SaveAnswers(downloadanswers)

	if err := deployer.CopyRemoteFileToLocal(vmconfig, constants.GetRemoteVMPrivateKeyPath(), constants.GetVMPrivateKeyPath()); err != nil {
		fmt.Println("Failed to download meta data")
		return nil, nil, err
	}

	if err := deployer.CopyRemoteFileToLocal(vmconfig, constants.GetRemoteVMPublicKeyPath(), constants.GetVMPublicKeyPath()); err != nil {
		fmt.Println("Failed to download meta data")
		return nil, nil, err
	}

	if err := deployer.DownloadKubeCtlConfig(vmconfig); err != nil {
		fmt.Println("Failed to download meta data")
		return nil, nil, err
	}

	return downloadanswers, vms, nil
}
