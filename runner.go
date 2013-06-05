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
	qo     queryObject
	conv   ColumnConverter
	expand bool
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
	// Expand map and slice markers
	query, args := r.query, r.args
	if r.expand {
		query, args = substituteMapAndArrayMarks(r.query, r.args...)
	}
	r.logQuery(query, args)
	var (
		rows *sql.Rows
		err  error
	)
	if v == nil {
		_, err = r.qo.Exec(query, args...)
		return err

	} else {
		rows, err = r.qo.Query(query, args...)
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
		keys: cols,
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
			var cv interface{}
			cont = append(cont, &cv)
		}
		err := rows.Scan(cont...)
		if err != nil {
			return err
		}

		// Map values
		mppr.values = cont
		err = mppr.unpack(v)
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
	setValue(first, reflect.ValueOf(v).Elem())
	return nil
}

func (r *runner) Logger() *Logger {
	return r.logger
}

func (r *runner) SetLogger(l *Logger) {
	r.logger = l
}

func (r *runner) logQuery(rquery string, rargs []interface{}) {
	if l := r.Logger(); l != nil {
		if r.txnId != "" {
			l.Txnf("         %s: ", r.txnId[:7])
		}
		l.Queryf(rquery)
		args := []string{}
		for _, a := range rargs {
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
		if len(rargs) > 0 {
			l.Argsf(" [%s]", strings.Join(args, ", "))
		}
		l.Println()
	}
}
