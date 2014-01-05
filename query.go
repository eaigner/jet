package jet

import (
	"database/sql"
	"sync"
)

type jetQuery struct {
	m     sync.Mutex
	db    *Db
	qo    queryObject
	id    string
	query string
	args  []interface{}
}

// newQuery initiates a new query for the provided query object (either *sql.Tx or *sql.DB)
func newQuery(qo queryObject, db *Db, query string, args ...interface{}) *jetQuery {
	return &jetQuery{
		qo:    qo,
		db:    db,
		id:    newQueryId(),
		query: query,
		args:  args,
	}
}

func (q *jetQuery) Run() (err error) {
	return q.Rows(nil)
}

func (q *jetQuery) Rows(v interface{}) (err error) {
	q.m.Lock()
	defer q.m.Unlock()

	// disable lru in transactions
	useLru := true
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
	stmt, ok := q.db.lru.get(query)
	if !useLru || !ok {
		stmt, err = q.qo.Prepare(query)
		if err != nil {
			return err
		}
		if useLru {
			q.db.lru.put(query, stmt)
		}
	}

	// If no rows need to be unpacked use Exec
	if v == nil {
		_, err := stmt.Exec(args...)
		return err
	}

	// run query
	rows, err := stmt.Query(args...)
	if err != nil {
		return err
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
