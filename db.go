package jet

import (
	"database/sql"
)

type Db struct {
	*sql.DB

	ColumnConverter ColumnConverter
	LRUCache        *LRUCache

	driver string
	source string
}

// Open opens a new database connection.
func Open(driverName, dataSourceName string) (*Db, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	j := &Db{
		ColumnConverter: SnakeCaseConverter, // default
		driver:          driverName,
		source:          dataSourceName,
	}
	j.DB = db

	return j, nil
}

// Begin starts a new transaction
func (db *Db) Begin() (*Tx, error) {
	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}
	return &Tx{
		db: db,
		tx: tx,
	}, nil
}

// Query creates a prepared query that can be run with Rows or Run.
func (db *Db) Query(query string, args ...interface{}) Runnable {
	return newQuery(db, db).prepare(query, args...)
}
