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

	"github.com/jeffwubj/kubev/pkg/kubev/deployer"
	"github.com/jeffwubj/kubev/pkg/kubev/model"
	"github.com/jeffwubj/kubev/pkg/kubev/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	survey "gopkg.in/AlecAivazis/survey.v1"
)

// scaleCmd represents the scale command
var scaleCmd = &cobra.Command{
	Use:   "scale",
	Short: "Add more workder nodes or remove existing workder nodes",
	Long:  ``,
	Run:   runScale,
}

func init() {
	rootCmd.AddCommand(scaleCmd)
}

func runScale(cmd *cobra.Command, args []string) {
	if !utils.FileExists(viper.ConfigFileUsed()) {
		fmt.Println("There is no config file, run config and deploy before scale")
		return
	}

	answers, err := readConfig()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	vms, err := utils.ReadK8sNodes()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	workernodes := vms.WorkerNodes
askagain:
	number := len(workernodes)
	survey.AskOne(&survey.Input{
		Message: fmt.Sprintf("There are %d worker nodes in current cluster, how many workder nodes do you want?", number),
	}, &number, nil)

	if number <= 0 || number > 1000 {
		fmt.Println("Please input a validate worker nodes number")
		goto askagain
	}

	if answers.WorkerNodes == number {
		fmt.Printf("Cluster already has %d worker nodes, no extra action needed", number)
	} else if answers.WorkerNodes > number { // DELETE
		fmt.Printf("Changing cluster with %d nodes...\n", number)
		todelete := answers.WorkerNodes - number
		for i := 0; i < todelete; i++ {
			x := workernodes[len(workernodes)-1]
			workernodes = workernodes[:len(workernodes)-1]
			if err := deployer.DestorySingle(answers, x); err != nil {
				fmt.Printf("Failed to delete %s: %s\n", x.VMName, err.Error())
				break
			}
			if err := deployer.DeleteWorkerNodeFromKubenretes(x, vms); err != nil {
				fmt.Printf("Failed to remove %s: %s\n", x.VMName, err.Error())
				break
			}
			answers.WorkerNodes = answers.WorkerNodes - 1
		}
	} else { // ADD
		joincmd, err := deployer.GetKubeAdmJoinCommand(vms.MasterNode)
		if err != nil {
			fmt.Printf("Failed to join master node: %s\n", err.Error())
			return
		}
		vms.JoinString = joincmd
		fmt.Printf("Changing cluster with %d nodes...\n", number)
		toadd := number - answers.WorkerNodes
		worker := "kubev-esx-worker"
		if answers.IsVCenter {
			worker = "kubev-vc-worker"
		}
		for i := 0; i < toadd; i++ {
			newnode := &model.K8sNode{
				MasterNode: false,
				VMName:     fmt.Sprintf("%s-%d", worker, i+answers.WorkerNodes+1),
				Ready:      false,
			}
			if err := deployer.DeployWorkderNode(newnode, answers, vms); err != nil {
				fmt.Printf("Failed to add new worker node %s: %s\n", newnode.VMName, err.Error())
				break
			} else {
				workernodes = append(workernodes, newnode)
				answers.WorkerNodes = answers.WorkerNodes + 1
			}
		}

	}

	vms.WorkerNodes = workernodes
	utils.SaveK8sNodes(vms)
	SaveAnswers(answers)
}
