package params

import (
	"bitbucket.org/timabell/sql-data-viewer/schema"
	"fmt"
	"html/template"
	"strings"
)

type SortCol struct {
	Column     *schema.Column
	Descending bool
}

type TableParams struct {
	RowLimit int
	CardView bool
	Filter   FieldFilterList
	Sort     []SortCol
}

type FieldFilter struct {
	Field  *schema.Column
	Values []string
}

type FieldFilterList []FieldFilter

func (filterList FieldFilterList) AsQueryString() template.URL {
	var parts []string
	for _, part := range filterList {
		// todo: support multiple values correctly
		parts = append(parts, fmt.Sprintf("%s=%s", part.Field, strings.Join(part.Values, ",")))
	}
	return template.URL(strings.Join(parts, "&"))
}
