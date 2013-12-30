package jet

type jetQuery struct {
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

func (q *jetQuery) processedQna() (string, []interface{}) {
	query, args := substituteMapAndArrayMarks(q.query, q.args...)

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
	return query, enc
}

func (q *jetQuery) Run() error {
	query, args := q.processedQna()

	// prepare
	stmt, err := q.qo.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(args...)
	return err
}

func (q *jetQuery) Rows(v interface{}) error {
	if v == nil {
		return q.Run()
	}
	query, args := q.processedQna()

	// prepare
	stmt, err := q.qo.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// log
	if q.db.LogFunc != nil {
		q.db.LogFunc(q.id, query, args...)
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
