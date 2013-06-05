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
	db   *sql.DB
	conv ColumnConverter
}

func (d *db) SetMaxIdleConns(n int) {
	d.db.SetMaxIdleConns(n)
}

func (d *db) SetColumnConverter(conv ColumnConverter) {
	d.conv = conv
}

func (d *db) Begin() (Tx, error) {
	tx2, err := d.db.Begin()
	if err != nil {
		return nil, err
	}
	t := &tx{
		tx:   tx2,
		conv: d.conv,
		id:   newAlphanumericId(40),
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
		conv:   d.conv,
		query:  query,
		args:   args,
		logger: d.runner.logger,
	}
	return d
}
