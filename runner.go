package jet

import (
	"fmt"
	"strings"
)

type runner struct {
	qo     queryObject
	txnId  string
	query  string
	args   []interface{}
	logger *Logger
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
	// Log
	r.logQuery()
	// Query
	rows, err := r.qo.Query(r.query, r.args...)
	if err != nil {
		return err
	}
	defer rows.Close()
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

func (r *runner) Value() (interface{}, error) {
	var m map[string]interface{}
	err := r.Rows(&m, 1)
	if err != nil {
		return nil, err
	}
	if x := len(m); x != 1 {
		return nil, fmt.Errorf("expected 1 column for Value(), got %d columns (%v)", x, m)
	}
	for _, v := range m {
		return v, nil
	}
	panic("unreachable")
}

func (r *runner) Logger() *Logger {
	return r.logger
}

func (r *runner) SetLogger(l *Logger) {
	r.logger = l
}

func (r *runner) logQuery() {
	if l := r.Logger(); l != nil {
		if r.txnId != "" {
			l.Txnf("\t%s: ", r.txnId[:6])
		}
		l.Queryf(r.query)
		args := []string{}
		for _, a := range r.args {
			args = append(args, fmt.Sprintf(`"%v"`, a))
		}
		if len(r.args) > 0 {
			l.Argsf(" [%s]", strings.Join(args, ", "))
		}
		l.Println()
	}
}
