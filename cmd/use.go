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
	"os"
	"path/filepath"

	"github.com/jeffwubj/kubev/pkg/kubev/cacher"
	"github.com/jeffwubj/kubev/pkg/kubev/constants"
	"github.com/jeffwubj/kubev/pkg/kubev/deployer"
	"github.com/jeffwubj/kubev/pkg/kubev/model"
	"github.com/jeffwubj/kubev/pkg/kubev/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// useCmd represents the use command
var useCmd = &cobra.Command{
	Use:   "use",
	Short: "Use a ticket in this command to play to shared Kuberntes cluster",
	Long: `Ticket will be printed in command kubev deploy and shared to other host 
that will use this cluster`,
	Run: runUse,
}

func init() {
	rootCmd.AddCommand(useCmd)
	useCmd.Flags().String("token", "", "token printed in the 'kubev deploy' command")
	useCmd.MarkFlagRequired("token")
	viper.BindPFlags(useCmd.Flags())
}

func runUse(cmd *cobra.Command, args []string) {
	token := viper.GetString("token")
	ip, err := utils.DecodeToken(token)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if !utils.Is_ipv4(ip) {
		fmt.Println("Incorrect token")
		return
	}

	vmconfig := &model.K8sNode{
		IP: ip,
	}

	if err := deployer.DownloadKubeCtlConfig(vmconfig); err != nil {
		fmt.Println("Failed to download meta data")
		return
	}
	fmt.Println("kubectl config file deployed")

	if err := deployer.CopyRemoteFileToLocal(vmconfig, constants.GetRemoteKubeVConfigFilePath(), constants.GetKubeVConfigFilePath()); err != nil {
		fmt.Println("Failed to download meta data")
		return
	}
	answers, err := readConfig()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	k8sversion := answers.KubernetesVersion
	fmt.Printf("Kubernetes version is %s\n", k8sversion)
	_, err = cacher.Cache(false, constants.KubeCtlBinaryName, k8sversion)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	symlink := filepath.Join("/usr/local/bin/", constants.KubeCtlBinaryName)
	os.Remove(symlink)
	kubectlLocalPath := constants.GetLocalK8sKitFilePath(constants.KubeCtlBinaryName, k8sversion)
	if err := os.Symlink(kubectlLocalPath, symlink); err != nil {
		fmt.Println(err.Error())
		fmt.Printf("Failed to link kubectl, please put %s into your path.\n", kubectlLocalPath)
		return
	}
	fmt.Println("kubectl is ready, enjoy your Kubernetes cluster")
}
