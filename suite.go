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

// Migration represents a migration step with up and down SQL strings.
type Migration struct {
	Up   string
	Down string
	id   int
}

// Add adds a migration step to the suite.
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

	m.id = len(s.Migrations)
	s.Migrations = append(s.Migrations, m)
}

// AddSQL is a shorthand for Add taking strings for the up and down SQL.
func (s *Suite) AddSQL(up, down string) {
	s.Add(&Migration{
		Up:   up,
		Down: down,
	})
}

// Step applies 1 migration (upward).
func (s *Suite) Step(db *Db) (int, int, error) {
	return s.Run(db, true, 1)
}

// Rollback rolls back 1 step.
func (s *Suite) Rollback(db *Db) (int, int, error) {
	return s.Run(db, false, 1)
}

// Migrate applies all migrations.
func (s *Suite) Migrate(db *Db) (int, int, error) {
	return s.Run(db, true, math.MaxInt32)
}

// Reset rolls back all migrations
func (s *Suite) Reset(db *Db) (int, int, error) {
	return s.Run(db, false, math.MaxInt32)
}

// Run runs the entire migration suite. If up is false, it will run in reverse.
// It returns the current migration id, the number of steps applied in this run and an error (if one occurred).
func (s *Suite) Run(db *Db, up bool, maxSteps int) (int, int, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if l := len(s.Migrations); l == 0 {
		return -1, 0, errors.New("cannot run suite, no migrations set")
	}
	err := db.Query(s.Stmts.CreateTableSQL).Run()
	if err != nil {
		return -1, 0, err
	}
	var row struct {
		Version int
	}
	row.Version = -1
	err = db.Query(s.Stmts.SelectVersionSQL).Rows(&row)
	if err != nil {
		return -1, 0, err
	}
	step := 0
	stepsApplied := 0
	current := row.Version
	for _, m := range s.buildList(up, row.Version) {
		if step++; maxSteps > 0 && step > maxSteps {
			break
		}

		txn, err := db.DB.Begin()
		if err != nil {
			return current, stepsApplied, err
		}

		next := m.id
		if up {
			_, err = txn.Exec(m.Up)
		} else {
			next--
			_, err = txn.Exec(m.Down)
		}
		if err != nil {
			return current, stepsApplied, err
		}

		if current == -1 {
			_, err = txn.Exec(s.Stmts.InsertVersionSQL, next)
		} else {
			_, err = txn.Exec(s.Stmts.UpdateVersionSQL, next, current)
		}
		if err != nil {
			return current, stepsApplied, err
		}

		if err := txn.Commit(); err != nil {
			return current, stepsApplied, err
		}
		current = next
		stepsApplied++
	}
	return current, stepsApplied, nil
}

func (s *Suite) buildList(up bool, version int) []*Migration {
	a := []*Migration{}
	for i, m := range s.Migrations {
		m.id = i + 1
		if up {
			if m.id > version {
				a = append(a, m)
			}
		} else {
			if m.id <= version {
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
