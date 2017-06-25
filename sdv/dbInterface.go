package sdv

import "database/sql"

// This a kind of "active model",
// a thing that holds the currently known data about the connected
// database but that also knows how to get more information on-demand.

// todo rename to idiomatic DbReader
type dbInterface interface{
	GetTables() (tables []TableName, err error)
	AllFks() (allFks GlobalFkList)
	GetRows(query RowFilter, table TableName, rowLimit int) (rows *sql.Rows, err error)
}
