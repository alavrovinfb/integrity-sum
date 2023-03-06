package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/ScienceSoft-Inc/integrity-sum/internal/core/models"
	"github.com/ScienceSoft-Inc/integrity-sum/pkg/api"
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
	var count int
	query := `SELECT COUNT(*) FROM hashfiles WHERE name_deployment=$1;`
	err := ar.db.QueryRow(query, deploymentName).Scan(&count)
	if err != nil {
		ar.logger.Error("err while scan row in database ", err)
		return false, err
	}
	if count == 0 {
		ar.logger.Info("no rows in database")
		return false, nil
	}
	return true, nil
}

// const insertHashQuery = `
// 	INSERT INTO hashfiles (
// 			file_name,
// 			full_file_path,
// 			hash_sum,
// 			algorithm,
// 			name_pod,
// 			image_tag,
// 			time_of_creation,
// 			name_deployment
// 		)
// 	VALUES($1, $2, $3, $4, $5, $6, $7, $8);`

// SaveHashData iterates through all elements of the slice and triggers the save to database function
// func (ar AppRepository) SaveHashData(allHashData []*api.HashData, deploymentData *models.DeploymentData) error {
// 	tx, err := ar.db.Begin()
// 	if err != nil {
// 		ar.logger.Error("open transaction error ", err)
// 		return err
// 	}

// 	for _, hash := range allHashData {
// 		_, err = tx.Exec(insertHashQuery, hash.FileName, hash.FullFilePath, hash.Hash, hash.Algorithm, deploymentData.NamePod,
// 			deploymentData.Image, deploymentData.Timestamp, deploymentData.NameDeployment)
// 		if err != nil {
// 			err = tx.Rollback()
// 			if err != nil {
// 				ar.logger.Error("err in Rollback", err)
// 				return err
// 			}
// 			ar.logger.Error("err while save data in database ", err)
// 			return err
// 		}
// 	}

// 	return tx.Commit()
// }

// SaveHashData iterates through all elements of the slice and save them to the DB
func (ar AppRepository) SaveHashData(allHashData []*api.HashData, deploymentData *models.DeploymentData) error {
	query := `
	INSERT INTO hashfiles (
			file_name,
			full_file_path,
			hash_sum,
			algorithm,
			name_pod,
			image_tag,
			time_of_creation,
			name_deployment
		)
	VALUES `

	fieldsCount := 8
	defaultHashCount := 20
	valuesString := make([]string, 0, defaultHashCount)
	args := make([]any, 0, defaultHashCount*fieldsCount)

	i := 0
	for _, v := range allHashData {
		valuesString = append(valuesString,
			fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
				i*fieldsCount+1, i*fieldsCount+2, i*fieldsCount+3, i*fieldsCount+4,
				i*fieldsCount+5, i*fieldsCount+6, i*fieldsCount+7, i*fieldsCount+8))
		args = append(args,
			v.FileName,
			v.FullFilePath,
			v.Hash,
			v.Algorithm,
			deploymentData.NamePod,
			deploymentData.Image,
			deploymentData.Timestamp,
			deploymentData.NameDeployment,
		)
		i++
	}
	query += strings.Join(valuesString, ", ")
	return BulkInsertTx(context.Background(), query, args...)
}

// TODO: verify bulkInsert with tx.Prepare(pq.CopyIn(...))

// GetHashData retrieves data from the database using the path and algorithm
func (ar AppRepository) GetHashData(dirFiles, algorithm string, deploymentData *models.DeploymentData) ([]*models.HashDataFromDB, error) {
	var allHashDataFromDB []*models.HashDataFromDB

	query := `SELECT id,file_name,full_file_path,hash_sum,algorithm,image_tag,name_pod,name_deployment
		FROM hashfiles WHERE full_file_path LIKE $1 and algorithm=$2 and name_pod=$3`

	rows, err := ar.db.Query(query, "%"+dirFiles+"%", algorithm, deploymentData.NamePod)
	if err != nil {
		ar.logger.Error("err while getting data from database ", err)
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var hashDataFromDB models.HashDataFromDB
		err = rows.Scan(&hashDataFromDB.ID, &hashDataFromDB.FileName, &hashDataFromDB.FullFilePath,
			&hashDataFromDB.Hash, &hashDataFromDB.Algorithm, &hashDataFromDB.ImageContainer,
			&hashDataFromDB.NamePod, &hashDataFromDB.NameDeployment)
		if err != nil {
			ar.logger.Error("err while scan data from database ", err)
			return nil, err
		}
		allHashDataFromDB = append(allHashDataFromDB, &hashDataFromDB)
	}

	return allHashDataFromDB, nil
}

// DeleteFromTable removes data from the table that matches the name of the deployment
func (ar AppRepository) DeleteFromTable(nameDeployment string) error {
	query := `DELETE FROM hashfiles WHERE name_deployment=$1;`
	_, err := ar.db.Exec(query, nameDeployment)
	if err != nil {
		ar.logger.Error("err while deleting rows in database", err)
		return err
	}
	return nil
}
