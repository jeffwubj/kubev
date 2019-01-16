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

package utils

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

func EncodeToken(vmconfig *model.K8sNode) string {
	token := vmconfig.IP
	data := []byte(token)
	return base64.StdEncoding.EncodeToString(data)
}

func DecodeToken(token string) (string, error) {
	decodeBytes, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return "", fmt.Errorf("Invalide token")
	}
	return string(decodeBytes), nil
}

func Is_ipv4(host string) bool {
	parts := strings.Split(host, ".")

	if len(parts) < 4 {
		return false
	}

	for _, x := range parts {
		if i, err := strconv.Atoi(x); err == nil {
			if i < 0 || i > 255 {
				return false
			}
		} else {
			return false
		}

	}
	return true
}
