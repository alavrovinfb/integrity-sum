package data

import (
	"database/sql"
	"fmt"
	"github.com/ScienceSoft-Inc/integrity-sum/pkg/k8s"
	"github.com/sirupsen/logrus"
	"strings"
)

//go:generate mockgen -source=hash_storage.go -destination=mocks/mock_hash_storage.go

type IHashStorage interface {
	PrepareQuery(allHashData []*HashData, deploymentData *k8s.DeploymentData) (string, []any)
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

// PrepareQuery creates a query and a set of arguments for preparing data for insertion into the database
func (hs HashStorage) PrepareQuery(allHashData []*HashData, deploymentData *k8s.DeploymentData) (string, []any) {
	fieldsCount := 5
	defaultHashCount := len(allHashData)
	valuesString := make([]string, 0, defaultHashCount)
	args := make([]any, 0, defaultHashCount*fieldsCount)

	i := 0
	for _, v := range allHashData {
		valuesString = append(valuesString,
			fmt.Sprintf("($%d,$%d,$%d,$%d,(SELECT id from releases WHERE name=$%d))",
				i*fieldsCount+1,
				i*fieldsCount+2,
				i*fieldsCount+3,
				i*fieldsCount+4,
				i*fieldsCount+5,
			))
		args = append(args,
			v.FullFileName,
			v.Hash,
			v.Algorithm,
			deploymentData.NamePod,
			deploymentData.NameDeployment,
		)
		i++
	}

	query := `INSERT INTO filehashes (full_file_name,hash_sum,algorithm,name_pod,release_id) VALUES `
	query += strings.Join(valuesString, ",")

	return query, args
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
