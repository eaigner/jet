package jet

type Migration interface {
	Up(tx Tx)
	Down(tx Tx)
}

type MigrationSuite interface {
	Add(m Migration)
	Run(jet Db) error
}

type Db interface {
	Begin() Tx
	Exec(query string, args ...interface{}) error
	Query(v interface{}, query string, args ...interface{}) error
	Count() int64
}

type Tx interface {
	Commit() error
	Query(sql string, args ...interface{}) error
}
