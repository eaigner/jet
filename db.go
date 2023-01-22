package jet

import (
	"context"
	"database/sql"
	"strings"
)

// LogFunc can be set on the Db instance to allow query logging.
type LogFunc func(queryId, query string, args ...interface{})

type Db struct {
	*sql.DB

	// LogFunc is the log function to use for query logging.
	// Defaults to nil.
	LogFunc LogFunc

	// The column converter to use.
	// Defaults to SnakeCaseConverter.
	ColumnConverter ColumnConverter

	driver            string
	source            string
	lru               *lru
	skipPreparedStmts bool
}

// Open opens a new database connection.
func Open(driverName, dataSourceName string, preparedStmtCacheSize int) (*Db, error) {
	return OpenFunc(driverName, dataSourceName, sql.Open, preparedStmtCacheSize)
}

// OpenFunc opens a new database connection by using the passed `fn`.
func OpenFunc(driverName, dataSourceName string, fn func(string, string) (*sql.DB, error), preparedStmtCacheSize int) (*Db, error) {
	db, err := fn(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	j := &Db{
		ColumnConverter:   SnakeCaseConverter, // default
		driver:            driverName,
		source:            dataSourceName,
		lru:               newLru(preparedStmtCacheSize),
		skipPreparedStmts: strings.Contains(dataSourceName, "interpolateParams=true"),
	}
	j.DB = db

	return j, nil
}

// SetMaxCachedStatements sets the max number of statements to cache in the LRU. The default is 500.
func (db *Db) SetMaxCachedStatements(n int) {
	db.lru.maxItems = n
}

// Begin starts a new transaction
func (db *Db) Begin() (*Tx, error) {
	qid := newQueryId()
	if db.LogFunc != nil {
		db.LogFunc(qid, "BEGIN")
	}
	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}
	return &Tx{
		db:  db,
		tx:  tx,
		qid: qid,
	}, nil
}

// Query creates a prepared query that can be run with Rows or Run.
func (db *Db) Query(query string, args ...interface{}) Runnable {
	return db.QueryContext(context.Background(), query, args...)
}

// QueryContext creates a prepared query that can be run with Rows or Run.
func (db *Db) QueryContext(ctx context.Context, query string, args ...interface{}) Runnable {
	return newQuery(ctx, db.skipPreparedStmts, db, db, query, args...)
}
