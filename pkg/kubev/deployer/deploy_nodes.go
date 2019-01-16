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

package deployer

import (
	"fmt"

	"github.com/jeffwubj/kubev/pkg/kubev/constants"
	"github.com/jeffwubj/kubev/pkg/kubev/model"
	"github.com/jeffwubj/kubev/pkg/kubev/utils"
)

func DeployNodes(answers *model.Answers) (*model.K8sNodes, error) {
	if err := generateSSHKey(); err != nil {
		return nil, err
	}

	var workderNodes []*model.K8sNode

	worker := "kubev-esx-worker"
	if answers.IsVCenter {
		worker = "kubev-vc-worker"
	}

	for i := 1; i <= answers.WorkerNodes; i++ {
		workderNodes = append(workderNodes, &model.K8sNode{
			MasterNode: false,
			VMName:     fmt.Sprintf("%s-%d", worker, i),
			Ready:      false,
		})
	}

	masterName := "kubev-esx-master"
	if answers.IsVCenter {
		masterName = "kubev-vc-master"
	}

	k8sNodes := &model.K8sNodes{
		MasterNode: &model.K8sNode{
			MasterNode: true,
			VMName:     masterName,
			Ready:      false,
		},
		WorkerNodes: workderNodes,
	}

	_, err := CreateVM(k8sNodes.MasterNode, answers)
	if err != nil {
		return nil, err
	}

	fmt.Printf("%s cloned\n", k8sNodes.MasterNode.VMName)
	modify_known_hosts(k8sNodes.MasterNode.IP)
	if err := ConfigVM(k8sNodes.MasterNode); err != nil {
		return nil, err
	}

	if !k8sNodes.MasterNode.Ready {
		err = UpdateMasterNode(k8sNodes)
		if err != nil {
			return nil, err
		}
	}

	// TODO Do it in parallel
	for _, vm := range k8sNodes.WorkerNodes {
		_, err := CreateVM(vm, answers)
		if err != nil {
			return nil, err
		}
		fmt.Printf("%s cloned\n", vm.VMName)
		modify_known_hosts(vm.IP)
		if err := ConfigVM(vm); err != nil {
			return nil, err
		}
		if err := UpdateWorkerNode(vm, k8sNodes); err != nil {
			return nil, err
		}

	}

	k8sNodes.MasterNode.Ready = true

	for _, vm := range k8sNodes.WorkerNodes {
		vm.Ready = true
	}

	utils.SaveK8sNodes(k8sNodes)

	// RISK, WA for now
	// TODO remove password before upload and re-add password after recovery, password will can be read from user input

	if err := CopyLocalFileToRemote(k8sNodes.MasterNode, constants.GetK8sNodesConfigFilePath(), constants.GetRemoteK8sNodesConfigFilePath()); err != nil {
		fmt.Println("Failed to upload meta data, but cluster has been deployed successfully")
		return k8sNodes, err
	}

	if err := CopyLocalFileToRemote(k8sNodes.MasterNode, constants.GetKubeVConfigFilePath(), constants.GetRemoteKubeVConfigFilePath()); err != nil {
		fmt.Println("Failed to upload meta data, but cluster has been deployed successfully")
		return k8sNodes, err
	}

	if err := CopyLocalFileToRemote(k8sNodes.MasterNode, constants.GetVMPrivateKeyPath(), constants.GetRemoteVMPrivateKeyPath()); err != nil {
		fmt.Println("Failed to upload meta data, but cluster has been deployed successfully")
		return k8sNodes, err
	}

	if err := CopyLocalFileToRemote(k8sNodes.MasterNode, constants.GetVMPublicKeyPath(), constants.GetRemoteVMPublicKeyPath()); err != nil {
		fmt.Println("Failed to upload meta data, but cluster has been deployed successfully")
		return k8sNodes, err
	}

	return k8sNodes, nil
}
