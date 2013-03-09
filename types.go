package jet

type Migration interface {
	Up()
	Down()
}

type MigrationSuite interface {
	Add(m Migration)
	Run() error
}

type Jet interface {
	Begin() Tx
	Query(v interface{}, query string, args ...interface{}) error
	Count() int64
}

type Tx interface {
	Commit() error
	Query(sql string, args ...interface{}) error
}
