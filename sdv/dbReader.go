package sdv

import (
	"database/sql"
	"sql-data-viewer/schema"
)

type dbReader interface{
	GetTables() (tables []schema.TableName, err error)
	AllFks() (allFks schema.GlobalFkList, err error)
	GetRows(query schema.RowFilter, table schema.TableName, rowLimit int) (rows *sql.Rows, err error)
	Columns(table schema.TableName) (columns []string, err error)
}
