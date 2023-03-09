package models

import (
	"k8s.io/client-go/kubernetes"
)

type KubeData struct {
	Clientset  *kubernetes.Clientset
	Namespace  string
	TargetName string
	TargetType string
}

type DataFromK8sAPI struct {
	KubeData       *KubeData
	DeploymentData *DeploymentData
}
