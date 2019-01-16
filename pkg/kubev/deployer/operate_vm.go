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
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ThomasRooney/gexpect"
	"github.com/jeffwubj/kubev/pkg/kubev/constants"
	"github.com/jeffwubj/kubev/pkg/kubev/model"
	"github.com/pkg/sftp"
	cryptossh "golang.org/x/crypto/ssh"
)

func ConfigVM(vmconfig *model.K8sNode) error {
	fmt.Printf("config VM %s:%s\n", vmconfig.VMName, vmconfig.IP)
	if !needConfigPhoton(vmconfig.IP) {
		fmt.Printf("no need to change password of %s\n", vmconfig.IP)
		return nil
	}

	if err := changePhotonDefaultPassword(vmconfig.IP); err != nil {
		fmt.Println("change password failed")
		return err
	}

	fmt.Printf("change default password of %s:%s succeed\n", vmconfig.VMName, vmconfig.IP)

	if err := configSSHInVM(vmconfig); err != nil {
		fmt.Println("config ssh failed\n")
		return err
	}

	return nil
}

func CopyLocalFileToRemote(vmconfig *model.K8sNode, localpath, remotepath string) error {
	fmt.Printf("copy file from %s to %s\n", localpath, remotepath)
	_, c, err := GetSSHRunner(vmconfig)
	if err != nil {
		return err
	}
	client, err := sftp.NewClient(c)
	if err != nil {
		return err
	}
	defer client.Close()

	srcFile, err := os.Open(localpath)
	if err != nil {
		fmt.Println("os.Open error : ", localpath)
		log.Fatal(err)

	}
	defer srcFile.Close()

	client.Mkdir(filepath.Dir(remotepath))

	dstFile, err := client.Create(remotepath)
	if err != nil {
		fmt.Println("sftpClient.Create error : ", remotepath)
		log.Fatal(err)

	}
	defer dstFile.Close()

	ff, err := ioutil.ReadAll(srcFile)
	if err != nil {
		fmt.Println("ReadAll error : ", localpath)
		log.Fatal(err)

	}
	dstFile.Write(ff)
	return nil
}

func DownloadKubeCtlConfig(vmconfig *model.K8sNode) error {
	_, c, err := GetSSHRunner2(vmconfig)
	if err != nil {
		return err
	}

	if err := PopuldateKubeConfig(c); err != nil {
		fmt.Println("Failed to write Kuberntes config file")
		return err
	}
	return nil
}

func CopyRemoteFileToLocal(vmconfig *model.K8sNode, remotepath, localpath string) error {
	_, c, err := GetSSHRunner2(vmconfig)
	if err != nil {
		return err
	}
	client, err := sftp.NewClient(c)
	if err != nil {
		return err
	}
	defer client.Close()

	srcFile, err := client.Open(remotepath)
	if err != nil {
		return err
	}

	os.Mkdir(filepath.Dir(localpath), os.ModePerm)
	dstFile, err := os.OpenFile(localpath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	err = dstFile.Sync()
	if err != nil {
		return err
	}

	return nil
}

func needConfigPhoton(ipAddress string) bool {
	err := executeSSHCommand("echo", "root", "kubernetes", ipAddress)
	if err != nil {
		return true
	}
	return false
}

func changePhotonDefaultPassword(ipAddress string) error {

	if err := modify_known_hosts(ipAddress); err != nil {
		return err
	}

	// TODO below does not work correctly
	cmd := fmt.Sprintf("ssh -o PubkeyAuthentication=no %s@%s", constants.PhotonVMUsername, ipAddress)

	child, err := gexpect.Spawn(cmd)
	if err != nil {
		return err
	}
	defer child.Close()
	timeout := 5 * time.Second
	if err := child.ExpectTimeout("Are you sure you want to continue connecting (yes/no)?", timeout); err != nil {
		return err
	}
	child.SendLine("yes")

	if err := child.ExpectTimeout("assword:", timeout); err != nil {
		return err
	}
	child.SendLine("changeme")

	if err := child.ExpectTimeout("assword:", timeout); err != nil {
		return err
	}
	child.SendLine("changeme")

	if err := child.ExpectTimeout("assword:", timeout); err != nil {
		return err
	}
	child.SendLine("kubernetes")

	if err := child.ExpectTimeout("assword:", timeout); err != nil {
		return err
	}
	child.SendLine("kubernetes")

	if err := child.ExpectTimeout("#", timeout); err != nil {
		return err
	}
	child.SendLine("exit")

	child.Start()
	return nil
}

func configSSHInVM(vmconfig *model.K8sNode) error {
	// read generated public ssh key
	keyfh, err := os.Open(constants.GetVMPublicKeyPath())
	if err != nil {
		return err
	}
	defer keyfh.Close()
	keycontent, err := ioutil.ReadAll(keyfh)
	if err != nil {
		return err
	}
	command := fmt.Sprintf("echo '%s' > /%s/.ssh/authorized_keys", string(keycontent), constants.PhotonVMUsername)
	if err := executeSSHCommand(command, constants.PhotonVMUsername, constants.PhotonVMPassword, vmconfig.IP); err != nil {
		return err
	}

	return nil
}

func executeSSHCommand(command, username, password, ip string) error {
	config := &cryptossh.ClientConfig{
		User: username,
		Auth: []cryptossh.AuthMethod{
			cryptossh.Password(password),
		},
		HostKeyCallback: cryptossh.InsecureIgnoreHostKey(),
	}

	client, err := cryptossh.Dial("tcp", fmt.Sprintf("%s:%d", ip, 22), config)
	if err != nil {
		return err
	}

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b

	if err := session.Run(command); err != nil {
		return err
	}

	return nil
}
