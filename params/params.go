package params

import (
	"bitbucket.org/timabell/sql-data-viewer/schema"
	"fmt"
	"html/template"
	"net/url"
	"strconv"
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

// explicitly not using pointer in order to modify a copy

func (tableParams TableParams) CardViewOff() TableParams {
	tableParams.CardView = false
	return tableParams
}

func (tableParams TableParams) CardViewOn() TableParams {
	tableParams.CardView = true
	return tableParams
}

func (tableParams TableParams) AddSort(col *schema.Column) TableParams {
	// todo: don't add again. maybe handle in markup. Or maybe use that to toggle desc and move to front of list

	tableParams.Sort = append(tableParams.Sort, SortCol{Column: col})
	return tableParams
}

func (tableParams TableParams) ClearSort() TableParams {
	tableParams.Sort = nil
	return tableParams
}

func (tableParams TableParams) ClearFilter() TableParams {
	tableParams.Filter = nil
	return tableParams
}

func (tableParams TableParams) AsQueryString() template.URL {
	parts := BuildFilterParts(tableParams.Filter)
	// todo: sort param
	//if len(tableParams.Sort){
	//	parts = append(parts, fmt.Sprintf("%s=%s", sortKey, tableParams.Sort))
	//}
	for _, sortCol := range tableParams.Sort {
		parts = append(parts, fmt.Sprintf("%s=%s", sortKey, sortCol.Column))
	}
	if tableParams.CardView {
		parts = append(parts, fmt.Sprintf("%s=%s", cardViewKey, "true"))
	}
	if tableParams.RowLimit > 0 {
		parts = append(parts, fmt.Sprintf("%s=%d", rowLimitKey, tableParams.RowLimit))
	}
	return template.URL(strings.Join(parts, "&"))
}

func (filterList FieldFilterList) AsQueryString() template.URL {
	parts := BuildFilterParts(filterList)
	return template.URL(strings.Join(parts, "&"))
}

func BuildFilterParts(filterList FieldFilterList) []string {
	var parts []string
	for _, part := range filterList {
		// todo: support multiple values correctly
		parts = append(parts, fmt.Sprintf("%s=%s", part.Field, strings.Join(part.Values, ",")))
	}
	return parts
}

// todo: more robust separation of query param keys
const rowLimitKey = "_rowLimit" // this should be reasonably safe from clashes with column names
const cardViewKey = "_cardView"
const sortKey = "_sort"

func ParseTableParams(raw url.Values, table *schema.Table) (tableParams *TableParams) {
	tableParams = &TableParams{}
	ParseRowLimit(raw, tableParams)
	ParseSortParams(raw, tableParams, table)
	ParseCardView(raw, tableParams)

	// exclude special params from column filters
	raw.Del(rowLimitKey)
	raw.Del(sortKey)
	raw.Del(cardViewKey)

	ParseFilters(raw, tableParams, table)

	return
}

func ParseFilters(raw url.Values, tableParams *TableParams, table *schema.Table) {
	if len(raw) > 0 {
		for k, v := range raw {
			_, col := table.FindColumn(k)
			if col == nil {
				panic("Column '" + k + "' not found")
			}
			tableParams.Filter = append(tableParams.Filter, FieldFilter{Field: col, Values: v})
		}
	}
}

func ParseCardView(raw url.Values, tableParams *TableParams) {
	cardViewString := raw.Get(cardViewKey)
	if cardViewString != "" {
		tableParams.CardView = cardViewString == "true"
	}
}

func ParseRowLimit(raw url.Values, tableParams *TableParams) {
	rowLimitString := raw.Get(rowLimitKey)
	if rowLimitString == "" {
		return
	}
	var err error
	tableParams.RowLimit, err = strconv.Atoi(rowLimitString)
	if err != nil {
		fmt.Println("error converting rows querystring value to int: ", err)
		panic(err)
	}
}

func ParseSortParams(raw url.Values, tableParams *TableParams, table *schema.Table) {
	sortString := raw.Get(sortKey)
	tableParams.Sort = []SortCol{}
	if sortString == "" {
		return
	}
	var err error
	columnStrings := strings.Split(sortString, ",")
	for _, columnString := range columnStrings {
		const descStr = "~desc"
		var columnName string
		var colSort = SortCol{}
		if strings.HasSuffix(columnString, descStr) {
			colSort.Descending = true
			columnName = strings.TrimSuffix(columnString, descStr)
		} else {
			columnName = columnString
		}
		_, column := table.FindColumn(columnName)
		if column == nil {
			panic("column not found for sorting: " + columnString)
		}
		colSort.Column = column
		tableParams.Sort = append(tableParams.Sort, colSort)
	}
	if err != nil {
		fmt.Println("error parsing Sort order", err)
		panic(err)
	}
	return
}
