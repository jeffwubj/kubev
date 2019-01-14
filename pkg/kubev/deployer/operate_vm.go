// Configure VM
// Upload kits

package deployer

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/ThomasRooney/gexpect"
	"github.com/jeffwubj/kubev/pkg/kubev/constants"
	"github.com/jeffwubj/kubev/pkg/kubev/model"
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
