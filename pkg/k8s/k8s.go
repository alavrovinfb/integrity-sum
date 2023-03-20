package k8s

import (
	"context"
	"fmt"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/logger"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
	"strings"
)

//go:generate mockgen -source=k8s.go -destination=mocks/mock_k8s.go

type IKuberService interface {
	Connect() error
	GetDataFromK8sAPI() (*DataFromK8sAPI, error)
	GetDataFromDeployment() (*DeploymentData, error)
	RestartPod() error
}

type KubeData struct {
	Namespace    string
	TargetName   string
	TargetType   string
	PodName      string
	PodNamespace string
}

type DeploymentData struct {
	Image          string
	NamePod        string
	Timestamp      string
	NameDeployment string
	ReleaseName    string
}

// TODO: seems like this struct is not needed
type DataFromK8sAPI struct {
	KubeData       *KubeData
	DeploymentData *DeploymentData
}

type KubeClient struct {
	logger    *logrus.Logger
	clientset *kubernetes.Clientset
}

var kubeData *KubeData

func init() {
	initKubeData()
}

// initKubeData initializes kubeData global variable
func initKubeData() {
	log := logger.Init(viper.GetString("verbose"))
	namespaceBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		log.Error(err)
	}
	namespace := string(namespaceBytes)
	podName := os.Getenv("POD_NAME")

	targetName := func(podName string) string {
		elements := strings.Split(podName, "-")
		newElements := elements[:len(elements)-2]
		return strings.Join(newElements, "-")
	}(podName)
	if targetName == "" {
		log.Fatalln("### 💥 Env var DEPLOYMENT_NAME was not set")
	}

	targetType := os.Getenv("DEPLOYMENT_TYPE")
	pNamespace := os.Getenv("POD_NAMESPACE")

	kubeData = &KubeData{
		Namespace:    namespace,
		TargetName:   targetName,
		TargetType:   targetType,
		PodName:      podName,
		PodNamespace: pNamespace,
	}
}

// NewKubeService creates a new service for working with the Kubernetes API
func NewKubeService(logger *logrus.Logger) *KubeClient {
	return &KubeClient{
		logger: logger,
	}
}

// Connect to Kubernetes API
func (ks *KubeClient) Connect() error {
	ks.logger.Info("### 🌀 Attempting to use in cluster config")
	config, err := rest.InClusterConfig()
	if err != nil {
		ks.logger.Error(err)
		return err
	}

	ks.logger.Info("### 💻 Connecting to Kubernetes API, using host: ", config.Host)
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		ks.logger.Error(err)
		return err
	}
	ks.clientset = clientset

	return nil
}

// GetDataFromK8sAPI returns data from deployment
// TODO: seems like it's not used
func (ks *KubeClient) GetDataFromK8sAPI() (*DataFromK8sAPI, error) {
	//kubeData, err := ks.GetKubeData()
	//if err != nil {
	//	ks.logger.Errorf("can't connect to K8sAPI: %s", err)
	//	return nil, err
	//}

	deploymentData, err := ks.GetDataFromDeployment()
	if err != nil {
		ks.logger.Errorf("error while getting data from kuberAPI %s", err)
		return nil, err
	}

	if err != nil {
		ks.logger.Errorf("err while getting data from configMap K8sAPI %s", err)
		return &DataFromK8sAPI{}, err
	}

	dataFromK8sAPI := &DataFromK8sAPI{
		KubeData:       kubeData,
		DeploymentData: deploymentData,
	}

	return dataFromK8sAPI, nil
}

// GetDataFromDeployment returns data from deployment
func (ks *KubeClient) GetDataFromDeployment() (*DeploymentData, error) {
	allDeploymentData, err := ks.clientset.AppsV1().Deployments(kubeData.Namespace).Get(
		context.Background(),
		kubeData.TargetName,
		metav1.GetOptions{},
	)

	if err != nil {
		ks.logger.Error("err while getting data from kuberAPI ", err)
		return nil, err
	}

	deploymentData := &DeploymentData{
		NamePod:        kubeData.PodName,
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
func (ks *KubeClient) RestartPod() error {
	// Deleting pod to force a restart
	err := ks.clientset.CoreV1().Pods(kubeData.PodNamespace).Delete(context.Background(), kubeData.PodName, metav1.DeleteOptions{})

	if err != nil {
		ks.logger.Printf("### 👎 Warning: Failed to delete pod %v, restart failed: %v", kubeData.PodName, err)
		return err
	}

	ks.logger.Printf("### ✅ Pod %v was forced to be restartd", kubeData.PodName)
	return nil
}
