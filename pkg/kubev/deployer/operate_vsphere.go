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
	"context"
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/jeffwubj/kubev/pkg/kubev/constants"
	"github.com/jeffwubj/kubev/pkg/kubev/driver"
	"github.com/jeffwubj/kubev/pkg/kubev/model"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/govc/importx"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/ovf"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
)

func DeployOVA(answers *model.Answers) (*object.VirtualMachine, error) {

	// TODO if template VM powered on, we need to delete it, otherwise all cloned VM will have same IP

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

	tempaltePath := getTemplateVMPath(answers)
	vm, err := finder.VirtualMachine(ctx, tempaltePath)
	if err == nil {
		return vm, nil
	}

	var resourcepool *object.ResourcePool
	if answers.IsVCenter {
		resourcepool, err = finder.ResourcePool(ctx, answers.Resourcepool)
		if err != nil {
			return nil, err
		}
	} else {
		host, err := finder.DefaultHostSystem(ctx)
		if err != nil {
			return nil, err
		}
		resourcepool, err = host.ResourcePool(ctx)
		if err != nil {
			return nil, err
		}
	}

	folder, err := finder.Folder(ctx, getVMFolder(answers))
	if err != nil {
		return nil, err
	}

	fpath := constants.GetLocalK8sKitFilePath(constants.PhotonOVAName, constants.DefaultPhotonVersion)

	fmt.Printf("Deploy %s to %s ...\n", fpath, getTemplateVMPath(answers))

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
		EntityName:         constants.DefaultVMTemplateName,
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

	lease.Complete(ctx)

	vm, err = finder.VirtualMachine(ctx, getTemplateVMPath(answers))
	if err != nil {
		return nil, err
	}

	vmConfigSpec := types.VirtualMachineConfigSpec{}
	vmConfigSpec.NumCPUs = int32(answers.Cpu)
	vmConfigSpec.MemoryMB = int64(answers.Memory)
	vm.Reconfigure(ctx, vmConfigSpec)

	return vm, nil
}

func CloneVM(vmConfig *model.K8sNode, answers *model.Answers) (*object.VirtualMachine, error) {
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

	var resourcepool *object.ResourcePool
	if answers.IsVCenter {
		resourcepool, err = finder.ResourcePool(ctx, answers.Resourcepool)
		if err != nil {
			return nil, err
		}
	} else {
		host, err := finder.DefaultHostSystem(ctx)
		if err != nil {
			return nil, err
		}
		resourcepool, err = host.ResourcePool(ctx)
		if err != nil {
			return nil, err
		}
	}

	clonedVM, err := finder.VirtualMachine(ctx, path.Join(getVMFolder(answers), vmConfig.VMName))
	if err != nil {
		tempaltePath := getTemplateVMPath(answers)
		vm, err := finder.VirtualMachine(ctx, tempaltePath)
		if err != nil {
			return nil, err
		}

		configSpecs := []types.BaseVirtualDeviceConfigSpec{}

		folder, err := finder.Folder(ctx, getVMFolder(answers))
		if err != nil {
			return nil, err
		}

		folderref := folder.Reference()
		resourcepoolref := resourcepool.Reference()
		datastoreref := datastore.Reference()

		relocateSpec := types.VirtualMachineRelocateSpec{
			DeviceChange: configSpecs,
			Folder:       &folderref,
			Pool:         &resourcepoolref,
			Datastore:    &datastoreref,
		}

		if !answers.IsVCenter {
			host, err := finder.DefaultHostSystem(ctx)
			if err != nil {
				return nil, err
			}
			hostref := host.Reference()
			relocateSpec.Host = &hostref
			relocateSpec.DiskMoveType = string(types.VirtualMachineRelocateDiskMoveOptionsMoveAllDiskBackingsAndAllowSharing)
		}

		cloneSpec := &types.VirtualMachineCloneSpec{
			PowerOn:  false,
			Template: false,
		}
		cloneSpec.Location = relocateSpec

		// Clone vm to another vm
		task, err := vm.Clone(ctx, folder, vmConfig.VMName, *cloneSpec)
		if err != nil {
			return nil, err
		}

		// TODO failed with 'The operation is not supported on the object.' on esx, license issue?
		_, err = task.WaitForResult(ctx, nil)
		if err != nil {
			return nil, err
		}

		clonedVM, err = finder.VirtualMachine(ctx, path.Join(getVMFolder(answers), vmConfig.VMName))
		if err != nil {
			return nil, err
		}
	}

	powerstate, err := clonedVM.PowerState(ctx)
	if err != nil {
		return nil, err
	}

	if powerstate == types.VirtualMachinePowerStatePoweredOn {
		task, err := clonedVM.PowerOff(ctx)
		if err != nil {
			return nil, err
		}
		_, err = task.WaitForResult(ctx, nil)
		if err != nil {
			return nil, err
		}
	}

	vmConfigSpec := types.VirtualMachineConfigSpec{}
	vmConfigSpec.NumCPUs = int32(answers.Cpu)
	vmConfigSpec.MemoryMB = int64(answers.Memory)
	task, err := clonedVM.Reconfigure(ctx, vmConfigSpec)
	if err != nil {
		return nil, err
	}
	_, err = task.WaitForResult(ctx, nil)
	if err != nil {
		return nil, err
	}
	task, err = clonedVM.PowerOn(ctx)
	if err != nil {
		return nil, err
	}
	_, err = task.WaitForResult(ctx, nil)
	if err != nil {
		return nil, err
	}

	ip, err := clonedVM.WaitForIP(ctx)
	if err != nil {
		return nil, err
	}

	vmConfig.DatacenterName = datacenter.Name()
	vmConfig.DatastoreName = datastore.Name()
	vmConfig.FolderPath = clonedVM.InventoryPath
	vmConfig.IP = ip
	vmConfig.Mo = clonedVM.Reference().String()

	return clonedVM, nil
}

func Destory(answers *model.Answers, k8snodes *model.K8sNodes) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := driver.NewClient(ctx, answers)
	if err != nil {
		return err
	}

	err = delete(ctx, client, answers, k8snodes.MasterNode)
	if err != nil {
		return err
	}

	for _, node := range k8snodes.WorkerNodes {
		err = delete(ctx, client, answers, node)
		if err != nil {
			return err
		}
	}

	return nil
}

func delete(ctx context.Context, client *govmomi.Client, answers *model.Answers, k8snode *model.K8sNode) error {
	mos := strings.Split(k8snode.Mo, ":")
	if len(mos) != 2 {
		return fmt.Errorf("incorrect configuration for section %s", k8snode.Mo)
	}

	moref := types.ManagedObjectReference{
		Type:  mos[0],
		Value: mos[1],
	}
	vm := object.NewVirtualMachine(client.Client, moref)

	task, err := vm.PowerOff(ctx)
	if err != nil {
		return err
	}
	err = task.Wait(ctx)
	if err != nil {
		return err
	}
	task, err = vm.Destroy(ctx)
	if err != nil {
		return err
	}
	err = task.Wait(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("%s has been destoried\n", k8snode.VMName)
	return nil
}

func PrepareResourcePool(answers model.Answers) error {

	return nil
}

func PrepareFolder(answers model.Answers) error {

	return nil
}

func ValidatevSphereAccount(answers *model.Answers) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := driver.NewClient(ctx, answers)
	if err != nil {
		return err
	}

	answers.IsVCenter = client.IsVC()

	return nil
}

func getTemplateVMPath(answers *model.Answers) string {
	return "/" + path.Join(answers.Datacenter, "vm", answers.Folder, constants.DefaultVMTemplateName)
}

func getVMFolder(answers *model.Answers) string {
	return "/" + path.Join(answers.Datacenter, "vm", answers.Folder)
}
