package jet

import (
	"database/sql"
)

func Open(driverName, dataSourceName string) (Db, error) {
	godb, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	return &db{godb: godb}, nil
}

type db struct {
	godb *sql.DB
}

func (j *db) Begin() Tx {
	return &tx{godb: j.godb}
}

func (j *db) Exec(query string, args ...interface{}) error {
	return j.Query(nil, query, args...)
}

func (j *db) Query(v interface{}, query string, args ...interface{}) error {
	// Query
	rows, err := j.godb.Query(query, args...)
	if err != nil {
		return err
	}
	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	for rows.Next() {
		// Scan values into containers
		containers := make([]interface{}, 0, len(cols))
		for i := 0; i < cap(containers); i++ {
			var v interface{}
			containers = append(containers, &v)
		}
		err := rows.Scan(containers...)
		if err != nil {
			return err
		}

		// Map values
		m := make(map[string]interface{}, len(cols))
		for i, col := range cols {
			m[col] = containers[i]
		}
		err = mapper{m}.unpack(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (j *db) Count() int64 {
	panic("not implemented")
}
