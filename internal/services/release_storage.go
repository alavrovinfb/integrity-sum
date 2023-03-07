package services

import (
	"github.com/ScienceSoft-Inc/integrity-sum/internal/models"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/ports"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/repositories"
	"github.com/sirupsen/logrus"
)

type ReleaseStorageService struct {
	ports.IReleaseStorageRepository
	alg    string
	logger *logrus.Logger
}

// NewHashStorageService creates a new struct HashService
func NewReleaseStorageService(r *repositories.ReleaseStorageRepository, alg string, logger *logrus.Logger) *ReleaseStorageService {
	return &ReleaseStorageService{
		IReleaseStorageRepository: r,
		alg:                       alg,
		logger:                    logger,
	}
}

// SaveHashData accesses the repository to save data to the database
func (rr ReleaseStorageService) Create(deploymentData *models.DeploymentData) error {
	err := rr.IReleaseStorageRepository.Create(deploymentData)
	if err != nil {
		rr.logger.Error("error while saving data to database", err)
		return err
	}
	return nil
}

// GetHashData accesses the repository to get data from the database
func (rr ReleaseStorageService) Get(deploymentData *models.DeploymentData) (*models.Release, error) {
	hashData, err := rr.IReleaseStorageRepository.Get(deploymentData)
	if err != nil {
		rr.logger.Error("hashData service didn't get hashData sum", err)
		return nil, err
	}
	return hashData, nil
}

func (hs ReleaseStorageService) Delete(nameDeployment string) error {
	err := hs.IReleaseStorageRepository.Delete(nameDeployment)
	if err != nil {
		hs.logger.Error("err while deleting rows in database", err)
		return err
	}
	return nil
}

// IsExistDeploymentNameInDB checks if the database is empty
//func (hs *ReleaseStorageService) IsExist(deploymentName string) bool {
//	isEmptyDB, err := hs.IHashStorageRepository.IsExist(deploymentName)
//	if err != nil && err != sql.ErrNoRows {
//		hs.logger.Fatalf("database check error %s", err)
//	}
//	return isEmptyDB
//}
