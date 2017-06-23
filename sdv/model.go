package sdv

import (
	"database/sql"
	"fmt"
	"log"
)

// alias to make it clear when we're dealing with table names
type TableName string

// alias to make it clear when we're dealing with column names
type ColumnName string

// filtering of results with column name / value(s) pairs,
// matches type of url.Values so can pass straight through
type RowFilter map[string][]string

// reference to a field in another table, part of a foreign key
type Ref struct {
	Table TableName  // target table for the fk
	Col   ColumnName // target col for the fk
}

// list of foreign keys, the column in the current table that the fk is defined on
type FkList map[ColumnName]Ref

// for each table in the database, the list of fks defined on that table
type GlobalFkList map[TableName]FkList


// filter the fk list down to keys that reference the "child" table
func (child TableName) FindParents(fks GlobalFkList) (parents GlobalFkList) {
	parents = GlobalFkList{}
	for srcTable, tableFks := range fks {
		newFkList := FkList{}
		for srcCol, ref := range tableFks {
			if ref.Table == child {
				// match; copy into new list
				newFkList[srcCol] = ref
				parents[srcTable] = newFkList
			}
		}
	}
	return
}

func AllFks(dbc *sql.DB) (allFks GlobalFkList) {
	tables, err := GetTables(dbc)
	if err != nil {
		fmt.Println("error getting table list while building global fk list", err)
		return
	}
	allFks = GlobalFkList{}
	for _, table := range tables {
		allFks[table] = fks(dbc, table)
	}
	return
}

func fks(dbc *sql.DB, table TableName) (fks FkList) {
	rows, err := dbc.Query("PRAGMA foreign_key_list('" + string(table) + "');")
	if err != nil {
		log.Println("select error", err)
		return
	}
	defer rows.Close()
	fks = FkList{}
	for rows.Next() {
		var id, seq int
		var parentTable, from, to, onUpdate, onDelete, match string
		rows.Scan(&id, &seq, &parentTable, &from, &to, &onUpdate, &onDelete, &match)
		thisRef := Ref{Col: ColumnName(to), Table: TableName(parentTable)}
		fks[ColumnName(from)] = thisRef
	}
	return
}

func GetTables() (tables []TableName, err error) {
	rows, err := dbc.Query("SELECT name FROM sqlite_master WHERE type='table';")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		rows.Scan(&name)
		tables = append(tables, TableName(name))
	}
	return tables, nil
}

