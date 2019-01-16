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
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/jeffwubj/kubev/pkg/kubev/cacher"
	"github.com/jeffwubj/kubev/pkg/kubev/deployer"
	"github.com/jeffwubj/kubev/pkg/kubev/model"
	"github.com/jeffwubj/kubev/pkg/kubev/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/AlecAivazis/survey.v1"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy Kubernetes based on config file",
	Long:  ``,
	Run:   runDeploy,
}

func init() {
	rootCmd.AddCommand(deployCmd)
}

func runDeploy(cmd *cobra.Command, args []string) {
	if !utils.FileExists(viper.ConfigFileUsed()) {
		fmt.Println("There is no config file, run config before deploy")
		return
	}

	vms, err := utils.ReadK8sNodes()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if vms.MasterNode != nil && vms.MasterNode.Ready {
		force := false
		survey.AskOne(&survey.Confirm{
			Message: "There is a running cluster config locally, do you want to overwrite local config?",
			Default: false,
		}, &force, nil)
		if !force {
			fmt.Println("Bye")
			return
		}
	}

	answers, err := readConfig()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	vmconfig, err := deployer.FindMasterNode(answers)

	if vmconfig != nil {
		overwrite := false
		survey.AskOne(&survey.Confirm{
			Message: fmt.Sprintf(`Found Kubernetes master node %s, do you want to overwrite this cluster? (You should use kubev scale to change existing cluster)`, vmconfig.VMName),
			Default: false,
		}, &overwrite, nil)
		if !overwrite {
			fmt.Println("Bye")
			return
		}
	}

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	cacher.CacheAll(viper.GetString("kubernetesversion"))

	vms, err = deployer.DeployNodes(answers)
	if err != nil {
		fmt.Println("Deploy nodes failed...")
		fmt.Println(err.Error())
		return
	}

	utils.SaveK8sNodes(vms)
}

func readConfig() (*model.Answers, error) {
	dat, err := ioutil.ReadFile(viper.ConfigFileUsed())
	if err != nil {
		return nil, err
	}

	viper.ReadConfig(bytes.NewBuffer(dat))

	return &model.Answers{
		Serverurl:         viper.GetString("serverurl"),
		Port:              viper.GetInt("port"),
		Username:          viper.GetString("username"),
		Password:          viper.GetString("password"),
		Datacenter:        viper.GetString("datacenter"),
		Datastore:         viper.GetString("datastore"),
		Resourcepool:      viper.GetString("resourcepool"),
		Folder:            viper.GetString("folder"),
		Cpu:               viper.GetInt("cpu"),
		Memory:            viper.GetInt("memory"),
		Network:           viper.GetString("network"),
		KubernetesVersion: viper.GetString("kubernetesversion"),
		WorkerNodes:       viper.GetInt("workernodes"),
		IsVCenter:         viper.GetBool("isvcenter"),
	}, nil
}
