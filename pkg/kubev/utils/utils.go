package utils

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"vmware/f8s/pkg/utils"

	"github.com/jeffwubj/kubev/pkg/kubev/constants"
	"github.com/jeffwubj/kubev/pkg/kubev/model"
	"github.com/phayes/permbits"
)

// BinaryExists checks whether binary exists with executable permission
func BinaryExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, errors.New(path + " does not exist")
	} else if err == nil {
		permissions, err := permbits.Stat(path)
		if err != nil {
			return false, err
		}
		if !permissions.GroupExecute() ||
			!permissions.UserExecute() ||
			!permissions.OtherExecute() {
			return false, errors.New(path + " has no execute permission")
		}
		return true, nil
	}

	return false, err
}

func FileExists(path string) bool {
	_, err := os.Stat(path)

	if os.IsNotExist(err) {
		return false
	} else if err == nil {

		// if info, err := os.Stat(path); err != nil {
		// 	return false, err
		// }
		// else if info.Mode()&0111 != 0 {
		// 	return false, errors.New(path + " cannot be executed")
		// }
		return true
	}

	return false
}

func DeleteFile(path string) bool {
	if err := os.Remove(path); err != nil {
		return false
	}
	return true
}

func MakeBinaryExecutable(targetFilepath string) error {
	permissions, err := permbits.Stat(targetFilepath)
	if err != nil {
		return err
	}
	permissions.SetUserExecute(true)
	permissions.SetGroupExecute(true)
	permissions.SetOtherExecute(true)
	err = permbits.Chmod(targetFilepath, permissions)
	if err != nil {
		return errors.New("error setting permission on file")
	}
	return nil
}

func ReadK8sNodes() (*model.K8sNodes, error) {
	var cc model.K8sNodes

	path := constants.GetK8sNodesConfigFilePath()

	if !utils.FileExists(path) {
		return &cc, nil
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return &cc, err
	}

	if err := json.Unmarshal(data, &cc); err != nil {
		return &cc, err
	}
	return &cc, nil
}

func SaveK8sNodes(config *model.K8sNodes) error {
	data, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return err
	}

	profileConfigFile := constants.GetK8sNodesConfigFilePath()

	if err := os.MkdirAll(filepath.Dir(profileConfigFile), 0700); err != nil {
		return err
	}

	if err := saveConfigToFile(data, profileConfigFile); err != nil {
		return err
	}

	return nil
}

func saveConfigToFile(data []byte, file string) error {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return ioutil.WriteFile(file, data, 0600)
	}

	tmpfi, err := ioutil.TempFile(filepath.Dir(file), "config.json.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmpfi.Name())

	if err = ioutil.WriteFile(tmpfi.Name(), data, 0600); err != nil {
		return err
	}

	if err = tmpfi.Close(); err != nil {
		return err
	}

	if err = os.Remove(file); err != nil {
		return err
	}

	if err = os.Rename(tmpfi.Name(), file); err != nil {
		return err
	}
	return nil
}
