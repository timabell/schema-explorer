package schema

import (
	"fmt"
	"strings"
)

type SupportedFeatures struct {
	Schema bool
}

type Database struct {
	Tables   []*Table
	Fks      []*Fk
	Supports SupportedFeatures
}

type Table struct {
	Schema     string
	Name       string
	Columns    []*Column
	Fks        []*Fk
	InboundFks []*Fk
}

type Column struct {
	Name string
	Type string
	Fk   *Fk
}

// todo: convert to pointers to tables & columns for memory efficiency
type Fk struct {
	SourceTable        *Table
	SourceColumns      []*Column
	DestinationTable   *Table
	DestinationColumns []*Column
}

// filtering of results with column name / value(s) pairs,
// matches type of url.Values so can pass straight through
type RowFilter map[string][]string

// Simplified fk constructor for single-column foreign keys
func NewFk(sourceTable *Table, sourceColumn *Column, destinationTable *Table, destinationColumn *Column) *Fk {
	return &Fk{SourceTable: sourceTable, SourceColumns: []*Column{sourceColumn}, DestinationTable: destinationTable, DestinationColumns: []*Column{destinationColumn}}
}

func (table Table) String() string {
	if table.Schema == "" {
		return table.Name
	}
	return table.Schema + "." + table.Name
}

func columnsString(columns []*Column) string {
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
	return fmt.Sprintf("%s(%s) => %s(%s)", fk.SourceTable, columnsString(fk.SourceColumns), fk.DestinationTable, columnsString(fk.DestinationColumns))
}

// filter the fk list down to keys that reference the "child" table
// todo: not sure this should live here conceptually
//func (child Table) FindParents(fks GlobalFkList) (parents GlobalFkList) {
//	parents = GlobalFkList{}
//	for srcTable, tableFks := range fks {
//		newFkList := FkList{}
//		for srcCol, ref := range tableFks {
//			if ref.Table.String() == child.String() {
//				// match; copy into new list
//				newFkList[srcCol] = ref
//				parents[srcTable] = newFkList
//			}
//		}
//	}
//	return
//}

// returns nil if not found.
// searches on schema+name
func (database Database) FindTable(tableToFind *Table) (table *Table) {
	for _, table := range database.Tables {
		if table.Schema == tableToFind.Schema && table.Name == tableToFind.Name {
			return table
		}
	}
	return nil
}

func (table Table) FindCol(columnName string) (found bool, index int) {
	for index, col := range table.Columns {
		if col.Name == columnName {
			return true, index
		}
	}
	return false, 0
}
