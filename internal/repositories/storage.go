package repositories

import (
	"context"
	"database/sql"
	"sync"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"github.com/ScienceSoft-Inc/integrity-sum/internal/configs"
)

type Storage struct {
	db  *sql.DB
	log *logrus.Logger
}

var db *Storage
var openOnce sync.Once

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

func DB() *Storage {
	return db
}
func (db *Storage) SQL() *sql.DB {
	return db.db
}

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

func ExecQueryTx(ctx context.Context, sqlQuery string, args ...any) error {
	return WithTx(func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, sqlQuery, args...)
		return err
	})
}

func ConnectionToDB(logger *logrus.Logger) (*sql.DB, error) {
	logger.Info("Connecting to the database..")
	db, err := sql.Open("postgres", configs.GetDBConnString())
	if err != nil {
		logger.Error("Cannot connect to the database ")
		return nil, err
	}
	return db, nil
}
