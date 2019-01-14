package deployer

import (
	"fmt"

	"github.com/jeffwubj/kubev/pkg/kubev/model"
)

func DeployNodes(answers *model.Answers) (*model.K8sNodes, error) {

	if err := generateSSHKey(); err != nil {
		return nil, err
	}

	o, err := DeployOVA(answers)
	if err != nil {
		return nil, err
	}
	fmt.Println(o.Name() + " deployed")

	var workderNodes []*model.K8sNode

	for i := 1; i <= answers.WorkerNodes; i++ {
		workderNodes = append(workderNodes, &model.K8sNode{
			MasterNode: false,
			VMName:     fmt.Sprintf("kubev-worker-%d", i),
		})
	}

	k8sNodes := &model.K8sNodes{
		MasterNode: &model.K8sNode{
			MasterNode: true,
			VMName:     "kubev-master",
		},
		WorkerNodes: workderNodes,
	}

	_, err = CloneVM(k8sNodes.MasterNode, answers)
	if err != nil {
		return nil, err
	}

	fmt.Printf("%s deployed\n", k8sNodes.MasterNode.VMName)
	modify_known_hosts(k8sNodes.MasterNode.IP)
	if err := ConfigVM(k8sNodes.MasterNode); err != nil {
		return nil, err
	}

	for _, vm := range k8sNodes.WorkerNodes {
		_, err := CloneVM(vm, answers)
		if err != nil {
			return nil, err
		}
		fmt.Printf("%s deployed\n", vm.VMName)
		modify_known_hosts(vm.IP)
		if err := ConfigVM(vm); err != nil {
			return nil, err
		}
	}

	return k8sNodes, nil
}
