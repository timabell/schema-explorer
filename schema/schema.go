package schema

import "log"

type Table struct {
	Schema string
	Name   string
}

func (table Table) String() string {
	if table.Schema == "" {
		return table.Name
	}
	return table.Schema + "." + table.Name
}

// alias to make it clear when we're dealing with column names
type Column string

// filtering of results with column name / value(s) pairs,
// matches type of url.Values so can pass straight through
type RowFilter map[string][]string

// reference to a field in another table, part of a foreign key
type Ref struct {
	Table Table  // target table for the fk
	Col   Column // target col for the fk
}

// list of foreign keys, the column in the current table that the fk is defined on
type FkList map[Column]Ref

// for each table in the database, the list of fks defined on that table
type GlobalFkList map[string]FkList

// filter the fk list down to keys that reference the "child" table
// todo: not sure this should live here conceptually
func (child Table) FindParents(fks GlobalFkList) (parents GlobalFkList) {
	log.Println("looking for fks to ", child)
	parents = GlobalFkList{}
	for srcTable, tableFks := range fks {
		log.Println("reading fk ", srcTable, tableFks)
		newFkList := FkList{}
		for srcCol, ref := range tableFks {
			log.Println("reading tablefk ", srcCol, ref)
			if ref.Table.String() == child.String() {
				log.Println("match")
				// match; copy into new list
				newFkList[srcCol] = ref
				parents[srcTable] = newFkList
			}
		}
	}
	return
}
