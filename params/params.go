package params

import "bitbucket.org/timabell/sql-data-viewer/schema"

type TableParams struct {
	RowLimit int
	CardView bool
	Filter   schema.RowFilter
	Sort     schema.ColumnList
}
