// Deploy a master node

package deployer

import (
	"fmt"

	"github.com/jeffwubj/kubev/pkg/kubev/model"
)

func DeployMasterNode(answers *model.Answers) error {
	o, err := DeployOVA(answers)
	if err != nil {
		return err
	}
	fmt.Println(o.Name() + " deployed")

	v, err := CloneVM(answers)
	if err != nil {
		return err
	}

	fmt.Println(v.Name() + " cloned")
	return nil
}
