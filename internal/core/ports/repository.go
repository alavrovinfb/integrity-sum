package ports

import (
	"github.com/ScienceSoft-Inc/integrity-sum/internal/core/models"
)

//go:generate mockgen -source=repository.go -destination=mocks/mock_repository.go

type IAppRepository interface {
	IsExistDeploymentNameInDB(deploymentName string) (bool, error)
	// SaveHashData(allHashData []*api.HashData, deploymentData *models.DeploymentData) error
	GetHashData(dirFiles string, algorithm string, deploymentData *models.DeploymentData) ([]*models.HashDataFromDB, error)
	DeleteFromTable(nameDeployment string) error
}
