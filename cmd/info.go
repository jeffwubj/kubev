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

	"github.com/jeffwubj/kubev/pkg/kubev/utils"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Print kubev cluster informations",
	Long:  ``,
	Run:   runInfo,
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

func runInfo(cmd *cobra.Command, args []string) {
	if !utils.FileExists(viper.ConfigFileUsed()) {
		fmt.Println("There is no config file, deploy a cluster or run 'kubev recover' to find an existing cluster")
		return
	}

	answers, err := readConfig()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Kubernetes version is", answers.KubernetesVersion)
	fmt.Println("Host is", answers.Serverurl)

	vms, err := utils.ReadK8sNodes()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	if vms == nil || vms.MasterNode == nil {
		fmt.Println("There is no config file, deploy a cluster or run 'kubev recover' to find an existing cluster")
		return
	}

	data := [][]string{}
	data = append(data, []string{vms.MasterNode.VMName, "master", vms.MasterNode.IP})
	for _, vm := range vms.WorkerNodes {
		data = append(data, []string{vm.VMName, "worker", vm.IP})
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAME", "ROLES", "IP"})
	table.SetBorder(true)
	table.AppendBulk(data)
	table.Render()
}
