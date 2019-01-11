// Upload ova to vSphere
// Clone VM in vSphere
// Prepare resource pool
// Prepare folder
// Validate resources

package deployer

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/davecgh/go-spew/spew"
	"github.com/jeffwubj/kubev/pkg/kubev/constants"
	"github.com/jeffwubj/kubev/pkg/kubev/driver"
	"github.com/jeffwubj/kubev/pkg/kubev/model"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/govc/importx"
	"github.com/vmware/govmomi/ovf"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
)

func DeployOVA(answers model.Answers) (*types.ManagedObjectReference, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := driver.NewClient(ctx, answers)
	if err != nil {
		return nil, err
	}

	finder := find.NewFinder(client.Client, true)
	datacenter, err := finder.Datacenter(ctx, answers.Datacenter)
	if err != nil {
		return nil, err
	}

	finder.SetDatacenter(datacenter)
	datastore, err := finder.Datastore(ctx, answers.Datastore)
	if err != nil {
		return nil, err
	}

	resourcepool, err := finder.ResourcePool(ctx, answers.Resourcepool)
	if err != nil {
		return nil, err
	}

	folders, err := datacenter.Folders(context.TODO())
	if err != nil {
		return nil, err
	}

	folder := folders.VmFolder

	fpath := constants.GetLocalK8sKitFilePath(constants.PhotonOVAName, constants.DefaultPhotonVersion)

	fmt.Printf("Deploy %s to %s/%s/%s under folder %s and datastore %s ...\n",
		fpath, answers.Serverurl, datacenter.Name(), resourcepool.Name(), folder.Name(), datastore.Name())

	archive := &importx.TapeArchive{}
	archive.SetPath(fpath)
	archive.Client = client.Client
	r, _, err := archive.Open("*.ovf")
	if err != nil {
		return nil, err
	}
	defer r.Close()
	o, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// e, err := ovf.Unmarshal(bytes.NewReader(o))
	// if err != nil {
	// 	return err
	// }
	// spew.Dump(e.Network)

	var networks []types.OvfNetworkMapping

	network, err := finder.Network(ctx, answers.Network)
	if err != nil {
		return nil, err
	}

	networks = append(networks, types.OvfNetworkMapping{
		Name:    answers.Network,
		Network: network.Reference(),
	})

	cisp := types.OvfCreateImportSpecParams{
		DiskProvisioning:   "",
		EntityName:         "kube-template",
		IpAllocationPolicy: "",
		IpProtocol:         "",
		OvfManagerCommonParams: types.OvfManagerCommonParams{
			DeploymentOption: "",
			Locale:           "US"},
		PropertyMapping: nil,
		NetworkMapping:  networks,
	}

	m := ovf.NewManager(client.Client)
	spec, err := m.CreateImportSpec(ctx, string(o), resourcepool, datastore, cisp)
	if err != nil {
		return nil, err
	}

	lease, err := resourcepool.ImportVApp(ctx, spec.ImportSpec, folder, nil)
	if err != nil {
		return nil, err
	}
	info, err := lease.Wait(ctx, spec.FileItem)
	if err != nil {
		return nil, err
	}
	u := lease.StartUpdater(ctx, info)
	defer u.Done()
	for _, i := range info.Items {
		f, size, err := archive.Open(i.Path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		opts := soap.Upload{
			ContentLength: size,
		}
		lease.Upload(ctx, i, f, opts)
	}

	spew.Dump(info.Entity)
	lease.Complete(ctx)

	return &info.Entity, nil
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
