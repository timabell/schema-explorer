package sdv

import (
	"database/sql"
	"sql-data-viewer/schema"
)

type dbReader interface {
	CheckConnection() (err error)
	GetTables() (tables []schema.Table, err error)
	AllFks() (allFks schema.GlobalFkList, err error)
	GetRows(query schema.RowFilter, table schema.Table, rowLimit int) (rows *sql.Rows, err error)
}
