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
	"io/ioutil"
	"path"

	"github.com/jeffwubj/kubev/pkg/kubev/constants"
	"github.com/jeffwubj/kubev/pkg/kubev/model"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	cryptossh "golang.org/x/crypto/ssh"
	"k8s.io/minikube/pkg/minikube/assets"
)

func PrepareVM(vmconfig *model.K8sNode) error {
	fmt.Printf("Prepare k8s node %s...\n", vmconfig.VMName)

	k8sversion := viper.GetString("kubernetesversion")

	files := []assets.CopyableFile{
		assets.NewMemoryAssetTarget([]byte(constants.KubeletService), constants.KubeletServiceFile, "0640"),
		assets.NewMemoryAssetTarget([]byte(constants.KubeletSystemd), constants.KubeletSystemdConfFile, "0640"),
		// assets.NewMemoryAssetTarget([]byte(constants.DockerService), constants.DockerServiceFile, "0640"),
	}

	for _, bin := range []string{constants.KubeAdmBinaryName, constants.KubeletBinaryName, constants.CriCtlBinaryName} {
		binfile, err := assets.NewFileAsset(constants.GetLocalK8sKitFilePath(bin, k8sversion), "/usr/bin", bin, "0750")
		if err != nil {
			return err
		}
		files = append(files, binfile)
	}

	// fmt.Println("Prepare k8s binary files...")
	binfile, err := assets.NewFileAsset(constants.GetLocalK8sKitFilePath(constants.GuestKubeCtlBinaryName, k8sversion), "/usr/bin", constants.KubeCtlBinaryName, "0750")
	if err != nil {
		return err
	}
	files = append(files, binfile)

	// fmt.Println("Prepare CNI binary files...")
	// List all bins one bye one to make sure they are there
	for _, bin := range []string{"bridge", "dhcp", "flannel", "host-device", "host-local", "ipvlan", "loopback", "macvlan", "portmap", "ptp", "sample", "tuning", "vlan"} {
		binPath := path.Join(constants.GetLocalK8sKitPath(constants.CNIKits, k8sversion), bin)
		binfile, err := assets.NewFileAsset(binPath, "/opt/cni/bin/", bin, "0750")
		if err != nil {
			fmt.Println("Failed to read CNI binaries")
			return err
		}
		files = append(files, binfile)
	}

	runner, _, err := GetSSHRunner(vmconfig)
	if err != nil {
		return err
	}

	fmt.Println("Connected to guest.")

	fmt.Println("Copy files to guest...")
	for _, f := range files {
		if err := runner.Copy(f); err != nil {
			fmt.Println("Failed to copy files")
			return err
		}
	}

	fmt.Println("Configure services in guest...")
	err = runner.Run(`
	iptables --policy INPUT ACCEPT &&
	iptables --policy OUTPUT ACCEPT &&
	iptables --policy FORWARD ACCEPT
	`)
	if err != nil {
		return err
	}

	err = runner.Run(fmt.Sprintf("hostname %s", vmconfig.VMName))
	if err != nil {
		return err
	}

	fmt.Println("Restart services...")
	err = runner.Run(`
	systemctl daemon-reload &&
	systemctl enable kubelet &&
	systemctl enable kubelet.service &&
	systemctl enable docker &&
	systemctl start docker
	`)
	if err != nil {
		return err
	}
	return nil
}

func UpdateWorkerNode(vmconfig *model.K8sNode, k8snodes *model.K8sNodes) error {
	runner, _, err := GetSSHRunner(vmconfig)
	if err != nil {
		return err
	}

	if err := PrepareVM(vmconfig); err != nil {
		return err
	}

	runner.Run("kubeadm reset -f")
	if err != nil {
		return err
	}

	fmt.Println("Join worker node...")
	err = runner.Run(k8snodes.JoinString)
	if err != nil {
		return err
	}
	fmt.Printf("Install woker node %s finished\n", vmconfig.VMName)

	return nil
}

func GetSSHRunner(vmconfig *model.K8sNode) (*SSHRunner, *ssh.Client, error) {
	key, err := ioutil.ReadFile(constants.GetVMPrivateKeyPath())
	if err != nil {
		fmt.Println("Failed to SSH private keys")
		return nil, nil, err
	}

	privateKey, err := cryptossh.ParsePrivateKey(key)
	if err != nil {
		fmt.Println("Failed to parse private keys")
		return nil, nil, err
	}

	config := &cryptossh.ClientConfig{
		User: constants.PhotonVMUsername,
		Auth: []cryptossh.AuthMethod{
			cryptossh.PublicKeys(privateKey),
		},
		HostKeyCallback: cryptossh.InsecureIgnoreHostKey(),
	}

	c, err := cryptossh.Dial("tcp", fmt.Sprintf("%s:%d", vmconfig.IP, 22), config)
	if err != nil {
		fmt.Println("Failed to diag VM")
		return nil, nil, err
	}
	runner := NewSSHRunner(c)
	return runner, c, nil
}

func GetSSHRunner2(vmconfig *model.K8sNode) (*SSHRunner, *ssh.Client, error) {
	config := &cryptossh.ClientConfig{
		User: constants.PhotonVMUsername,
		Auth: []cryptossh.AuthMethod{
			cryptossh.Password(constants.PhotonVMPassword),
		},
		HostKeyCallback: cryptossh.InsecureIgnoreHostKey(),
	}

	c, err := cryptossh.Dial("tcp", fmt.Sprintf("%s:%d", vmconfig.IP, 22), config)
	if err != nil {
		fmt.Println("Failed to diag VM")
		return nil, nil, err
	}
	runner := NewSSHRunner(c)
	return runner, c, nil
}
