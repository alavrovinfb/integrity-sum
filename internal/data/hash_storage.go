package data

import (
	"database/sql"

	"github.com/ScienceSoft-Inc/integrity-sum/pkg/k8s"
	"github.com/sirupsen/logrus"
)

//go:generate mockgen -source=hash_storage.go -destination=mocks/mock_hash_storage.go

type IHashStorage interface {
	Create(allHashData []*HashData, deploymentData *k8s.DeploymentData) error
	Get(dirPath string, deploymentData *k8s.DeploymentData) ([]*HashData, error)
}

type HashData struct {
	ID           int
	Hash         string
	FullFileName string
	Algorithm    string
	NamePod      string
}

type HashStorage struct {
	db     *sql.DB
	alg    string
	logger *logrus.Logger
}

// NewHashStorage creates a new HashStorage structure to work with the database table.
func NewHashStorage(db *sql.DB, alg string, logger *logrus.Logger) *HashStorage {
	return &HashStorage{
		db:     db,
		alg:    alg,
		logger: logger,
	}
}

// Create saves data to the database
func (hs HashStorage) Create(allHashData []*HashData, deploymentData *k8s.DeploymentData) error {
	tx, err := hs.db.Begin()
	if err != nil {
		hs.logger.Error("err while saving data in database ", err)
		return err
	}
	query := `INSERT INTO filehashes (full_file_name, hash_sum, algorithm, name_pod, release_id)
	VALUES($1,$2,$3,$4,(SELECT id from releases WHERE name=$5));`

	for _, hash := range allHashData {
		_, err = tx.Exec(query, hash.FullFileName, hash.Hash, hash.Algorithm, deploymentData.NamePod, deploymentData.NameDeployment)
		if err != nil {
			err = tx.Rollback()
			if err != nil {
				hs.logger.Error("err in Rollback", err)
				return err
			}
			hs.logger.Error("err while save data in database ", err)
			return err
		}
	}
	return tx.Commit()
}

// Get gets data from the database
func (hs HashStorage) Get(dirFiles string, deploymentData *k8s.DeploymentData) ([]*HashData, error) {
	var allHashDataFromDB []*HashData

	query := `SELECT id,full_file_name, hash_sum, algorithm, name_pod
		FROM filehashes WHERE full_file_name LIKE $1 and algorithm=$2 and name_pod=$3`
	rows, err := hs.db.Query(query, "%"+dirFiles+"%", hs.alg, deploymentData.NamePod)
	if err != nil {
		hs.logger.Error(err)
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var hashDataFromDB HashData
		err = rows.Scan(&hashDataFromDB.ID, &hashDataFromDB.FullFileName,
			&hashDataFromDB.Hash, &hashDataFromDB.Algorithm, &hashDataFromDB.NamePod)
		if err != nil {
			hs.logger.Error(err)
			return nil, err
		}
		allHashDataFromDB = append(allHashDataFromDB, &hashDataFromDB)
	}

	return allHashDataFromDB, nil
}
