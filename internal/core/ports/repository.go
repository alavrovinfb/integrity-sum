package ports

import (
	"github.com/integrity-sum/internal/core/models"
	"github.com/integrity-sum/pkg/api"
)

//go:generate mockgen -source=repository.go -destination=mocks/mock_repository.go

type IAppRepository interface {
	IsExistDeploymentNameInDB(deploymentName string) (bool, error)
	SaveHashData(allHashData []*api.HashData, deploymentData *models.DeploymentData) error
	GetHashData(dirFiles string, algorithm string, deploymentData *models.DeploymentData) ([]*models.HashDataFromDB, error)
	DeleteFromTable(nameDeployment string) error
}
