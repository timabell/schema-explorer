package schema

import (
	"bytes"
	"fmt"
	"strings"
)

type SupportedFeatures struct {
	Schema       bool
	Descriptions bool
}

type Database struct {
	Tables            []*Table
	Fks               []*Fk
	Supports          SupportedFeatures
	Description       string
	DefaultSchemaName string
}

type Table struct {
	Schema      string
	Name        string
	Columns     ColumnList
	Fks         []*Fk
	InboundFks  []*Fk
	Description string
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
	Name        string
	Type        string
	Fk          *Fk
	Description string
}

// todo: convert to pointers to tables & columns for memory efficiency
type Fk struct {
	SourceTable        *Table
	SourceColumns      ColumnList
	DestinationTable   *Table
	DestinationColumns ColumnList
}

// Simplified fk constructor for single-column foreign keys
func NewFk(sourceTable *Table, sourceColumn *Column, destinationTable *Table, destinationColumn *Column) *Fk {
	return &Fk{SourceTable: sourceTable, SourceColumns: ColumnList{sourceColumn}, DestinationTable: destinationTable, DestinationColumns: ColumnList{destinationColumn}}
}

func (table Table) String() string {
	if table.Schema == "" {
		return table.Name
	}
	return table.Schema + "." + table.Name
}

func TableFromString(value string) Table {
	parts := strings.SplitN(value, ".", 2)
	if len(parts) == 2 {
		return Table{Schema: parts[0], Name: parts[1]}
	}
	return Table{Schema: "", Name: parts[0]}
}

func TableDebug(table *Table) string {
	return fmt.Sprintf("%s: | cols: %s | fks: %s | inboundFks: %s", table.String(), table.Columns, table.Fks, table.InboundFks)
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
	return fmt.Sprintf("%s(%s) => %s(%s)", fk.SourceTable, fk.SourceColumns.String(), fk.DestinationTable, fk.DestinationColumns.String())
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

func (database Database) DebugString() string {
	var buffer bytes.Buffer
	buffer.WriteString("database debug dump:\n")
	buffer.WriteString("tables:\n")
	for _, table := range database.Tables {
		buffer.WriteString("- ")
		buffer.WriteString(table.String())
		buffer.WriteString(" ")
		buffer.WriteString(fmt.Sprintf("%p", table))
		buffer.WriteString(" - ")
		buffer.WriteString(table.Description)
		buffer.WriteString("\n")
		for _, col := range table.Columns {
			buffer.WriteString("  - '")
			buffer.WriteString(col.Name)
			buffer.WriteString("'\t")
			buffer.WriteString(col.Type)
			buffer.WriteString("\t")
			buffer.WriteString(fmt.Sprintf("%p", col))
			buffer.WriteString("\t\"")
			buffer.WriteString(col.Description)
			buffer.WriteString("\"\n")
			if col.Fk != nil {
				buffer.WriteString("    - ")
				buffer.WriteString(col.Fk.String())
				buffer.WriteString(" ")
				buffer.WriteString(fmt.Sprintf("%p", col.Fk))
				buffer.WriteString("\n")
			}
		}
		for _, fk := range table.Fks {
			buffer.WriteString("  - ")
			buffer.WriteString(fk.String())
			buffer.WriteString(" ")
			buffer.WriteString(fmt.Sprintf("%p", fk))
			buffer.WriteString("\n")
			buffer.WriteString("    - ")
			buffer.WriteString(fk.SourceTable.String())
			buffer.WriteString("\t")
			buffer.WriteString(fmt.Sprintf("%p", fk.SourceTable))
			buffer.WriteString("\n")
			for _, col := range fk.SourceColumns {
				buffer.WriteString("        - '")
				buffer.WriteString(col.Name)
				buffer.WriteString("\t")
				buffer.WriteString(fmt.Sprintf("%p", col))
				buffer.WriteString("\n")
			}
			buffer.WriteString("    - ")
			buffer.WriteString(fk.DestinationTable.String())
			buffer.WriteString("\t")
			buffer.WriteString(fmt.Sprintf("%p", fk.DestinationTable))
			buffer.WriteString("\n")
			for _, col := range fk.DestinationColumns {
				buffer.WriteString("        - '")
				buffer.WriteString(col.Name)
				buffer.WriteString("\t")
				buffer.WriteString(fmt.Sprintf("%p", col))
				buffer.WriteString("\n")
			}
		}
		for _, fk := range table.InboundFks {
			buffer.WriteString("  - ")
			buffer.WriteString(fk.String())
			buffer.WriteString(" ")
			buffer.WriteString(fmt.Sprintf("%p", fk))
			buffer.WriteString("\n")
			buffer.WriteString("    - ")
			buffer.WriteString(fk.SourceTable.String())
			buffer.WriteString("\t")
			buffer.WriteString(fmt.Sprintf("%p", fk.SourceTable))
			buffer.WriteString("\n")
			for _, col := range fk.SourceColumns {
				buffer.WriteString("        - '")
				buffer.WriteString(col.Name)
				buffer.WriteString("\t")
				buffer.WriteString(fmt.Sprintf("%p", col))
				buffer.WriteString("\n")
			}
			buffer.WriteString("    - ")
			buffer.WriteString(fk.DestinationTable.String())
			buffer.WriteString("\t")
			buffer.WriteString(fmt.Sprintf("%p", fk.DestinationTable))
			buffer.WriteString("\n")
			for _, col := range fk.DestinationColumns {
				buffer.WriteString("        - '")
				buffer.WriteString(col.Name)
				buffer.WriteString("\t")
				buffer.WriteString(fmt.Sprintf("%p", col))
				buffer.WriteString("\n")
			}
		}
	}
	buffer.WriteString("fks:\n")
	for _, fk := range database.Fks {
		buffer.WriteString("- ")
		buffer.WriteString(fk.String())
		buffer.WriteString(" ")
		buffer.WriteString(fmt.Sprintf("%p", fk))
		buffer.WriteString("\n")
		buffer.WriteString("  - ")
		buffer.WriteString(fk.SourceTable.String())
		buffer.WriteString("\t")
		buffer.WriteString(fmt.Sprintf("%p", fk.SourceTable))
		buffer.WriteString("\n")
		for _, col := range fk.SourceColumns {
			buffer.WriteString("      - '")
			buffer.WriteString(col.Name)
			buffer.WriteString("\t")
			buffer.WriteString(fmt.Sprintf("%p", col))
			buffer.WriteString("\n")
		}
		buffer.WriteString("  - ")
		buffer.WriteString(fk.DestinationTable.String())
		buffer.WriteString("\t")
		buffer.WriteString(fmt.Sprintf("%p", fk.DestinationTable))
		buffer.WriteString("\n")
		for _, col := range fk.DestinationColumns {
			buffer.WriteString("      - '")
			buffer.WriteString(col.Name)
			buffer.WriteString("\t")
			buffer.WriteString(fmt.Sprintf("%p", col))
			buffer.WriteString("\n")
		}
	}
	return buffer.String()
}
