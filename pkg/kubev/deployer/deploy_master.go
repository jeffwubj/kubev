// Deploy a master node

package deployer

import (
	"jeffwubj/kubev/pkg/kubev/model"
)

func DeployMasterNode(answers model.Answers) error {
	if err := DeployOVA(answers); err != nil {
		return err
	}
	return nil
}
