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
	"github.com/jeffwubj/kubev/pkg/kubev/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	survey "gopkg.in/AlecAivazis/survey.v1"
)

// destoryCmd represents the destory command
var destoryCmd = &cobra.Command{
	Use:   "destory",
	Short: "Attention! This command will try to destory VMs in your cluster!",
	Long:  ``,
	Run:   runDestory,
}

func init() {
	rootCmd.AddCommand(destoryCmd)

}

func runDestory(cmd *cobra.Command, args []string) {
	if !utils.FileExists(viper.ConfigFileUsed()) {
		fmt.Println("There is no config file, run config before deploy")
		return
	}

	vms, err := utils.ReadK8sNodes()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if vms == nil || vms.MasterNode == nil || vms.WorkerNodes == nil {
		fmt.Println("Nothing to destory")
		return
	}

	if vms.MasterNode != nil && vms.MasterNode.Ready {
		force := false
		survey.AskOne(&survey.Confirm{
			Message: "There is a running cluster, do you want to destory it?",
			Default: false,
		}, &force, nil)
		if !force {
			fmt.Println("Bye")
			return
		}
	}

	answer := false
	survey.AskOne(&survey.Confirm{
		Message: "Are you willing to destory cluster, it will delete all VMs in this cluster?",
		Default: false,
	}, &answer, nil)
	if !answer {
		fmt.Println("Bye")
		return
	}

	answers, err := readConfig()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	err = deployer.Destory(answers, vms)
	if err != nil {
		fmt.Println("Destory nodes failed...")
		fmt.Println(err.Error())
		return
	}
	utils.DeleteFile(constants.GetK8sNodesConfigFilePath())
	utils.DeleteFile(viper.ConfigFileUsed())
}
