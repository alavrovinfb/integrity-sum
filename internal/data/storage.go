package data

import (
	"database/sql"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/ScienceSoft-Inc/integrity-sum/internal/configs"
)

type Storage struct {
	db  *sql.DB
	log *logrus.Logger
}

var db *Storage
var openOnce sync.Once

// Open creates a connection to the database and returns a pointer to a Storage struct with an active database connection and an error if any.
func Open(log *logrus.Logger) (*Storage, error) {
	var err error
	openOnce.Do(func() {
		var conn *sql.DB
		conn, err = ConnectionToDB(log)
		if err != nil {
			return
		}
		db = &Storage{
			db:  conn,
			log: log,
		}
	})
	return db, err
}

// Close closes the connection to the database.
func Close() {
	db.db.Close()
}

// DB returns a pointer to a Storage struct with an active database connection
func DB() *Storage {
	return db
}

// SQL returns a database connection
func (db *Storage) SQL() *sql.DB {
	return db.db
}

// WithTx creates a transaction and executes a function with this transaction as an argument
func WithTx(f func(txn *sql.Tx) error) error {
	txn, err := DB().db.Begin()
	if err != nil {
		return err
	}

	if err = f(txn); err != nil {
		if errTx := txn.Rollback(); errTx != nil {
			return errTx
		}
		return err
	}

	if err = txn.Commit(); err != nil {
		return err
	}
	return nil
}

// ConnectionToDB connects to the database
func ConnectionToDB(logger *logrus.Logger) (*sql.DB, error) {
	logger.Info("Connecting to the database..")
	db, err := sql.Open("postgres", configs.GetDBConnString())
	if err != nil {
		logger.Error("Cannot connect to the database ")
		return nil, err
	}
	return db, nil
}

// CheckOldData removes data when the threshold is exceeded from the database
func CheckOldData(algName string, logger *logrus.Logger) {
	ticker := time.NewTicker(viper.GetDuration("db-ticker-interval"))
	go func() {
		for {
			<-ticker.C
			logger.Debug("### ✅ checking old data in db")
			err := NewReleaseData(DB().SQL()).DeleteOldData(viper.GetString("db-threshold-timeout"))
			if err != nil {
				logger.WithError(err).Error("failed check old data in DB")
			}
		}
	}()
}
