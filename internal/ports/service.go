package ports

import (
	"context"
	"os"
	"sync"

	"github.com/ScienceSoft-Inc/integrity-sum/internal/models"
	"github.com/ScienceSoft-Inc/integrity-sum/pkg/api"
)

//go:generate mockgen -source=service.go -destination=mocks/mock_service.go

type IAppService interface {
	GetPID(procName string) (int, error)
	IsExistDeploymentNameInDB(deploymentName string) bool //+
	LaunchHasher(ctx context.Context, dirPath string, sig chan os.Signal) []*api.HashData
	Start(ctx context.Context, dirPath string, sig chan os.Signal, deploymentData *models.DeploymentData) error
	Check(ctx context.Context, dirPath string, sig chan os.Signal, deploymentData *models.DeploymentData, kuberData *models.KuberData) error
}

type IHashService interface {
	IsDataChanged(currentHashData []*api.HashData, hashSumFromDB []*models.HashData, dataFromDBbyRelease *models.Release, deploymentData *models.DeploymentData) bool
	CreateHash(path string) (*api.HashData, error)
	WorkerPool(jobs chan string, results chan *api.HashData)
	Worker(wg *sync.WaitGroup, jobs <-chan string, results chan<- *api.HashData)
}

type IKuberService interface {
	GetDataFromK8sAPI() (*models.DataFromK8sAPI, error)
	ConnectionToK8sAPI() (*models.KuberData, error)
	GetDataFromDeployment(kuberData *models.KuberData) (*models.DeploymentData, error)
	RolloutDeployment(kuberData *models.KuberData) error
}

type IHashStorageService interface {
	Create(allHashData []*api.HashData, deploymentData *models.DeploymentData) error
	Get(dirPath string, deploymentData *models.DeploymentData) ([]*models.HashData, error)
}

type IReleaseStorageService interface {
	Create(deploymentData *models.DeploymentData) error
	Get(deploymentData *models.DeploymentData) (*models.Release, error)
	Delete(nameDeployment string) error
}
