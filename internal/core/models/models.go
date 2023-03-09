package models

type HashDataFromDB struct {
	ID             int
	Hash           string
	FileName       string
	FullFilePath   string
	Algorithm      string
	ImageContainer string
	NamePod        string
	NameDeployment string
}

type KubeData struct {
	Namespace  string
	TargetName string
	TargetType string
}

type DeploymentData struct {
	Image          string
	NamePod        string
	Timestamp      string
	NameDeployment string
	ReleaseName    string
}

type DataFromK8sAPI struct {
	KubeData       *KubeData
	DeploymentData *DeploymentData
}
