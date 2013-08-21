package jet

import (
	"database/sql"
)

type Db struct {
	*sql.DB
	ColumnConverter ColumnConverter

	driver string
	source string

	//--
	tx      *sql.Tx
	stmt    *sql.Stmt
	qo      queryObject
	conv    ColumnConverter
	txnId   string
	query   string
	lastErr error
	args    []interface{}
	lru     *lruCache
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

// func (db *Db) logQuery() {
// 	l := r.Logger()
// 	if r.txnId != "" {
// 		l.Txnf("         %s: ", r.txnId[:7])
// 	}
// 	l.Queryf(r.query)
// 	args := []string{}
// 	for _, a := range r.args {
// 		var buf []byte
// 		switch t := a.(type) {
// 		case []uint8:
// 			buf = t
// 			if len(buf) > 5 {
// 				buf = buf[:5]
// 			}
// 		}
// 		if buf != nil {
// 			args = append(args, fmt.Sprintf(`<buf:%x...>`, buf))
// 		} else {
// 			args = append(args, fmt.Sprintf(`"%v"`, a))
// 		}

// 	}
// 	if len(r.args) > 0 {
// 		l.Argsf(" [%s]", strings.Join(args, ", "))
// 	}
// 	l.Println()
// }
