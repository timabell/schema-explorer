package trail

import (
	"github.com/timabell/schema-explorer/schema"
	"strings"
)

type TrailLog struct {
	Tables  []string
	Dynamic bool // whether this is dynamic from cookies or is from a permalink, for altering UI
}

func (trail TrailLog) AsCsv() string {
	return strings.Join(trail.Tables, ",")
}

func (trail *TrailLog) AddTable(table *schema.Table) {
	var exists = false
	for _, x := range trail.Tables {
		if x == table.String() {
			exists = true
		}
	}
	if !exists {
		trail.Tables = append(trail.Tables, table.String())
	}
}
