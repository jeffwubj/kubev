// Upload ova to vSphere
// Clone VM in vSphere
// Prepare resource pool
// Prepare folder
// Validate resources

package deployer

import (
	"context"
	"jeffwubj/kubev/pkg/kubev/driver"
	"jeffwubj/kubev/pkg/kubev/model"

	"github.com/davecgh/go-spew/spew"
)

func DeployOVA(answers model.Answers) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := driver.NewClient(ctx, answers)
	if err != nil {
		return err
	}
	spew.Dump(client.IsVC())
	return nil
}

func PrepareResourcePool(answers model.Answers) error {

	return nil
}

func PrepareFolder(answers model.Answers) error {

	return nil
}

func ValidatevSphereAccount(answers model.Answers) error {

	return nil
}

func ValidatevSphere(answers model.Answers) error {

	return nil
}
