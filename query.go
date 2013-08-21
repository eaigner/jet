package jet

import (
	"database/sql"
)

type query struct {
	db      *Db
	qo      queryObject
	id      string
	query   string
	args    []interface{}
	lru     *lruCache
	lastErr error
	stmt    *sql.Stmt
}

// newQuery initiates a new query for the provided query object (either *sql.Tx or *sql.DB)
func newQuery(qo queryObject, db *Db) *query {
	return &query{
		qo: qo,
		db: db,
		id: newAlphanumericId(7),
	}
}

func (q *query) Run() error {
	return q.Rows(nil)
}

func (q *query) Rows(v interface{}, maxRows ...int64) error {
	// Always clear the error after we are done with Rows.
	defer func() { q.lastErr = nil }()

	// Since Query doesn't return the error directly we do it here
	if q.lastErr != nil {
		return q.lastErr
	}

	// Determine max rows
	var max int64 = -1
	if len(maxRows) > 0 {
		max = maxRows[0]
	}
	var (
		rows *sql.Rows
		err  error
	)
	// if q.Logger() != nil {
	// 	r.logQuery()
	// }
	if v == nil {
		_, err = q.stmt.Exec(q.args...)
		return q.onErr(err)

	} else {
		rows, err = q.stmt.Query(q.args...)
		if err != nil {
			return q.onErr(err)
		}
		defer rows.Close()
	}
	cols, err := rows.Columns()
	if err != nil {
		return q.onErr(err)
	}
	var i int64 = 0
	mppr := &mapper{
		conv: q.db.ColumnConverter,
	}
	for {
		// Check if max rows has been reached
		if max >= 0 && i >= max {
			break
		}
		// Break if no more rows
		if !rows.Next() {
			break
		}
		// Scan values into containers
		cont := make([]interface{}, 0, len(cols))
		for i := 0; i < cap(cont); i++ {
			cont = append(cont, new(interface{}))
		}
		err := rows.Scan(cont...)
		if err != nil {
			return q.onErr(err)
		}

		// Map values
		err = mppr.unpack(cols, cont, v)
		if err != nil {
			return q.onErr(err)
		}
		i++
	}
	return nil
}

func (q *query) prepare(query string, args ...interface{}) Runnable {
	q2, a2 := substituteMapAndArrayMarks(query, args...)
	q.query = q2
	q.args = a2

	var stmt *sql.Stmt
	var lkey string
	if q.lru != nil {
		lkey = q.id + q.query
		stmt = q.lru.get(lkey)
	}
	if stmt == nil {
		var err error
		stmt, err = q.qo.Prepare(q.query)
		if err != nil {
			q.onErr(err)
			return q
		}
		if q.lru != nil {
			q.lru.set(lkey, stmt)
		}
	}
	q.stmt = stmt

	return q
}

func (q *query) onErr(err error) error {
	if err != nil {
		q.lastErr = err
		if q.lru != nil {
			q.lru.reset()
		}
	}
	return err
}
