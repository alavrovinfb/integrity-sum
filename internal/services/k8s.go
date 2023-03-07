package services

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/ScienceSoft-Inc/integrity-sum/internal/models"
)

type KuberService struct {
	logger *logrus.Logger
}

// NewHashService creates a new struct HashService
func NewKuberService(logger *logrus.Logger) *KuberService {
	return &KuberService{
		logger: logger,
	}
}

func (ks *KuberService) GetDataFromK8sAPI() (*models.DataFromK8sAPI, error) {
	kuberData, err := ks.ConnectionToK8sAPI()
	if err != nil {
		ks.logger.Errorf("can't connection to K8sAPI: %s", err)
		return nil, err
	}
	deploymentData, err := ks.GetDataFromDeployment(kuberData)
	if err != nil {
		ks.logger.Errorf("error get data from kuberAPI %s", err)
		return nil, err
	}

	if err != nil {
		ks.logger.Errorf("err while getting data from configMap K8sAPI %s", err)
		return &models.DataFromK8sAPI{}, err
	}

	dataFromK8sAPI := &models.DataFromK8sAPI{
		KuberData:      kuberData,
		DeploymentData: deploymentData,
	}

	return dataFromK8sAPI, nil
}

func (ks *KuberService) ConnectionToK8sAPI() (*models.KuberData, error) {
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
	kuberData := &models.KuberData{
		Clientset:  clientset,
		Namespace:  namespace,
		TargetName: targetName,
		TargetType: targetType,
	}
	return kuberData, nil
}

func (ks *KuberService) GetDataFromDeployment(kuberData *models.KuberData) (*models.DeploymentData, error) {
	allDeploymentData, err := kuberData.Clientset.AppsV1().Deployments(kuberData.Namespace).Get(context.Background(), kuberData.TargetName, metav1.GetOptions{})
	if err != nil {
		ks.logger.Error("err while getting data from kuberAPI ", err)
		return nil, err
	}

	deploymentData := &models.DeploymentData{
		NamePod:        os.Getenv("POD_NAME"),
		Timestamp:      fmt.Sprintf("%v", allDeploymentData.CreationTimestamp),
		NameDeployment: kuberData.TargetName,
	}

	for _, v := range allDeploymentData.Spec.Template.Spec.Containers {
		deploymentData.Image = v.Image
	}

	if value, ok := allDeploymentData.Annotations["meta.helm.sh/release-name"]; ok {
		deploymentData.ReleaseName = value
	}

	return deploymentData, nil
}

func (ks *KuberService) RolloutDeployment(kuberData *models.KuberData) error {
	patchData := fmt.Sprintf(`{"spec":{"template":{"metadata":{"annotations":{"kubectl.kubernetes.io/restartedAt":"%s"}}}}}`, time.Now().Format(time.RFC3339))
	_, err := kuberData.Clientset.AppsV1().Deployments(kuberData.Namespace).Patch(context.Background(), kuberData.TargetName, types.StrategicMergePatchType, []byte(patchData), metav1.PatchOptions{FieldManager: "kubectl-rollout"})
	if err != nil {
		ks.logger.Printf("### 👎 Warning: Failed to patch %v, restart failed: %v", kuberData.TargetType, err)
		return err
	} else {
		ks.logger.Printf("### ✅ Target %v, named %v was restarted!", kuberData.TargetType, kuberData.TargetName)
	}
	return nil
}
