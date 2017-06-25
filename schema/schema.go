package schema

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
// todo: not sure this should live here conceptually
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

