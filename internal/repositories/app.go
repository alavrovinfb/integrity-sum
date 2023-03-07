package repositories

import (
	"database/sql"
	"github.com/sirupsen/logrus"
)

type AppRepository struct {
	logger *logrus.Logger
	db     *sql.DB
}

func NewAppRepository(logger *logrus.Logger, db *sql.DB) *AppRepository {
	return &AppRepository{
		logger: logger,
		db:     db,
	}
}

// IsExistDeploymentNameInDB checks if the base is empty
func (ar AppRepository) IsExistDeploymentNameInDB(deploymentName string) (bool, error) {
	var id int64
	query := `SELECT id FROM releases WHERE name=$1 LIMIT 1;`
	row := ar.db.QueryRow(query, deploymentName)
	err := row.Scan(&id)
	if err != nil {
		ar.logger.Error("err while scan row in database ", err)
		return false, err
	}

	return true, nil
}
