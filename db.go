package jet

import (
	"context"
	"github.com/jmoiron/sqlx"
)

type Db struct {
	*sqlx.DB

	// LogFunc is the log function to use for query logging.
	// Defaults to nil.
	LogFunc LogFunc

	// The column converter to use.
	// Defaults to SnakeCaseConverter.
	ColumnConverter ColumnConverter

	lru                  *lru
	usePreparedStmtCache bool
	noPreparedStmts      bool
}

func Wrap(datasource *sqlx.DB, preparedStmtCacheSize int, noPreparedStmts bool) (*Db, error) {
	j := &Db{
		ColumnConverter:      SnakeCaseConverter, // default
		DB:                   datasource,
		lru:                  newLru(preparedStmtCacheSize),
		usePreparedStmtCache: preparedStmtCacheSize > 0,
		noPreparedStmts:      noPreparedStmts,
		LogFunc:              NoopLogFunc,
	}

	return j, nil
}

// Open opens a new database connection.
func Open(driverName, dataSourceName string, preparedStmtCacheSize int, noPreparedStmts bool) (*Db, error) {
	return OpenFunc(driverName, dataSourceName, sqlx.Open, preparedStmtCacheSize, noPreparedStmts)
}

// OpenFunc opens a new database connection by using the passed `fn`.
func OpenFunc(driverName, dataSourceName string, fn func(string, string) (*sqlx.DB, error), preparedStmtCacheSize int, noPreparedStmts bool) (*Db, error) {
	db, err := fn(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	return Wrap(db, preparedStmtCacheSize, noPreparedStmts)
}

// SetMaxCachedStatements sets the max number of statements to cache in the LRU. The default is 500.
func (db *Db) SetMaxCachedStatements(n int) {
	db.lru.maxItems = n
}

// Begin starts a new transaction
func (db *Db) Begin() (*Tx, error) {
	qid := newQueryId()
	db.LogFunc(context.Background(), qid, "BEGIN")

	tx, err := db.DB.Beginx()
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
	return newQuery(ctx, db.DB, db, query, args...)
}

func (db *Db) CacheSize() int {
	return db.lru.size()
}

func (db *Db) UsePreparedStmtsCache() bool {
	return db.usePreparedStmtCache
}

func (db *Db) NoPreparedStmts() bool {
	return db.noPreparedStmts
}
