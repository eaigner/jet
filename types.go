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
	Query(query string, args ...interface{}) Db
	Run() error
	Rows(v interface{}, maxRows ...int64) error
	Count() int64
}

type Tx interface {
	Commit() error
	Query(sql string, args ...interface{}) error
}
