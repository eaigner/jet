package jet

import (
	"database/sql"
)

func Open(driverName, dataSourceName string) (Db, error) {
	db2, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	return &db{db: db2}, nil
}

type db struct {
	db     *sql.DB
	runner *runner
}

func (d *db) Begin() (Tx, error) {
	tx2, err := d.db.Begin()
	if err != nil {
		return nil, err
	}
	return &tx{tx: tx2}, nil
}

func (d *db) Query(query string, args ...interface{}) Queryable {
	d.runner = &runner{
		qo:    d.db,
		query: query,
		args:  args,
	}
	return d
}

func (d *db) Run() error {
	return d.runner.Run()
}

func (d *db) Rows(v interface{}, maxRows ...int64) error {
	return d.runner.Rows(v, maxRows...)
}
