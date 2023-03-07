package repositories

import (
	"database/sql"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/models"
	"github.com/ScienceSoft-Inc/integrity-sum/pkg/api"
	"github.com/sirupsen/logrus"
)

type HashStorageRepository struct {
	logger *logrus.Logger
	db     *sql.DB
}

func NewHashStorageRepository(logger *logrus.Logger, db *sql.DB) *HashStorageRepository {
	return &HashStorageRepository{
		logger: logger,
		db:     db,
	}
}

// Create iterates through all elements of the slice and triggers the save to database function
func (hr HashStorageRepository) Create(allHashData []*api.HashData, deploymentData *models.DeploymentData) error {
	tx, err := hr.db.Begin()
	if err != nil {
		hr.logger.Error("err while saving data in database ", err)
		return err
	}
	query := `INSERT INTO filehashes (full_file_name, hash_sum, algorithm, name_pod, release_id)
	VALUES($1,$2,$3,$4,(SELECT id from releases WHERE name=$5));`

	for _, hash := range allHashData {
		_, err = tx.Exec(query, hash.FullFileName, hash.Hash, hash.Algorithm, deploymentData.NamePod, deploymentData.NameDeployment)
		if err != nil {
			err = tx.Rollback()
			if err != nil {
				hr.logger.Error("err in Rollback", err)
				return err
			}
			hr.logger.Error("err while save data in database ", err)
			return err
		}
	}
	return tx.Commit()
}

// GetHashData retrieves data from the database using the path and algorithm
func (hr HashStorageRepository) Get(dirFiles, algorithm string, deploymentData *models.DeploymentData) ([]*models.HashData, error) {
	var allHashDataFromDB []*models.HashData

	query := `SELECT id,full_file_name, hash_sum, algorithm, name_pod
		FROM filehashes WHERE full_file_name LIKE $1 and algorithm=$2 and name_pod=$3`
	rows, err := hr.db.Query(query, "%"+dirFiles+"%", algorithm, deploymentData.NamePod)
	if err != nil {
		hr.logger.Error(err)
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var hashDataFromDB models.HashData
		err = rows.Scan(&hashDataFromDB.ID, &hashDataFromDB.FullFileName,
			&hashDataFromDB.Hash, &hashDataFromDB.Algorithm, &hashDataFromDB.NamePod)
		if err != nil {
			hr.logger.Error(err)
			return nil, err
		}
		allHashDataFromDB = append(allHashDataFromDB, &hashDataFromDB)
	}

	return allHashDataFromDB, nil
}
