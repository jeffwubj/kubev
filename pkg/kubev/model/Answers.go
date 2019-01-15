package model

type Answers struct {
	Serverurl         string
	Port              int
	Username          string
	Password          string
	IsVCenter         bool
	Datacenter        string
	Datastore         string
	Resourcepool      string
	Folder            string
	Cpu               int
	Memory            int
	Network           string
	KubernetesVersion string
	WorkerNodes       int
}
