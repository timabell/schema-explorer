package sdv

import (
	"bitbucket.org/timabell/sql-data-viewer/sqlite"
	"bitbucket.org/timabell/sql-data-viewer/mssql"
)

func getDbReader() dbReader {
	var reader dbReader
	switch driver {
	case "mssql":
		reader = mssql.NewMssql(db)
	case "sqlite":
		reader = sqlite.NewSqlite(db)
	}
	return reader
}
