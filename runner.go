package jet

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

type runner struct {
	db     *sql.DB
	tx     *sql.Tx
	stmt   *sql.Stmt
	qo     queryObject
	conv   ColumnConverter
	expand bool
	txnId  string
	query  string
	args   []interface{}
	errors []error
	logger *Logger
}

func (r *runner) copy() *runner {
	c := new(runner)
	*c = *r
	return c
}

func (r *runner) prepare(query string) Queryable {
	if r.query == query && r.stmt != nil {
		return r
	}
	var err error
	r.stmt, err = r.qo.Prepare(query)
	if err != nil {
		panic(err)
	}
	r.query = query
	return r
}

func (r *runner) Query(query string, args ...interface{}) Queryable {
	if r.expand {
		query, args = substituteMapAndArrayMarks(query, args...)
	}
	r.args = args
	return r.prepare(query)
}

func (r *runner) Run() error {
	return r.Rows(nil)
}

func (r *runner) Rows(v interface{}, maxRows ...int64) error {
	if len(r.errors) > 0 {
		return r.errors[0]
	}
	// Determine max rows
	var max int64 = -1
	if len(maxRows) > 0 {
		max = maxRows[0]
	}
	if r.Logger() != nil {
		r.logQuery()
	}
	var (
		rows *sql.Rows
		err  error
	)
	if v == nil {
		_, err = r.stmt.Exec(r.args...)
		return err

	} else {
		rows, err = r.stmt.Query(r.args...)
		if err != nil {
			return err
		}
		defer rows.Close()
	}
	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	var i int64 = 0
	mppr := &mapper{
		conv: r.conv,
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
			return err
		}

		// Map values
		err = mppr.unpack(cols, cont, v)
		if err != nil {
			return err
		}
		i++
	}
	return nil
}

func (r *runner) Value(v interface{}) error {
	var m map[string]interface{}
	err := r.Rows(&m, 1)
	if err != nil {
		return err
	}
	if x := len(m); x != 1 {
		return fmt.Errorf("expected 1 column for Value(), got %d columns (%v)", x, m)
	}
	var first interface{}
	for _, v := range m {
		first = v
		break
	}
	setValue(reflect.ValueOf(first), reflect.ValueOf(v).Elem())
	return nil
}

func (r *runner) Logger() *Logger {
	return r.logger
}

func (r *runner) SetLogger(l *Logger) {
	r.logger = l
}

func (r *runner) logQuery() {
	l := r.Logger()
	if r.txnId != "" {
		l.Txnf("         %s: ", r.txnId[:7])
	}
	l.Queryf(r.query)
	args := []string{}
	for _, a := range r.args {
		var buf []byte
		switch t := a.(type) {
		case []uint8:
			buf = t
			if len(buf) > 5 {
				buf = buf[:5]
			}
		}
		if buf != nil {
			args = append(args, fmt.Sprintf(`<buf:%x...>`, buf))
		} else {
			args = append(args, fmt.Sprintf(`"%v"`, a))
		}

	}
	if len(r.args) > 0 {
		l.Argsf(" [%s]", strings.Join(args, ", "))
	}
	l.Println()
}
