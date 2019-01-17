# Getting Started on vSphere with Kuberntes

## Why kubev
I want to have a Kubernetes cluster in my ESX, previously I am using [Kubernetes Anywhere](https://github.com/kubernetes/kubernetes-anywhere), but obviously it has been deprecated.
Then I tried quiet some other options, but unluckily none of them are easy to use(especially I am behind GFW) and I had to setup a cluster manually.
This project aims to automate that work together with bunch of facility tools.

So **kubev** is a CLI tool to provision a Kubernetes cluster in vCenter or ESX and modify host to be able to manage that cluster.

## How will it work
The whole idea is to download virtual machine template then upload it to vCenter/ESX, and cloned the template to different Kubernetes nodes and configure them to be Kubernetes ready.
To make it easy to use, kubev will automatcially download **EVERYTHING** including:
* vm ova
* kubectl
* kubelet
* kubeadm
* ...

Ideally, just one single kubev binary is enough to deploy an experimental Kubernetes cluster in vSphere or ESX.
### Prerequisites
  * kubev is for Mac **ONLY** now, willing to migrate it to linux or windows if someone really need it.
  * Deployment requires DHCP server in the VM network.
  * vCenter user with following minimal set of privileges.
      ```
      Datastore > Allocate space
      Datastore > Low level file Operations
      Folder > Create Folder
      Folder > Delete Folder
      Network > Assign network
      Resource > Assign virtual machine to resource pool
      Virtual machine > Configuration > Add new disk
      Virtual Machine > Configuration > Add existing disk
      Virtual Machine > Configuration > Add or remove device
      Virtual Machine > Configuration > Change CPU count
      Virtual Machine > Configuration > Change resource
      Virtual Machine > Configuration > Memory
      Virtual Machine > Configuration > Modify device settings
      Virtual Machine > Configuration > Remove disk
      Virtual Machine > Configuration > Rename
      Virtual Machine > Configuration > Settings
      Virtual machine > Configuration > Advanced
      Virtual Machine > Interaction > Power off
      Virtual Machine > Interaction > Power on
      Virtual Machine > Inventory > Create from existing
      Virtual Machine > Inventory > Create new
      Virtual Machine > Inventory > Remove
      Virtual Machine > Provisioning > Clone virtual machine
      Virtual Machine > Provisioning > Customize
      Virtual Machine > Provisioning > Read customization specifications
      vApp > Import
      Profile-driven storage -> Profile-driven storage view
      ```
 
 ### Phases
 The whole kubev procedure includes below phases by different commands:
 * Config
 * Deploy
 * Use
 * Recover
 * Scale
 * Destory
 
 ### Install kubev
 ```
 go get github.com/jeffwubj/kubev
 ```
 
 ### Config
 `kubev config`

 This should be the first command to start, you will need to input vCenter/ESX information together Kubernetes information. You will need to prepare Datacenter, cluster before running it.
 
 ### Deploy
 `kubev deploy`
 
 After run `kubev config`, we should run this command to deploy a Kubernetes cluster in vCenter or ESX.
 Underneath, it will download kubectl, kubelet, kubeadm, vitual machine templates and deploy them to vCenter or ESX.
 
 After deploy succeed, it will print a `kubev use --token xxx` command, this command can be run in another host, kubev will then automatically download kubectl and config files to manage this cluster.
 
 ### Use
 `kubev use --token xxx`
 
 In another host, we can run this command to prepare host to manage Kubernetes cluster provisioned previously.
 
 ### Recover
 `kubev recover`
 
 This command will searching vCenter or ESX for cluster deployed by kubev and make host being able to manage this cluster.
 
 ### Scale
 `kubev scale`
 
 This command can add more nodes or remove existing nodes from managed cluster.
 
 ### Destory
 `kubev destory`
 
 This will destory all nodes deployed by kubev, **be carful on this**
 
 ### Notes
 kubev will deploy several virtual machines in vCenter or ESX, they will have name kubev-xxx-xxx, do not modify them manually otherwise the cluster may not work well.
 
 
 
