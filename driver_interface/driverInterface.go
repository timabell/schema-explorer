package driver_interface

import (
	"bitbucket.org/timabell/sql-data-viewer/params"
	"bitbucket.org/timabell/sql-data-viewer/schema"
	"database/sql"
)

type DbReader interface {
	// does select or something to make sure we have a working db connection,
	// after this has succeeded Connected() will return true
	CheckConnection(databaseName string) (err error)

	// true if CheckConnection() has been run and succeeded
	Connected() bool

	// parse the whole schema info into memory
	ReadSchema(databaseName string) (database *schema.Database, err error)

	// populate the table row counts
	UpdateRowCounts(database *schema.Database) (err error)

	// get some data, obeying sorting, filtering etc in the table params
	GetSqlRows(databaseName string, table *schema.Table, params *params.TableParams, peekFinder *PeekLookup) (rows *sql.Rows, err error)

	// get a count for the supplied filters, for use with paging and overview info
	GetRowCount(databaseName string, table *schema.Table, params *params.TableParams) (rowCount int, err error)

	// get breakdown of most common values in each column
	GetAnalysis(databaseName string, table *schema.Table) (analysis []schema.ColumnAnalysis, err error)

	// get list of databases on this server (if supported)
	ListDatabases() (databaseList []string, err error)

	CanSwitchDatabase() bool

	GetConfiguredDatabaseName() string

	SetTableDescription(database string, table string, description string) (err error)
}
