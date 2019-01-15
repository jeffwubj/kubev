package deployer

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/jeffwubj/kubev/pkg/kubev/constants"
	"github.com/jeffwubj/kubev/pkg/kubev/model"
	"github.com/jeffwubj/kubev/pkg/kubev/utils"
	"github.com/pkg/sftp"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	"k8s.io/minikube/pkg/util/kubeconfig"
)

func UpdateMasterNode(k8snodes *model.K8sNodes) error {
	fmt.Println("Prepare k8s master node...")

	vmconfig := k8snodes.MasterNode

	if err := PrepareVM(vmconfig); err != nil {
		return err
	}

	k8sversion := viper.GetString("kubernetesversion")

	runner, c, err := GetSSHRunner(vmconfig)
	if err != nil {
		return err
	}

	fmt.Println("Install Kubernetes...")
	output, err := runner.CombinedOutput(constants.KubeAdmInit)
	if err != nil {
		return err
	}
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		rawline := strings.TrimSpace(line)
		if strings.Contains(rawline, "kubeadm join") {
			k8snodes.JoinString = rawline
		}
	}

	if err := PopuldateKubeConfig(c); err != nil {
		fmt.Println("Failed to write Kuberntes config file")
		return err
	}

	symlink := filepath.Join("/usr/local/bin/", constants.KubeCtlBinaryName)
	os.Remove(symlink)
	kubectlLocalPath := constants.GetLocalK8sKitFilePath(constants.KubeCtlBinaryName, k8sversion)
	if err := os.Symlink(kubectlLocalPath, symlink); err != nil {
		fmt.Println(err.Error())
		fmt.Printf("Failed to link kubectl, please put %s into your path.\n", kubectlLocalPath)
	}

	symlink = filepath.Join("/usr/local/bin/", constants.DockerBinaryName)
	os.Remove(symlink)
	dockerLocalPath := constants.GetLocalK8sKitFilePath(constants.DockerBinaryName, k8sversion)

	dockerAliasContent := fmt.Sprintf(constants.DockerAlias, dockerLocalPath, vmconfig.IP+":2375")

	content := []byte(dockerAliasContent)
	ioutil.WriteFile(symlink, content, os.ModePerm)
	err = ioutil.WriteFile(symlink, content, 0644)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Printf("Failed to link docker, please put %s into your path.\n", dockerLocalPath)
	}

	fmt.Println("deploy master is done")

	return nil
}

func PopuldateKubeConfig(c *ssh.Client) error {
	fmt.Println("Populate Kubernetes configure file")
	oldKubeConfig, err := kubeconfig.ReadConfigOrNew(constants.GetK8sConfigPath())
	if err != nil {
		return err
	}

	client, err := sftp.NewClient(c)
	if err != nil {
		return err
	}
	defer client.Close()

	srcFile, err := client.Open("/etc/kubernetes/admin.conf")
	if err != nil {
		return err
	}

	os.MkdirAll(constants.GetK8sConfigFolder(), os.ModePerm)
	dstFile, err := os.OpenFile(constants.GetK8sTmpConfigPath(), os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	// flush in-memory copy
	err = dstFile.Sync()
	if err != nil {
		return err
	}

	newKubeConfig, err := kubeconfig.ReadConfigOrNew(constants.GetK8sTmpConfigPath())
	if err != nil {
		return err
	}

	for clusterName, cluster := range newKubeConfig.Clusters {
		oldKubeConfig.Clusters[clusterName] = cluster
	}

	for _, context := range newKubeConfig.Contexts {
		oldKubeConfig.Contexts["kubev-cluster"] = context
		oldKubeConfig.CurrentContext = "kubev-cluster"
	}

	for userName, authInfo := range newKubeConfig.AuthInfos {
		oldKubeConfig.AuthInfos[userName] = authInfo
	}

	if err := kubeconfig.WriteConfig(oldKubeConfig, constants.GetK8sConfigPath()); err != nil {
		return err
	}

	utils.DeleteFile(constants.GetK8sTmpConfigPath())

	return nil
}
