package ports

import (
	"github.com/ScienceSoft-Inc/integrity-sum/internal/data"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/models"
)

//go:generate mockgen -source=data.go -destination=mocks/mock_data.go

type IHashStorage interface {
	Create(allHashData []*data.HashData, deploymentData *models.DeploymentData) error
	Get(dirPath string, deploymentData *models.DeploymentData) ([]*data.HashData, error)
}

type IReleaseStorage interface {
	Create(deploymentData *models.DeploymentData) error
	Get(deploymentData *models.DeploymentData) (*data.Release, error)
	Delete(nameDeployment string) error
	DeleteOldData() error
	IsExistDeploymentNameInDB(deploymentName string) bool
}
