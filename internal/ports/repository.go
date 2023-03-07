package ports

import (
	"github.com/ScienceSoft-Inc/integrity-sum/internal/models"
	"github.com/ScienceSoft-Inc/integrity-sum/pkg/api"
)

//go:generate mockgen -source=repository.go -destination=mocks/mock_repository.go

type IAppRepository interface {
	IsExistDeploymentNameInDB(deploymentName string) (bool, error)
}

type IHashStorageRepository interface {
	Create(allHashData []*api.HashData, deploymentData *models.DeploymentData) error
	Get(dirFiles string, algorithm string, deploymentData *models.DeploymentData) ([]*models.HashData, error)
}

type IReleaseStorageRepository interface {
	Create(deploymentData *models.DeploymentData) error
	Get(deploymentData *models.DeploymentData) (*models.Release, error)
	Delete(nameDeployment string) error
	DeleteOldData() error
}
