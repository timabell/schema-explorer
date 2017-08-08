package sdv

import (
	"bitbucket.org/timabell/sql-data-viewer/sqlite"
	"bitbucket.org/timabell/sql-data-viewer/mssql"
)

func getDbReader(driver string, db string) dbReader {
	var reader dbReader
	switch driver {
	case "mssql":
		reader = mssql.NewMssql(db)
	case "sqlite":
		reader = sqlite.NewSqlite(db)
	case "":
		panic("Driver choice missing")
	default:
		panic("Unsupported driver choice " + driver)
	}
	return reader
}
