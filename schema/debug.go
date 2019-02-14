package schema

import (
	"bytes"
	"fmt"
)

func TableDebug(table *Table) string {
	return fmt.Sprintf("%s: | pk: %s | cols: %s | fks: %s | inboundFks: %s", table.String(), table.Pk, table.Columns, table.Fks, table.InboundFks)
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
			if col.Fks != nil {
				for _, fk := range col.Fks {
					buffer.WriteString("    - ")
					buffer.WriteString(fk.String())
					buffer.WriteString(" ")
					buffer.WriteString(fmt.Sprintf("%p", fk))
					buffer.WriteString("\n")
				}
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
