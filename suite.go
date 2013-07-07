package jet

import (
	"errors"
	"fmt"
	"math"
	"sync"
)

const (
	TableName  string = "migrations"
	ColumnName string = "version"
)

var (
	EOM error = errors.New("EOM")
)

type Suite struct {
	Migrations []*Migration
	Stmts      *Stmts
	mtx        sync.Mutex
}

type Stmts struct {
	CreateTableSQL   string
	SelectVersionSQL string
	InsertVersionSQL string
	UpdateVersionSQL string
}

func (s *Suite) Add(m *Migration) {
	if m == nil {
		panic("nil migration")
	}
	if s.Stmts == nil {
		s.Stmts = &Stmts{
			CreateTableSQL:   fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s" ( "%s" INTEGER )`, TableName, ColumnName),
			SelectVersionSQL: fmt.Sprintf(`SELECT "%s" FROM "%s" LIMIT 1`, ColumnName, TableName),
			InsertVersionSQL: fmt.Sprintf(`INSERT INTO "%s" ( "%s" ) VALUES ( $1 )`, TableName, ColumnName),
			UpdateVersionSQL: fmt.Sprintf(`UPDATE "%s" SET "%s" = $1 WHERE "%s" = $2`, TableName, ColumnName, ColumnName),
		}
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

func (s *Suite) Step(db Db) (error, int64, int) {
	return s.Run(db, true, 1)
}

func (s *Suite) Rollback(db Db) (error, int64, int) {
	return s.Run(db, false, 1)
}

func (s *Suite) Migrate(db Db) (error, int64, int) {
	return s.Run(db, true, math.MaxInt32)
}

func (s *Suite) Reset(db Db) (error, int64, int) {
	return s.Run(db, false, math.MaxInt32)
}

func (s *Suite) Run(db Db, up bool, maxSteps int) (error, int64, int) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if l := len(s.Migrations); l == 0 {
		return errors.New("cannot run suite, no migrations set"), -1, 0
	}
	err := db.Query(s.Stmts.CreateTableSQL).Run()
	if err != nil {
		return err, -1, 0
	}
	var row struct {
		Version int64
	}
	row.Version = -1
	err = db.Query(s.Stmts.SelectVersionSQL).Rows(&row, 1)
	if err != nil {
		return err, -1, 0
	}
	step := 0
	stepsApplied := 0
	current := row.Version
	for _, m := range s.buildList(up, row.Version) {
		if step++; maxSteps > 0 && step > maxSteps {
			break
		}
		txn, err := db.Begin()
		if err != nil {
			return err, current, stepsApplied
		}
		next := m.Id
		if up {
			err = txn.Query(m.Up).Run()
		} else {
			next--
			err = txn.Query(m.Down).Run()
		}
		if err != nil {
			return err, current, stepsApplied
		}
		if current == -1 {
			err = txn.Query(s.Stmts.InsertVersionSQL, next).Run()
		} else {
			err = txn.Query(s.Stmts.UpdateVersionSQL, next, current).Run()
		}
		if err != nil {
			return err, current, stepsApplied
		}
		if err := txn.Commit(); err != nil {
			return err, current, stepsApplied
		}
		current = next
		stepsApplied++
	}
	return nil, current, stepsApplied
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
