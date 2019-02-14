package schema

import (
	"fmt"
	"strings"
)

type SupportedFeatures struct {
	Schema               bool
	Descriptions         bool
	FkNames              bool
	PagingWithoutSorting bool
}

type Database struct {
	Tables            []*Table
	Fks               []*Fk
	Indexes           []*Index
	Supports          SupportedFeatures
	Description       string
	DefaultSchemaName string
}

type Pk struct {
	Name    string
	Columns ColumnList
}

type Index struct {
	Name        string
	Columns     ColumnList
	IsUnique    bool
	IsClustered bool
	IsDisabled  bool
	Table       *Table
}

func (index Index) String() string {
	unique := ""
	if index.IsUnique {
		unique = "Unique "
	}
	return fmt.Sprintf("%sIndex %s on %s(%s)", unique, index.Name, index.Table.String(), index.Columns.String())
}

type Table struct {
	Schema      string
	Name        string
	Columns     ColumnList
	Pk          *Pk
	Fks         []*Fk
	InboundFks  []*Fk
	Indexes     []*Index
	Description string
	RowCount    *int       // pointer to allow us to tell the difference between zero and unknown
	PeekColumns ColumnList // list of columns to show as a preview when this is a target for a join, e.g. the "Name" column. The schema readers are not expected to populate this field.
}

type TableList []*Table

// implement sort.Interface for list of tables https://stackoverflow.com/a/19948360/10245
func (tables TableList) Len() int {
	return len(tables)
}
func (tables TableList) Swap(i, j int) {
	tables[i], tables[j] = tables[j], tables[i]
}
func (tables TableList) Less(i, j int) bool {
	return tables[i].String() < tables[j].String()
}

type ColumnList []*Column

type Column struct {
	Position       int
	Name           string
	Type           string
	Fks            []*Fk
	InboundFks     []*Fk
	Indexes        []*Index
	Description    string
	IsInPrimaryKey bool
	Nullable       bool
}

type Fk struct {
	Id                 int
	Name               string
	SourceTable        *Table
	SourceColumns      ColumnList
	DestinationTable   *Table
	DestinationColumns ColumnList
}

// Simplified fk constructor for single-column foreign keys
func NewFk(name string, sourceTable *Table, sourceColumn *Column, destinationTable *Table, destinationColumn *Column) *Fk {
	return &Fk{Name: name, SourceTable: sourceTable, SourceColumns: ColumnList{sourceColumn}, DestinationTable: destinationTable, DestinationColumns: ColumnList{destinationColumn}}
}

func (table Table) String() string {
	if table.Schema == "" {
		return table.Name
	}
	return table.Schema + "." + table.Name
}

// reconstructs schema+name from "schema.name" string
func TableFromString(value string) Table {
	parts := strings.SplitN(value, ".", 2)
	if len(parts) == 2 {
		return Table{Schema: parts[0], Name: parts[1]}
	}
	return Table{Schema: "", Name: parts[0]}
}

func (columns ColumnList) String() string {
	var columnNames []string
	for _, col := range columns {
		columnNames = append(columnNames, col.Name)
	}
	return strings.Join(columnNames, ",")
}

func (column Column) String() string {
	return column.Name
}

func (fk Fk) String() string {
	return fmt.Sprintf("%s %s(%s) => %s(%s)", fk.Name, fk.SourceTable, fk.SourceColumns.String(), fk.DestinationTable, fk.DestinationColumns.String())
}

// returns nil if not found.
// searches on schema+name
func (database Database) FindTable(tableToFind *Table) (table *Table) {
	for _, table := range database.Tables {
		if (!database.Supports.Schema || table.Schema == tableToFind.Schema) && table.Name == tableToFind.Name {
			return table
		}
	}
	return nil
}

func (table Table) FindColumn(columnName string) (index int, column *Column) {
	for index, col := range table.Columns {
		if col.Name == columnName {
			return index, col
		}
	}
	return -1, nil
}
