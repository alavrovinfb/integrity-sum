package models

import (
	"k8s.io/client-go/kubernetes"
)

type KuberData struct {
	Clientset  *kubernetes.Clientset
	Namespace  string
	TargetName string
	TargetType string
}

type DataFromK8sAPI struct {
	KuberData      *KuberData
	DeploymentData *DeploymentData
}
