package params

import "bitbucket.org/timabell/sql-data-viewer/schema"

type SortCol struct {
	Column     *schema.Column
	Descending bool
}

type TableParams struct {
	RowLimit int
	CardView bool
	Filter   schema.RowFilter
	Sort     []SortCol
}
