package jet

import (
	"database/sql"
)

type Migration interface {
	Up(tx Tx)
	Down(tx Tx)
}

type MigrationSuite interface {
	Add(m Migration)
	Run(jet Db) error
}

type Db interface {
	Queryable
	Begin() (Tx, error)
}

type Tx interface {
	Queryable
	Commit() error
}

type Queryable interface {
	Query(query string, args ...interface{}) Queryable
	Run() error
	Rows(v interface{}, maxRows ...int64) error
}

type queryObject interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}
