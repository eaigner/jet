package jet

import (
	"context"
	"database/sql"
	"sync"
)

type jetQuery struct {
	m                    sync.Mutex
	db                   *Db
	qo                   queryObject
	id                   string
	query                string
	args                 []interface{}
	ctx                  context.Context
	usePreparedStmtCache bool
	noPreparedStmts      bool
}

// newQuery initiates a new query for the provided query object (either *sql.Tx or *sql.DB)
func newQuery(ctx context.Context, usePreparedStmtCache bool, noPreparedStmts bool, qo queryObject, db *Db, query string, args ...interface{}) *jetQuery {
	return &jetQuery{
		qo:                   qo,
		db:                   db,
		id:                   newQueryId(),
		query:                query,
		args:                 args,
		ctx:                  ctx,
		usePreparedStmtCache: usePreparedStmtCache,
		noPreparedStmts:      noPreparedStmts,
	}
}

func (q *jetQuery) Run() (err error) {
	return q.Rows(nil)
}

func (q *jetQuery) Rows(v interface{}) (err error) {
	q.m.Lock()
	defer q.m.Unlock()

	if q.ctx == nil {
		q.ctx = context.Background()
	}

	// disable lru in transactions
	useLru := q.usePreparedStmtCache
	switch q.qo.(type) {
	case *sql.Tx:
		useLru = false
	}

	query, args := substituteMapAndArrayMarks(q.query, q.args...)

	// clear query from cache on error
	defer func() {
		if useLru && err != nil {
			q.db.lru.del(query)
		}
	}()

	// encode complex args
	enc := make([]interface{}, 0, len(args))
	for _, a := range args {
		v, ok := a.(ComplexValue)
		if ok {
			enc = append(enc, v.Encode())
		} else {
			enc = append(enc, a)
		}
	}
	args = enc

	// log
	if q.db.LogFunc != nil {
		q.db.LogFunc(q.id, query, args...)
	}

	// prepare statement
	var rows *sql.Rows
	var ok bool
	var stmt *sql.Stmt

	if q.noPreparedStmts {
		if v == nil {
			_, err := q.db.DB.ExecContext(q.ctx, query, args...)
			return err
		}

		rows, err = q.db.DB.QueryContext(q.ctx, query, args...)
	} else {
		if useLru {
			stmt, ok = q.db.lru.get(query)
		}
		if !ok {
			stmt, err = q.qo.Prepare(query)
			if err != nil {
				return err
			}
			if useLru {
				q.db.lru.put(query, stmt)
			} else {
				defer stmt.Close()
			}
		}
		// If no rows need to be unpacked use Exec
		if v == nil {
			_, err := stmt.ExecContext(q.ctx, args...)
			return err
		}

		// run query
		rows, err = stmt.QueryContext(q.ctx, args...)
		if err != nil {
			return err
		}
	}

	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	var i int64 = 0
	colMapper := &mapper{
		conv: q.db.ColumnConverter,
	}
	for {
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
			return err
		}

		// Map values
		err = colMapper.unpack(cols, cont, v)
		if err != nil {
			return err
		}
		i++
	}
	return nil
}
