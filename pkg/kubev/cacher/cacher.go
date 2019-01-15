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

package cacher

import (
	"fmt"
	"os"
	"path"

	download "github.com/jeffwubj/go-download"
	"github.com/jeffwubj/kubev/pkg/kubev/constants"
	"github.com/jeffwubj/kubev/pkg/kubev/utils"
	"github.com/mholt/archiver"
)

func CacheAll(kubernetesVersion string) error {
	for _, binName := range []string{
		constants.KubeCtlBinaryName,
		constants.KubeAdmBinaryName,
		constants.KubeletBinaryName,
		constants.GuestKubeCtlBinaryName,
		constants.CriCtlBinaryName,
		constants.CNIKits,
		constants.DockerBinaryName,
	} {
		if _, err := cache(false, binName, kubernetesVersion); err != nil {
			return err
		}
	}

	if _, err := cache(false, constants.PhotonOVAName, constants.DefaultPhotonVersion); err != nil {
		return err
	}
	return nil
}

func cache(force bool, kitName, kitVersion string) (string, error) {

	targetDir := constants.GetLocalK8sKitPath(kitName, kitVersion)
	targetFilepath := path.Join(targetDir, kitName)

	_, err := os.Stat(targetFilepath)
	// If it exists, do no verification and continue
	if err == nil && !force {
		if kitName == constants.KubeCtlBinaryName {
			if err := utils.MakeBinaryExecutable(targetFilepath); err != nil {
				return "", err
			}
		}
		return targetFilepath, nil
	}
	if !os.IsNotExist(err) {
		return "", err
	}

	if err = os.MkdirAll(targetDir, 0777); err != nil {
		return "", err
	}

	url := constants.GetK8sKitReleaseURL(kitName, kitVersion)
	options := download.FileOptions{
		Mkdirs: download.MkdirAll,
		Options: download.Options{
			ProgressBars: &download.ProgressBarOptions{
				MaxWidth: 80,
			},
		},
	}

	// fmt.Println(url)
	// fmt.Println(targetFilepath)
	fmt.Printf("Downloading %s %s\n", kitName, kitVersion)

	if kitName == constants.CriCtlBinaryName || kitName == constants.DockerBinaryName {
		tarTargetFilepath := targetFilepath + ".tar.gz"
		if err := download.ToFile(url, tarTargetFilepath, options); err != nil {
			fmt.Println(err.Error())
			return "", err
		}
		fmt.Printf("Finished Downloading %s %s\n", kitName, kitVersion)

		if err := archiver.Unarchive(tarTargetFilepath, constants.GetLocalK8sKitPath(kitName, kitVersion)); err != nil {
			return "", err
		}

		return targetFilepath, nil
	} else if kitName == constants.CNIKits {
		if err := download.ToFile(url, targetFilepath, options); err != nil {
			fmt.Println(err.Error())
			return "", err
		}
		fmt.Printf("Finished Downloading %s %s\n", kitName, kitVersion)
		if err := archiver.Unarchive(targetFilepath, constants.GetLocalK8sKitPath(kitName, kitVersion)); err != nil {
			return "", err
		}
		return targetFilepath, nil
	}
	if err := download.ToFile(url, targetFilepath, options); err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	if kitName == constants.KubeCtlBinaryName || kitName == constants.DockerBinaryName {
		if err := utils.MakeBinaryExecutable(targetFilepath); err != nil {
			return "", err
		}
	}

	fmt.Printf("Finished Downloading %s %s\n", kitName, kitVersion)

	return targetFilepath, nil

}
