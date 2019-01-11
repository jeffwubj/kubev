package utils

import (
	"errors"
	"os"

	"github.com/phayes/permbits"
	"github.com/spf13/viper"
)

// SaveConfig saves Viper configurations into config file that is in use
func SaveConfig() {
	viper.WriteConfigAs(viper.ConfigFileUsed())
}

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
