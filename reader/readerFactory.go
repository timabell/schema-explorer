package reader

import (
	"bitbucket.org/timabell/sql-data-viewer/mssql"
	"bitbucket.org/timabell/sql-data-viewer/pg"
	"bitbucket.org/timabell/sql-data-viewer/sqlite"
)

func GetDbReader(driver string, options DbReaderOptions) DbReader {
	// should this handle the arg parsing?
	// pass in options interface, each driver has own concrete that it must be of that type
	// something converts args to options object
	var reader DbReader
	switch driver {
	case "mssql":
		reader = mssql.NewMssql("todo")
	case "pg":
		reader = pg.NewPg("todo")
	case "sqlite":
		reader = sqlite.NewSqlite("todo")
	case "":
		panic("Driver choice missing")
	default:
		panic("Unsupported driver choice " + driver)
	}
	return reader
}

func SupportedDrivers() (result []string) {
	result = []string{"mssql", "pg", "sqlite"}
	return
}
