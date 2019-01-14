// Deploy kubectl in host
// Clone VM in host

package deployer

import (
	"io/ioutil"
	"path"
	"regexp"

	"github.com/docker/machine/libmachine/ssh"
	"github.com/jeffwubj/kubev/pkg/kubev/constants"
	"github.com/jeffwubj/kubev/pkg/kubev/utils"
)

func modify_known_hosts(ip string) error {
	filepath := path.Join(constants.GetHomeFolder(), ".ssh", "known_hosts")
	if !utils.FileExists(filepath) {
		return nil
	}

	read, err := ioutil.ReadFile(filepath)
	if err != nil {
		panic(err)
	}

	re := regexp.MustCompile("(?m)[\r\n]+^.*" + ip + ".*$")
	newContents := re.ReplaceAllString(string(read), "")
	regex, err := regexp.Compile("\n\n")
	if err != nil {
		panic(err)
	}
	newContents = regex.ReplaceAllString(newContents, "\n")
	err = ioutil.WriteFile(filepath, []byte(newContents), 0)
	if err != nil {
		panic(err)
	}
	return nil
}

func generateSSHKey() error {
	if !utils.FileExists(constants.GetVMPrivateKeyPath()) ||
		!utils.FileExists(constants.GetVMPublicKeyPath()) {
		if err := ssh.GenerateSSHKey(constants.GetVMPrivateKeyPath()); err != nil {
			return err
		}
	}
	return nil
}
