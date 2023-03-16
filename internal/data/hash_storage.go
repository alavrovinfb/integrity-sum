package data

import (
	"database/sql"
	"fmt"
	"strings"
)

type HashData struct {
	Hash         string
	FullFileName string
	Algorithm    string
	PodName      string
	ReleaseId    int
}

type HashDataOutput struct {
	ID           int
	Hash         string
	FullFileName string
	Algorithm    string
	NamePod      string
	ReleaseId    int
}

type HashStorage struct {
	db *sql.DB
}

func NewHashData(db *sql.DB) *HashStorage {
	return &HashStorage{db: db}
}

// PrepareQuery creates a query and a set of arguments for preparing data for insertion into the database
func (hs HashStorage) PrepareQuery(hashData []*HashData, releaseName string) (string, []any) {
	fieldsCount := 5
	defaultHashCount := len(hashData)
	valuesString := make([]string, 0, defaultHashCount)
	args := make([]any, 0, defaultHashCount*fieldsCount)

	i := 0
	for _, v := range hashData {
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
			v.PodName,
			releaseName,
		)
		i++
	}

	query := `INSERT INTO filehashes (full_file_name,hash_sum,algorithm,name_pod,release_id) VALUES `
	query += strings.Join(valuesString, ",")

	return query, args
}

// Get gets data from the database
func (hs HashStorage) Get(alg string, dirName string, podName string) ([]*HashDataOutput, error) {
	var dataHashes []*HashDataOutput

	query := `SELECT id,full_file_name, hash_sum, algorithm, name_pod, release_id
		FROM filehashes WHERE full_file_name LIKE $1 and algorithm=$2 and name_pod=$3`
	rows, err := hs.db.Query(query, "%"+dirName+"%", alg, podName)
	if err != nil {
		// hs.logger.Error(err)
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var dataHash HashDataOutput
		err = rows.Scan(&dataHash.ID, &dataHash.FullFileName,
			&dataHash.Hash, &dataHash.Algorithm, &dataHash.NamePod, &dataHash.ReleaseId)
		if err != nil {
			// hs.logger.Error(err)
			return nil, err
		}
		dataHashes = append(dataHashes, &dataHash)
	}

	return dataHashes, nil
}
