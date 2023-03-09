package repositories

import (
	"database/sql"
	"fmt"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/models"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"time"
)

type ReleaseStorage struct {
	db     *sql.DB
	alg    string
	logger *logrus.Logger
}

// NewHashStorage creates a new struct HashService
func NewReleaseStorage(db *sql.DB, alg string, logger *logrus.Logger) *ReleaseStorage {
	return &ReleaseStorage{
		db:     db,
		alg:    alg,
		logger: logger,
	}
}

// Create accesses the repository to save data to the database
func (rs ReleaseStorage) Create(deploymentData *models.DeploymentData) error {
	query := `INSERT INTO releases (name, created_at, image) VALUES($1,$2,$3);`
	_, err := rs.db.Exec(query, deploymentData.NameDeployment, time.Now(), deploymentData.Image)
	if err != nil {
		rs.logger.Error("error while creating data to database", err)
		return err
	}
	fmt.Println("create releases: ", deploymentData.NameDeployment, time.Now(), deploymentData.Image)
	return nil
}

// Get accesses the repository to get data from the database
func (rs ReleaseStorage) Get(deploymentData *models.DeploymentData) (*models.Release, error) {
	var hashData models.Release
	query := "SELECT id, name, created_at, image FROM releases WHERE name=$1"

	row := rs.db.QueryRow(query, deploymentData.NameDeployment)
	err := row.Scan(&hashData.ID, &hashData.Name, &hashData.CreatedAt, &hashData.Image)
	if err != nil {
		rs.logger.Error("err while scan row in database ", err)
		return nil, err
	}
	return &hashData, nil

}

func (rs ReleaseStorage) Delete(nameDeployment string) error {
	query := `DELETE FROM releases WHERE name=$1;`
	_, err := rs.db.Exec(query, nameDeployment)
	if err != nil {
		rs.logger.Error("err while deleting rows in database", err)
		return err
	}
	return nil
}

func (rs ReleaseStorage) DeleteOldData() error {
	// query to delete old data
	threshold := viper.GetString("threshold")
	query := "DELETE FROM releases WHERE created_at < (NOW() - INTERVAL 10 MINUTE)"
	_, err := rs.db.Exec(query, threshold)
	if err != nil {
		rs.logger.Error("err while deleting rows in database", err)
		return err
	}
	return nil
}

// IsExistDeploymentNameInDB checks if the database is empty
func (rs ReleaseStorage) IsExistDeploymentNameInDB(deploymentName string) bool {
	var count int
	query := `SELECT COUNT(*) FROM releases WHERE name=$1;`
	err := rs.db.QueryRow(query, deploymentName).Scan(&count)
	if err != nil {
		rs.logger.Fatalf("err while scan row in database %s", err)
		return false
	}
	if count == 0 {
		rs.logger.Info("no rows in database")
		return false
	}
	return true
}
