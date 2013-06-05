package jet

import (
	"database/sql"
)

func Open(driverName, dataSourceName string) (Db, error) {
	db2, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	v := &db{}
	v.runner.db = db2
	v.runner.qo = db2
	return v, nil
}

type db struct {
	runner
}

func (d *db) SetMaxIdleConns(n int) {
	d.db.SetMaxIdleConns(n)
}

func (d *db) ExpandMapAndSliceMarker(f bool) {
	d.expand = f
}

func (d *db) SetColumnConverter(conv ColumnConverter) {
	d.conv = conv
}

func (d *db) Begin() (Tx, error) {
	tx2, err := d.db.Begin()
	if err != nil {
		return nil, err
	}
	t := &tx{}
	t.tx = tx2
	t.qo = tx2
	t.conv = d.conv
	t.logger = d.logger
	t.txnId = newAlphanumericId(40) // TODO(erik): possible performance bottleneck!

	if l := d.Logger(); l != nil {
		t.SetLogger(l)
		l.Txnf("BEGIN    %s", t.txnId).Println()
	}
	return t, nil
}

func (d *db) Query(query string, args ...interface{}) Queryable {
	d.query = query
	d.args = args
	r := &runner{}
	*r = d.runner
	return r
}
