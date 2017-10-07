package sdv

import (
	"bitbucket.org/timabell/sql-data-viewer/schema"
	"database/sql"
)

type dbReader interface {
	CheckConnection() (err error)
	GetTables() (tables []schema.Table, err error)
	AllFks() (allFks schema.GlobalFkList, err error)
	GetRows(query schema.RowFilter, table schema.Table, rowLimit int) (rows *sql.Rows, err error)
	GetColumns(table schema.Table) (cols []schema.Column, err error)
}
