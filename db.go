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
	godb  *sql.DB
	query string
	args  []interface{}
}

func (j *db) Begin() Tx {
	return &tx{godb: j.godb}
}

func (j *db) Query(query string, args ...interface{}) Db {
	j.query = query
	j.args = args
	return j
}

func (j *db) Run() error {
	return j.Rows(nil)
}

func (j *db) Rows(v interface{}, maxRows ...int64) error {
	// Determine max rows
	var max int64 = -1
	if len(maxRows) > 0 {
		max = maxRows[0]
	}
	// Query
	rows, err := j.godb.Query(j.query, j.args...)
	if err != nil {
		return err
	}
	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	var i int64 = 0
	for rows.Next() {
		// Check if max rows has been reached
		if max >= 0 && i >= max {
			break
		}
		// Scan values into containers
		containers := make([]interface{}, 0, len(cols))
		for i := 0; i < cap(containers); i++ {
			var cv interface{}
			containers = append(containers, &cv)
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
		i++
	}
	return nil
}

func (j *db) Count() int64 {
	panic("not implemented")
}
