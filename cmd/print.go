package cmd

import (
	"fmt"

	"github.com/jeffwubj/kubev/pkg/kubev/deployer"
	"github.com/jeffwubj/kubev/pkg/kubev/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var printCmd = &cobra.Command{
	Use:   "print",
	Short: "Print join cmd for kubeadm",
	Long:  ``,
	Run:   runPrint,
}

func init() {
	rootCmd.AddCommand(printCmd)
}

func runPrint(cmd *cobra.Command, args []string) {
	if !utils.FileExists(viper.ConfigFileUsed()) {
		fmt.Println("There is no config file, run config and deploy before scale")
		return
	}

	vms, err := utils.ReadK8sNodes()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	joincmd, err := deployer.GetKubeAdmJoinCommand(vms.MasterNode)
	if err != nil {
		fmt.Printf("Failed to join master node: %s\n", err.Error())
		return
	}
	fmt.Println(joincmd)
}
