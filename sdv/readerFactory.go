package sdv

import (
	"bitbucket.org/timabell/sql-data-viewer/mssql"
	"bitbucket.org/timabell/sql-data-viewer/pg"
	"bitbucket.org/timabell/sql-data-viewer/sqlite"
)

func getDbReader(driver string, db string) dbReader {
	var reader dbReader
	switch driver {
	case "mssql":
		reader = mssql.NewMssql(db)
	case "pg":
		reader = pg.NewPg(db)
	case "sqlite":
		reader = sqlite.NewSqlite(db)
	case "":
		panic("Driver choice missing")
	default:
		panic("Unsupported driver choice " + driver)
	}
	return reader
}
