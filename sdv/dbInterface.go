package sdv

import (
	"database/sql"
	"sql-data-viewer/schema"
)

// This a kind of "active model",
// a thing that holds the currently known data about the connected
// database but that also knows how to get more information on-demand.

// todo rename to idiomatic DbReader
type dbInterface interface{
	GetTables() (tables []schema.TableName, err error)
	AllFks() (allFks schema.GlobalFkList, err error)
	GetRows(query schema.RowFilter, table schema.TableName, rowLimit int) (rows *sql.Rows, err error)
}
