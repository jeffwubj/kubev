// Deploy a master node

package deployer

import (
	"github.com/jeffwubj/kubev/pkg/kubev/model"
)

func DeployMasterNode(answers model.Answers) error {
	if err := DeployOVA(answers); err != nil {
		return err
	}
	return nil
}
