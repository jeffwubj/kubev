package constants

import (
	"fmt"
	"path"
	"runtime"

	homedir "github.com/mitchellh/go-homedir"
)

const KubeletService = `
[Unit]
Description=kubelet: The Kubernetes Node Agent
Documentation=http://kubernetes.io/docs/

[Service]
ExecStart=/usr/bin/kubelet
Restart=always
StartLimitInterval=0
RestartSec=10

[Install]
WantedBy=multi-user.target
`

const KubeletSystemd = `
# Note: This dropin only works with kubeadm and kubelet v1.11+
[Service]
Environment="KUBELET_KUBECONFIG_ARGS=--bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --kubeconfig=/etc/kubernetes/kubelet.conf"
Environment="KUBELET_CONFIG_ARGS=--config=/var/lib/kubelet/config.yaml"
# This is a file that "kubeadm init" and "kubeadm join" generates at runtime, populating the KUBELET_KUBEADM_ARGS variable dynamically
EnvironmentFile=-/var/lib/kubelet/kubeadm-flags.env
# This is a file that the user can use for overrides of the kubelet args as a last resort. Preferably, the user should use
# the .NodeRegistration.KubeletExtraArgs object in the configuration files instead. KUBELET_EXTRA_ARGS should be sourced from this file.
EnvironmentFile=-/etc/sysconfig/kubelet
ExecStart=
ExecStart=/usr/bin/kubelet $KUBELET_KUBECONFIG_ARGS $KUBELET_CONFIG_ARGS $KUBELET_KUBEADM_ARGS $KUBELET_EXTRA_ARGS
`

// TODO: lots of todo's...
const KubeAdmInit = `
sysctl net.bridge.bridge-nf-call-iptables=1 &&
kubeadm reset -f &&
kubeadm init --image-repository registry.aliyuncs.com/google_containers --kubernetes-version v1.13.0 &&
mkdir -p /root/.kube &&
cp /etc/kubernetes/admin.conf /root/.kube/config &&
kubectl taint nodes --all node-role.kubernetes.io/master- &&
kubectl apply -f "https://cloud.weave.works/k8s/net?k8s-version=$(kubectl version | base64 | tr -d '\n')"
`

const DockerService = `
[Unit]
Description=Docker Application Container Engine
Documentation=https://docs.docker.com
After=network-online.target
Wants=network-online.target

[Service]
Type=notify
# the default is not to use systemd for cgroups because the delegate issues still
# exists and systemd currently does not support the cgroup feature set required
# for containers run by docker
ExecStart=/usr/bin/dockerd -H unix:///var/run/docker.sock -H tcp://0.0.0.0:2375
ExecReload=/bin/kill -s HUP $MAINPID
# Having non-zero Limit*s causes performance problems due to accounting overhead
# in the kernel. We recommend using cgroups to do container-local accounting.
LimitNOFILE=infinity
LimitNPROC=infinity
LimitCORE=infinity
# Uncomment TasksMax if your systemd version supports it.
# Only systemd 226 and above support this version.
#TasksMax=infinity
TimeoutStartSec=0
# set delegate yes so that systemd does not reset the cgroups of docker containers
Delegate=yes
# kill only the docker process, not all processes in the cgroup
KillMode=process
# restart the docker process if it exits prematurely
Restart=on-failure
StartLimitBurst=3
StartLimitInterval=60s

[Install]
WantedBy=multi-user.target
`

const (
	DefaultPhotonVersion     = "v2.0"
	KubeAdmBinaryName        = "kubeadm"
	KubeCtlBinaryName        = "kubectl"
	DockerBinaryName         = "docker"
	CriCtlBinaryName         = "crictl"
	CNIKits                  = "cni.tgz"
	GuestKubeCtlBinaryName   = "kubectl.guest"
	KubeletBinaryName        = "kubelet"
	PhotonOVAName            = "photon.ova"
	DefaultVMName            = "Photon"
	PhotonVMUsername         = "root"
	PhotonVMOriginalPassword = "changeme"
	PhotonVMPassword         = "kubernetes"
	KubeletServiceFile       = "/etc/systemd/system/kubelet.service"
	KubeletSystemdConfFile   = "/etc/systemd/system/kubelet.service.d/10-kubeadm.conf"
	DockerServiceFile        = "/usr/lib/systemd/system/docker.service"
	DefaultVMTemplateName    = "kube-template"
)

func GetHomeFolder() string {
	home, _ := homedir.Dir()
	return home
}

func GetKubeVHomeFolder() string {
	return path.Join(GetHomeFolder(), ".kubev")
}

func GetK8sNodesConfigFilePath() string {
	return path.Join(GetKubeVHomeFolder(), "config.json")
}

func GetLocalK8sKitPath(binaryName, version string) string {
	return path.Join(GetKubeVHomeFolder(), "cache", binaryName, version)
}

func GetVMPrivateKeyPath() string {
	return path.Join(GetKubeVHomeFolder(), "id_rsa")
}

func GetVMPublicKeyPath() string {
	return GetVMPrivateKeyPath() + ".pub"
}

func GetLocalK8sKitFilePath(binaryName, version string) string {
	if binaryName == DockerBinaryName {
		return path.Join(GetKubeVHomeFolder(), "cache", binaryName, version, binaryName, binaryName)
	}
	return path.Join(GetKubeVHomeFolder(), "cache", binaryName, version, binaryName)
}

func GetK8sKitReleaseURL(binaryName, version string) string {
	if binaryName == KubeCtlBinaryName {
		return fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/%s/bin/%s/amd64/kubectl", version, runtime.GOOS)
	} else if binaryName == PhotonOVAName {
		// TODO 3.0 not GA just BETA
		return "https://bintray.com/vmware/photon/download_file?file_path=2.0%2FGA%2Fova%2Fphoton-custom-hw11-2.0-304b817.ova"
	} else if binaryName == CriCtlBinaryName {
		return "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.12.0/crictl-v1.12.0-linux-amd64.tar.gz"
	} else if binaryName == CNIKits {
		return "https://github.com/containernetworking/plugins/releases/download/v0.7.4/cni-plugins-amd64-v0.7.4.tgz"
	} else if binaryName == DockerBinaryName {
		return "https://download.docker.com/mac/static/stable/x86_64/docker-17.06.0-ce.tgz"
	} else if version == "v1.13.0" {
		// Aliyun mirror has no 1.13 kubeadm
		if binaryName == GuestKubeCtlBinaryName {
			binaryName = KubeCtlBinaryName
		}
		return fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/%s/bin/linux/%s/%s", version, runtime.GOARCH, binaryName)
	}
	return GetKubernetesReleaseURL(binaryName, version)
}

func GetKubernetesReleaseURL(binaryName, version string) string {
	return fmt.Sprintf("https://kubernetes.oss-cn-hangzhou.aliyuncs.com/kubernetes-release/release/%s/bin/linux/%s/%s", version, runtime.GOARCH, binaryName)
}
