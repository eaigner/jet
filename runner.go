package jet

type runner struct {
	qo    queryObject
	query string
	args  []interface{}
}

func (r *runner) Query(query string, args ...interface{}) Queryable {
	r.query = query
	r.args = args
	return r
}

func (r *runner) Run() error {
	return r.Rows(nil)
}

func (r *runner) Rows(v interface{}, maxRows ...int64) error {
	// Determine max rows
	var max int64 = -1
	if len(maxRows) > 0 {
		max = maxRows[0]
	}
	// Query
	rows, err := r.qo.Query(r.query, r.args...)
	if err != nil {
		return err
	}
	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	var i int64 = 0
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
