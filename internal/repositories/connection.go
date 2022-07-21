package repositories

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/integrity-sum/internal/core/models"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

func ConnectionToDB(logger *logrus.Logger) (*sql.DB, error) {
	connectionDB := validateDBConnectionValues(logger)

	DBURL := fmt.Sprintf("host=%v port=%s user=%s dbname=%s sslmode=disable password=%s", connectionDB.DbHost, connectionDB.DbPort, connectionDB.DbUser, connectionDB.DbName, connectionDB.DbPassword)

	db, err := sql.Open(connectionDB.Dbdriver, DBURL)
	if err != nil {
		logger.Info("Cannot connect to database ", connectionDB.Dbdriver)
		return db, err
	} else {
		logger.Info("Connected to the database ", connectionDB.Dbdriver)
	}

	return db, nil
}

func validateDBConnectionValues(logger *logrus.Logger) *models.ConnectionDB {
	DbDriver, ok := os.LookupEnv("DB_DRIVER")
	if !ok {
		DbDriver = "postgres"
		logger.Info("DB_DRIVER was not set. setting by default: %s", DbDriver)
	}

	DbHost, ok := os.LookupEnv("DB_HOST")
	if !ok {
		DbHost = "localhost"
		logger.Info("DB_HOST was not set. setting by default: %s", DbHost)
	}

	DbPassword, ok := os.LookupEnv("DB_PASSWORD")
	if !ok {
		logger.Info("DB_PASSWORD was not set. setting by default: %s", DbPassword)
	}

	DbUser, ok := os.LookupEnv("DB_USER")
	if !ok {
		logger.Info("DB_USER was not set. setting by default: %s", DbUser)
	}

	DbPort, ok := os.LookupEnv("DB_PORT")
	if !ok {
		DbPort = "5432"
		logger.Info("DB_PORT was not set. setting by default: %s", DbPort)
	}

	DbName, ok := os.LookupEnv("DB_NAME")
	if !ok {
		DbName = "pastgres-db"
		logger.Info("DB_NAME was not set. setting by default: %s", DbName)
	}

	connectionDB := &models.ConnectionDB{
		Dbdriver:   DbDriver,
		DbUser:     DbUser,
		DbPassword: DbPassword,
		DbPort:     DbPort,
		DbHost:     DbHost,
		DbName:     DbName,
	}
	return connectionDB
}
