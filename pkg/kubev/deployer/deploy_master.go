// Deploy a master node

package deployer

import (
	"fmt"

	"github.com/jeffwubj/kubev/pkg/kubev/model"
)

func DeployMasterNode(answers model.Answers) error {
	o, err := DeployOVA(answers)
	if err != nil {
		return err
	}
	fmt.Println(o.String() + " deployed")
	return nil
}
