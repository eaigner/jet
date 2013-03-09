package jet

import (
	"database/sql"
)

func Open(driverName, dataSourceName string) (Jet, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	return &jet{db: db}, nil
}

type jet struct {
	db *sql.DB
}

func (j *jet) Begin() Tx {
	return &tx{db: j.db}
}

func (j *jet) Query(v interface{}, query string, args ...interface{}) error {
	// Query
	rows, err := j.db.Query(query, args...)
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

func (j *jet) Count() int64 {
	return 0
}
