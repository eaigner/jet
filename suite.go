package jet

import (
	"errors"
	"fmt"
	"math"
)

const (
	TableName  string = "migrations"
	ColumnName string = "version"
)

var (
	EOM error = errors.New("EOM")
)

type Suite struct {
	Migrations       []*Migration
	CreateTableSQL   string
	SelectVersionSQL string
	InsertVersionSQL string
	UpdateVersionSQL string
}

func NewSuite() *Suite {
	return &Suite{
		CreateTableSQL:   fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s" ( "%s" INTEGER )`, TableName, ColumnName),
		SelectVersionSQL: fmt.Sprintf(`SELECT "%s" FROM "%s" LIMIT 1`, ColumnName, TableName),
		InsertVersionSQL: fmt.Sprintf(`INSERT INTO "%s" ( "%s" ) VALUES ( $1 )`, TableName, ColumnName),
		UpdateVersionSQL: fmt.Sprintf(`UPDATE "%s" SET "%s" = $1 WHERE "%s" = $2`, TableName, ColumnName, ColumnName),
	}
}

func (s *Suite) Add(m *Migration) {
	if m == nil {
		panic("nil migration")
	}
	if s.Migrations == nil {
		s.Migrations = []*Migration{m}
	} else {
		s.Migrations = append(s.Migrations, m)
	}
}

func (s *Suite) AddSQL(up, down string) {
	s.Add(&Migration{
		Up:   up,
		Down: down,
	})
}

func (s *Suite) Step(db Db) (error, int64) {
	return s.Run(db, true, 1)
}

func (s *Suite) Rollback(db Db) (error, int64) {
	return s.Run(db, false, 1)
}

func (s *Suite) Migrate(db Db) (error, int64) {
	return s.Run(db, true, math.MaxInt32)
}

func (s *Suite) Reset(db Db) (error, int64) {
	return s.Run(db, false, math.MaxInt32)
}

func (s *Suite) Run(db Db, up bool, maxSteps int) (error, int64) {
	if l := len(s.Migrations); l == 0 {
		return errors.New("cannot run suite, no migrations set"), -1
	}
	err := db.Query(s.CreateTableSQL).Run()
	if err != nil {
		return err, -1
	}
	var row struct {
		Version int64
	}
	err = db.Query(s.SelectVersionSQL).Rows(&row, 1)
	if err != nil {
		return err, -1
	}
	current := row.Version
	for _, m := range s.buildList(up, row.Version) {
		txn, err := db.Begin()
		if err != nil {
			return err, -1
		}
		next := m.Id
		if up {
			txn.Query(m.Up).Run()
		} else {
			next--
			txn.Query(m.Down).Run()
		}
		if current == 0 {
			txn.Query(s.InsertVersionSQL, next).Run()
		} else {
			txn.Query(s.UpdateVersionSQL, next, current).Run()
		}
		if err := txn.Commit(); err != nil {
			return err, -1
		}
		current = next
	}
	return nil, current
}

func (s *Suite) buildList(up bool, version int64) []*Migration {
	a := []*Migration{}
	for i, m := range s.Migrations {
		m.Id = int64(i + 1)
		if up {
			if m.Id > version {
				a = append(a, m)
			}
		} else {
			if m.Id <= version {
				a = append(a, m)
			}
		}
	}
	if !up {
		reverse(a)
	}
	return a
}

func reverse(a []*Migration) []*Migration {
	for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
		a[i], a[j] = a[j], a[i]
	}
	return a
}
