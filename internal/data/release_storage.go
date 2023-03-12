package data

import (
	"database/sql"
	"time"

	"github.com/ScienceSoft-Inc/integrity-sum/pkg/k8s"
	"github.com/sirupsen/logrus"
)

//go:generate mockgen -source=release_storage.go -destination=mocks/mock_release_storage.go

type IReleaseStorage interface {
	PrepareQuery(deploymentData *k8s.DeploymentData) (string, []any)
	Get(deploymentData *k8s.DeploymentData) (*Release, error)
	Delete(nameDeployment string) error
	DeleteOldData() error
	IsExistDeploymentNameInDB(deploymentName string) bool
}

type Release struct {
	ID          int
	Name        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ReleaseType string
	Image       string
}

type ReleaseStorage struct {
	db     *sql.DB
	alg    string
	logger *logrus.Logger
}

// NewReleaseStorage creates a new ReleaseStorage structure to work with the database table
func NewReleaseStorage(db *sql.DB, alg string, logger *logrus.Logger) *ReleaseStorage {
	return &ReleaseStorage{
		db:     db,
		alg:    alg,
		logger: logger,
	}
}

// PrepareQuery creates a query and a set of arguments for preparing data for insertion into the database
func (rs ReleaseStorage) PrepareQuery(deploymentData *k8s.DeploymentData) (string, []any) {
	args := make([]any, 0)

	query := `INSERT INTO releases (name, created_at, updated_at, release_type, image) VALUES($1,$2,$3,$4,$5);`
	args = append(args,
		deploymentData.NameDeployment,
		time.Now(),
		time.Now(),
		"type",
		deploymentData.Image,
	)
	return query, args
}

// Get gets data from the database
func (rs ReleaseStorage) Get(deploymentData *k8s.DeploymentData) (*Release, error) {
	var hashData Release
	query := "SELECT id, name, created_at, updated_at, release_type, image FROM releases WHERE name=$1"

	row := rs.db.QueryRow(query, deploymentData.NameDeployment)
	err := row.Scan(&hashData.ID, &hashData.Name, &hashData.CreatedAt, &hashData.UpdatedAt, &hashData.ReleaseType, &hashData.Image)
	if err != nil {
		rs.logger.Error("err while scan row in database ", err)
		return nil, err
	}
	return &hashData, nil
}

// Delete removes the data with the release name from the database
func (rs ReleaseStorage) Delete(nameDeployment string) error {
	query := `DELETE FROM releases WHERE name=$1;`
	_, err := rs.db.Exec(query, nameDeployment)
	if err != nil {
		rs.logger.Error("err while deleting rows in database", err)
		return err
	}
	return nil
}

// DeleteOldData removes data when the threshold is exceeded from the database
func (rs ReleaseStorage) DeleteOldData() error {
	// query to delete old data
	query := "DELETE FROM releases WHERE created_at < (NOW() - INTERVAL 10 MINUTE)"
	_, err := rs.db.Exec(query)
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

// Update changes column updated_at with current timestamp
func (rs ReleaseStorage) Update(deploymentName string) error {
	rs.logger.Debug("update timestamp releases")
	query := `UPDATE  releases SET updated_at=NOW() WHERE name=$1;`
	_, err := rs.db.Exec(query, deploymentName)
	if err != nil {
		rs.logger.Error("error while updating timestamp to database", err)
		return err
	}
	return nil
}
