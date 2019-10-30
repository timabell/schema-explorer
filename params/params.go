package params

import (
	"github.com/timabell/schema-explorer/schema"
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
	SkipRows int
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

// for building sort links
func (tableParams TableParams) AddSort(col *schema.Column) TableParams {
	var newSort []SortCol

	newSortCol := SortCol{Column: col}
	exists := false
	for _, sortCol := range tableParams.Sort {
		if sortCol.Column == col {
			exists = true
			sortCol.Descending = !sortCol.Descending
		}
		newSort = append(newSort, sortCol)
	}
	if !exists {
		newSort = append(newSort, newSortCol)
	}

	// return a copy of the tableparams with a new sort
	tableParams.Sort = newSort
	return tableParams
}

func (tableParams TableParams) SortPosition(col *schema.Column) int {
	for index, c := range tableParams.Sort {
		if c.Column.Name == col.Name {
			return index + 1
		}
	}
	return -1
}

func (tableParams TableParams) PrevPage() TableParams {
	skip := tableParams.SkipRows - tableParams.RowLimit
	if skip < 0 {
		skip = 0
	}
	tableParams.SkipRows = tableParams.SkipRows - tableParams.RowLimit
	return tableParams
}

func (tableParams TableParams) NextPage() TableParams {
	tableParams.SkipRows = tableParams.SkipRows + tableParams.RowLimit
	return tableParams
}

func (tableParams TableParams) IsSortedAsc(col *schema.Column) bool {
	for _, c := range tableParams.Sort {
		if c.Column.Name == col.Name && !c.Descending {
			return true
		}
	}
	return false
}

func (tableParams TableParams) IsSortedDesc(col *schema.Column) bool {
	for _, c := range tableParams.Sort {
		if c.Column.Name == col.Name && c.Descending {
			return true
		}
	}
	return false
}

func (tableParams TableParams) IsSorted(col *schema.Column) bool {
	for _, c := range tableParams.Sort {
		if c.Column.Name == col.Name {
			return true
		}
	}
	return false
}

func (tableParams TableParams) ClearSort() TableParams {
	tableParams.Sort = nil
	return tableParams
}

func (tableParams TableParams) ClearFilter() TableParams {
	tableParams.Filter = nil
	return tableParams
}

func (tableParams TableParams) ClearPaging() TableParams {
	tableParams.RowLimit = 0
	tableParams.SkipRows = 0
	return tableParams
}

func (tableParams TableParams) AsQueryString() template.URL {
	parts := BuildFilterParts(tableParams.Filter)

	sortParts := BuildSortParts(tableParams)
	parts = append(parts, sortParts...)

	if tableParams.CardView {
		parts = append(parts, fmt.Sprintf("%s=%s", cardViewKey, "true"))
	}

	if tableParams.RowLimit > 0 {
		parts = append(parts, fmt.Sprintf("%s=%d", rowLimitKey, tableParams.RowLimit))
	}

	if tableParams.SkipRows > 0 {
		parts = append(parts, fmt.Sprintf("%s=%d", skipKey, tableParams.SkipRows))
	}

	return template.URL(strings.Join(parts, "&"))
}

// Get the 1-based start row number for display in templates
func (tableParams TableParams) FromRow() int {
	return tableParams.SkipRows + 1
}

// Get the highest row number for display in templates
func (tableParams TableParams) ToRow() int {
	return tableParams.SkipRows + tableParams.RowLimit
}

func BuildSortParts(tableParams TableParams) []string {
	var sort []string
	for _, sortCol := range tableParams.Sort {
		if sortCol.Descending {
			sort = append(sort, sortCol.Column.Name+descStr)
		} else {
			sort = append(sort, sortCol.Column.Name)
		}
	}
	var parts []string
	if len(sort) > 0 {
		parts = append(parts, fmt.Sprintf("%s=%s", sortKey, strings.Join(sort, ",")))
	}
	return parts
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
const skipKey = "_skip"
const cardViewKey = "_cardView"
const sortKey = "_sort"

func ParseTableParams(raw url.Values, table *schema.Table) (tableParams *TableParams) {
	tableParams = &TableParams{}
	ParseRowLimit(raw, tableParams)
	ParseSkip(raw, tableParams)
	ParseSortParams(raw, tableParams, table)
	ParseCardView(raw, tableParams)

	// exclude special params from column filters
	raw.Del(rowLimitKey)
	raw.Del(skipKey)
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

func ParseSkip(raw url.Values, tableParams *TableParams) {
	skipString := raw.Get(skipKey)
	if skipString == "" {
		return
	}
	var err error
	tableParams.SkipRows, err = strconv.Atoi(skipString)
	if err != nil {
		fmt.Println("error converting skip querystring value to int: ", err)
		panic(err)
	}
}

const descStr = "~desc"

func ParseSortParams(raw url.Values, tableParams *TableParams, table *schema.Table) {
	sortString := raw.Get(sortKey)
	tableParams.Sort = []SortCol{}
	if sortString == "" {
		return
	}
	var err error
	columnStrings := strings.Split(sortString, ",")
	for _, columnString := range columnStrings {
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
