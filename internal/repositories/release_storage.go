package repositories

import (
	"database/sql"
	"fmt"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/models"
	"github.com/sirupsen/logrus"
	"time"
)

type ReleaseStorageRepository struct {
	logger *logrus.Logger
	db     *sql.DB
}

func NewReleaseStorageRepository(logger *logrus.Logger, db *sql.DB) *ReleaseStorageRepository {
	return &ReleaseStorageRepository{
		logger: logger,
		db:     db,
	}
}

// IsExistDeploymentNameInDB checks if the base is empty
//func (rr ReleaseStorageRepository) IsExist(deploymentName string) (bool, error) {
//	var id int64
//	query := `SELECT id FROM releases WHERE name=$1 LIMIT 1;`
//	row := rr.db.QueryRow(query, deploymentName)
//	err := row.Scan(&id)
//	if err != nil {
//		rr.logger.Error("err while scan row in database ", err)
//		return false, err
//	}
//
//	return true, nil
//}

// Create iterates through all elements of the slice and triggers the save to database function
func (rr ReleaseStorageRepository) Create(deploymentData *models.DeploymentData) error {

	query1 := `INSERT INTO releases (name, created_at, image) VALUES($1,$2,$3);`
	_, err := rr.db.Exec(query1, deploymentData.NameDeployment, time.Now(), deploymentData.Image)
	if err != nil {
		rr.logger.Error("err while save data in database ", err)
		return err
	}
	return nil
}

// GetHashData retrieves data from the database using the path and algorithm
func (rr ReleaseStorageRepository) Get(deploymentData *models.DeploymentData) (*models.Release, error) {
	var allHashDataFromDB models.Release

	query := `SELECT id, name, created_at, image FROM releases WHERE name=$1`
	row := rr.db.QueryRow(query, deploymentData.NameDeployment)
	err := row.Scan(&allHashDataFromDB.ID, &allHashDataFromDB.Name, &allHashDataFromDB.CreatedAt, &allHashDataFromDB.Image)
	if err != nil {
		rr.logger.Error("err while scan row in database ", err)
		return nil, err
	}
	return &allHashDataFromDB, nil
}

// DeleteFromTable removes data from the table that matches the name of the deployment
func (rr ReleaseStorageRepository) Delete(nameDeployment string) error {
	query := `DELETE FROM releases WHERE name=$1;`
	_, err := rr.db.Exec(query, nameDeployment)
	if err != nil {
		rr.logger.Error("err while deleting rows in database", err)
		return err
	}
	return nil
}
func (rr ReleaseStorageRepository) DeleteOldData() error {
	// Set the threshold for deleting old data
	threshold := time.Now().AddDate(0, 0, -1) // one day
	//threshold2 := time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)
	//
	//year, month, dey := time.Now().Add(time.Minute * 10).Date()
	//loc := time.Now().Location()
	//needDate := time.Date(year, month, dey, 0, 10, 0, 0, loc)

	start := time.Now()
	fmt.Println("время текущее", start)
	//afterThreeMinutes := start.Add(time.Minute * 5)
	fmt.Println("настало время удалять ", threshold)

	// Delete old data
	_, err := rr.db.Exec("DELETE FROM releases WHERE created_at < $1", threshold)
	if err != nil {
		rr.logger.Error(err)
	}
	return nil
}
