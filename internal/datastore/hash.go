package datastore

import "database/sql"

type HashStorer struct {
	db *sql.DB
}

func New(db *sql.DB) HashStorer {
	return HashStorer{db: db}
}

func (h HashStorer) Get() {
}

func (h HashStorer) Create() {
}

func (h HashStorer) Delete() {
}
