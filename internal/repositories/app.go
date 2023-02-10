package repositories

import (
	"github.com/sirupsen/logrus"

	"github.com/integrity-sum/internal/core/ports"
)

type AppRepository struct {
	ports.IHashRepository
	logger *logrus.Logger
}

func NewAppRepository(logger *logrus.Logger) *AppRepository {
	return &AppRepository{
		IHashRepository: NewHashRepository(logger),
		logger:          logger,
	}
}

// IsExistDeploymentNameInDB checks if the base is empty
func (ar AppRepository) IsExistDeploymentNameInDB(deploymentName string) (bool, error) {
	db, err := ConnectionToDB(ar.logger)
	if err != nil {
		ar.logger.Error("failed to connection to database %s", err)
		return false, err
	}
	defer db.Close()

	var count int
	query := "SELECT COUNT(*) FROM hashfiles WHERE name_deployment=$1 LIMIT 1;"
	row := db.QueryRow(query, deploymentName)
	err = row.Scan(&count)
	if err != nil {
		ar.logger.Error("err while scan row in database ", err)
		return false, err
	}

	if count < 1 {
		return true, nil
	}
	return false, nil
}
