package services

import (
	"github.com/ScienceSoft-Inc/integrity-sum/internal/models"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/ports"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/repositories"
	"github.com/ScienceSoft-Inc/integrity-sum/pkg/api"
	"github.com/sirupsen/logrus"
)

type HashStorageService struct {
	ports.IHashStorageRepository
	alg    string
	logger *logrus.Logger
}

// NewHashStorageService creates a new struct HashService
func NewHashStorageService(r *repositories.HashStorageRepository, alg string, logger *logrus.Logger) *HashStorageService {
	return &HashStorageService{
		IHashStorageRepository: r,
		alg:                    alg,
		logger:                 logger,
	}
}

// SaveHashData accesses the repository to save data to the database
func (hs HashStorageService) Create(allHashData []*api.HashData, deploymentData *models.DeploymentData) error {
	err := hs.IHashStorageRepository.Create(allHashData, deploymentData)
	if err != nil {
		hs.logger.Error("error while saving data to database", err)
		return err
	}
	return nil
}

// GetHashData accesses the repository to get data from the database
func (hs HashStorageService) Get(dirFiles string, deploymentData *models.DeploymentData) ([]*models.HashData, error) {
	hashData, err := hs.IHashStorageRepository.Get(dirFiles, hs.alg, deploymentData)
	if err != nil {
		hs.logger.Error("hashData service didn't get hashData sum", err)
		return nil, err
	}
	return hashData, nil
}
