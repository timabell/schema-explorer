package driver_interface

import (
	"github.com/timabell/schema-explorer/schema"
	"fmt"
)

// GetRows adds extra columns for peeking over foreign keys in the selected table,
// which then need to be known about by the renderer. This class is the bridge between
// the two sides.
type PeekLookup struct {
	Table                  *schema.Table
	Fks                    []*schema.Fk
	OutboundPeekStartIndex int
	InboundPeekStartIndex  int
	PeekColumnCount        int
}

// Figures out the index of the peek column in the returned dataset for the given fk & column.
// Intended to be used by the renderer to get the data it needs for peeking.
func (peekFinder *PeekLookup) Find(peekFk *schema.Fk, peekCol *schema.Column) (peekDataIndex int) {
	peekDataIndex = peekFinder.OutboundPeekStartIndex
	for _, storedFk := range peekFinder.Fks {
		for _, col := range storedFk.DestinationTable.PeekColumns {
			if peekFk == storedFk && peekCol == col {
				return
			}
			peekDataIndex++
		}
	}
	panic(fmt.Sprintf("didn't find peek fk %s col %s in PeekLookup data", peekFk, peekCol))
}

func (peekFinder *PeekLookup) FindInbound(peekFk *schema.Fk) (peekDataIndex int) {
	for ix, fk := range peekFinder.Table.InboundFks {
		if peekFk == fk {
			peekDataIndex = peekFinder.InboundPeekStartIndex + ix
			return
		}
	}
	panic(fmt.Sprintf("Didn't find inbound fk %s in table.InboundFks", peekFk))
}
