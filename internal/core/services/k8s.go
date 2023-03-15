package services

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
	"strings"

	"github.com/ScienceSoft-Inc/integrity-sum/internal/core/models"
)

type KubeClient struct {
	logger    *logrus.Logger
	clientset *kubernetes.Clientset
}

// NewKubeService creates a new service for working with the Kubernetes API
func NewKubeService(logger *logrus.Logger) *KubeClient {
	return &KubeClient{
		logger: logger,
	}
}

// Connect to Kubernetes API
func (ks *KubeClient) Connect() (*kubernetes.Clientset, error) {
	if ks.clientset != nil {
		return ks.clientset, nil
	}

	ks.logger.Info("### 🌀 Attempting to use in cluster config")
	config, err := rest.InClusterConfig()
	if err != nil {
		ks.logger.Error(err)
		return nil, err
	}

	ks.logger.Info("### 💻 Connecting to Kubernetes API, using host: ", config.Host)
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		ks.logger.Error(err)
		return nil, err
	}

	ks.clientset = clientset

	return clientset, nil
}

// GetDataFromK8sAPI returns data from deployment
func (ks *KubeClient) GetDataFromK8sAPI() (*models.DataFromK8sAPI, error) {
	kubeData, err := ks.GetKubeData()
	if err != nil {
		ks.logger.Errorf("can't connect to K8sAPI: %s", err)
		return nil, err
	}
	deploymentData, err := ks.GetDataFromDeployment(kubeData)
	if err != nil {
		ks.logger.Errorf("error while getting data from kuberAPI %s", err)
		return nil, err
	}

	if err != nil {
		ks.logger.Errorf("err while getting data from configMap K8sAPI %s", err)
		return &models.DataFromK8sAPI{}, err
	}

	dataFromK8sAPI := &models.DataFromK8sAPI{
		KubeData:       kubeData,
		DeploymentData: deploymentData,
	}

	return dataFromK8sAPI, nil
}

// GetKubeData returns kubeData
func (ks *KubeClient) GetKubeData() (*models.KubeData, error) {
	clientset, err := ks.Connect()
	if err != nil {
		return nil, err
	}

	namespaceBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		ks.logger.Error(err)
		return nil, err
	}
	namespace := string(namespaceBytes)

	podName := os.Getenv("POD_NAME")

	targetName := func(podName string) string {
		elements := strings.Split(podName, "-")
		newElements := elements[:len(elements)-2]
		return strings.Join(newElements, "-")
	}(podName)
	if targetName == "" {
		ks.logger.Fatalln("### 💥 Env var DEPLOYMENT_NAME was not set")
	}
	targetType := os.Getenv("DEPLOYMENT_TYPE")
	kubeData := &models.KubeData{
		Clientset:  clientset,
		Namespace:  namespace,
		TargetName: targetName,
		TargetType: targetType,
	}
	return kubeData, nil
}

// GetDataFromDeployment returns data from deployment
func (ks *KubeClient) GetDataFromDeployment(kubeData *models.KubeData) (*models.DeploymentData, error) {
	allDeploymentData, err := kubeData.Clientset.AppsV1().Deployments(kubeData.Namespace).Get(context.Background(), kubeData.TargetName, metav1.GetOptions{})
	if err != nil {
		ks.logger.Error("err while getting data from kuberAPI ", err)
		return nil, err
	}

	deploymentData := &models.DeploymentData{
		NamePod:        os.Getenv("POD_NAME"),
		Timestamp:      fmt.Sprintf("%v", allDeploymentData.CreationTimestamp),
		NameDeployment: kubeData.TargetName,
	}

	for _, v := range allDeploymentData.Spec.Template.Spec.Containers {
		deploymentData.Image = v.Image
	}

	if value, ok := allDeploymentData.Annotations["meta.helm.sh/release-name"]; ok {
		deploymentData.ReleaseName = value
	}

	return deploymentData, nil
}

// RestartPod restarts pod
func (ks *KubeClient) RestartPod(kubeData *models.KubeData) error {
	// TODO: maybe get from deploymentData
	pName := os.Getenv("POD_NAME")
	pNamespace := os.Getenv("POD_NAMESPACE")

	// Deleting pod to force a restart
	err := kubeData.Clientset.CoreV1().Pods(pNamespace).Delete(context.Background(), pName, metav1.DeleteOptions{})
	if err != nil {
		ks.logger.Printf("### 👎 Warning: Failed to delete pod %v, restart failed: %v", pName, err)
		return err
	} else {
		ks.logger.Printf("### ✅ Pod %v was forced to be restartd", pName)
	}
	return nil
}
