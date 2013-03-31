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
	runner
	db *sql.DB
}

func (d *db) Begin() (Tx, error) {
	tx2, err := d.db.Begin()
	if err != nil {
		return nil, err
	}
	t := &tx{
		tx: tx2,
		id: newAlphanumericId(40),
	}
	if l := d.Logger(); l != nil {
		t.SetLogger(l)
		l.Txnf("BEGIN    %s", t.id).Println()
	}
	return t, nil
}

func (d *db) Query(query string, args ...interface{}) Queryable {
	d.runner = runner{
		qo:     d.db,
		query:  query,
		args:   args,
		logger: d.runner.logger,
	}
	return d
}
