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

	answers, err := readConfig()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	cacher.CacheAll(viper.GetString("kubernetesversion"))

	if err := deployer.DeployMasterNode(answers); err != nil {
		fmt.Println("Deploy master failed...")
		fmt.Println(err.Error())
	}

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
		Cluster:           viper.GetString("cluster"),
		Resourcepool:      viper.GetString("resourcepool"),
		Folder:            viper.GetString("folder"),
		Cpu:               viper.GetInt("cpu"),
		Memory:            viper.GetInt("memory"),
		Network:           viper.GetString("network"),
		KubernetesVersion: viper.GetString("kubernetesversion"),
	}, nil
}
