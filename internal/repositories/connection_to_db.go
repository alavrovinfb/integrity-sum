package repositories

import (
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"github.com/integrity-sum/internal/configs"
)

func ConnectionToDB(logger *logrus.Logger) (*sql.DB, error) {
	logger.Info("Connecting to the database..")
	db, err := sql.Open("postgres", configs.GetDBConnString())
	if err != nil {
		logger.Error("Cannot connect to the database ")
		return nil, err
	}
	return db, nil
}
