package reader

import (
	"bitbucket.org/timabell/sql-data-viewer/params"
	"bitbucket.org/timabell/sql-data-viewer/schema"
	"database/sql"
	"fmt"
	"github.com/jessevdk/go-flags"
	"log"
	"os"
	"strings"
)

type SseOptions struct {
	Driver                *string `short:"d" long:"driver" required:"true" description:"Driver to use" choice:"mssql" choice:"pg" choice:"sqlite" env:"schemaexplorer_driver"`
	Live                  *bool   `short:"l" long:"live" description:"update html templates & schema information from disk on every page load" env:"schemaexplorer_live"`
	ConnectionDisplayName *string `short:"n" long:"display-name" description:"A display name for this connection" env:"schemaexplorer_display_name"`
	ListenOnAddress       *string `short:"a" long:"listen-on-address" description:"address to listen on" default:"localhost" env:"schemaexplorer_listen_on_address"` // localhost so that it's secure by default, only listen for local connections
	ListenOnPort          *int    `short:"p" long:"listen-on-port" description:"port to listen on" default:"8080" env:"schemaexplorer_listen_on_port"`
}

// todo: arg parsing and options shouldn't be here
var Options = SseOptions{}
var ArgParser = flags.NewParser(&Options, flags.Default)

type DbReader interface {
	// does select or something to make sure we have a working db connection
	CheckConnection() (err error)

	// parse the whole schema info into memory
	ReadSchema() (database *schema.Database, err error)

	// populate the table row counts
	UpdateRowCounts(database *schema.Database) (err error)

	// get some data, obeying sorting, filtering etc in the table params
	GetSqlRows(table *schema.Table, params *params.TableParams) (rows *sql.Rows, err error)

	// get a count for the supplied filters, for use with paging and overview info
	GetRowCount(table *schema.Table, params *params.TableParams) (rowCount int, err error)

	// get breakdown of most common values in each column
	GetAnalysis(table *schema.Table) (analysis []schema.ColumnAnalysis, err error)
}

type DbReaderOptions interface{}

type CreateReader func() DbReader

// Single row of data
type RowData []interface{}

var creators = make(map[string]CreateReader)

func init() {
	ArgParser.EnvNamespace = "schemaexplorer"
	ArgParser.NamespaceDelimiter = "-"
}

func RegisterReader(name string, opt interface{}, creator CreateReader) {
	creators[name] = creator
	group, err := ArgParser.AddGroup(name, fmt.Sprintf("Options for %s database", name), opt)
	if err != nil {
		panic(err)
	}
	group.Namespace = name
	group.EnvNamespace = name
}

func GetDbReader() DbReader {
	createReader := creators[*Options.Driver]
	if createReader == nil {
		log.Printf("Unknown reader '%s'", *Options.Driver)
		os.Exit(1)
	}
	return createReader()
}

// GetRows adds extra columns for peeking over foreign keys in the selected table,
// which then need to be known about by the renderer. This class is the bridge between
// the two sides.
type PeekLookup struct {
	fks []*schema.Fk // foreign keys referenced in this table/query
}

// Figures out the index of the peek column in the returned dataset for the given fk & column.
// Intended to be used by the renderer to get the data it needs for peeking.
func (peekFinder PeekLookup) Find(peekFk *schema.Fk, peekCol *schema.Column) (peekDataIndex int){
	peekDataIndex = 0
	for _, storedFk := range peekFinder.fks{
		for _, col := range storedFk.DestinationTable.PeekColumns{
			if peekFk == storedFk && peekCol == col{
				return
			}
			peekDataIndex++
		}
	}
	panic("didn't find peek fk/col in PeekLookup data")
}

func GetRows(reader DbReader, table *schema.Table, params *params.TableParams) (rowsData []RowData, peekFinder *PeekLookup, err error) {
	rows, err := reader.GetSqlRows(table, params)
	if rows == nil {
		panic("GetSqlRows() returned nil")
	}
	defer rows.Close()
	if len(table.Columns) == 0 {
		panic("No columns found when reading table data table")
	}
	rowsData, err = GetAllData(len(table.Columns), rows)
	if err != nil {
		return nil, nil, err
	}
	return
}

func GetAllData(colCount int, rows *sql.Rows) (rowsData []RowData, err error) {
	for rows.Next() {
		row, err := GetRow(colCount, rows)
		if err != nil {
			return nil, err
		}
		rowsData = append(rowsData, row)
	}
	return
}

func GetRow(colCount int, rows *sql.Rows) (rowsData RowData, err error) {
	// http://stackoverflow.com/a/23507765/10245 - getting ad-hoc column data
	singleRow := make([]interface{}, colCount)
	rowDataPointers := make([]interface{}, colCount)
	for i := 0; i < colCount; i++ {
		rowDataPointers[i] = &singleRow[i]
	}
	err = rows.Scan(rowDataPointers...)
	if err != nil {
		log.Println("error reading row data", err)
		return nil, err
	}
	return singleRow, err
}

func DbValueToString(colData interface{}, dataType string) *string {
	var stringValue string
	uuidLen := 16
	switch {
	case colData == nil:
		return nil
	case dataType == "money": // mssql money
		fallthrough
	case dataType == "decimal": // mssql decimal
		fallthrough
	case dataType == "numeric": // mssql numeric
		stringValue = fmt.Sprintf("%s", colData) // seems to come back as byte array for a string, surprising, could be a driver thing
	case dataType == "integer":
		stringValue = fmt.Sprintf("%d", colData)
	case dataType == "float":
		stringValue = fmt.Sprintf("%f", colData)
	case dataType == "uniqueidentifier": // mssql guid
		bytes := colData.([]byte)
		if len(bytes) != uuidLen {
			panic(fmt.Sprintf("Unexpected byte-count for uniqueidentifier, expected %d, got %d. Value: %+v", uuidLen, len(bytes), colData))
		}
		stringValue = fmt.Sprintf("%x%x%x%x-%x%x-%x%x-%x%x-%x%x%x%x%x%x",
			bytes[3], bytes[2], bytes[1], bytes[0], bytes[5], bytes[4], bytes[7], bytes[6], bytes[8], bytes[9], bytes[10], bytes[11], bytes[12], bytes[13], bytes[14], bytes[15])
	case dataType == "text": // sqlite
		fallthrough
	case strings.Contains(strings.ToLower(dataType), "varchar"): // sqlite is [N]VARCHAR sqlserver is [n]varchar
		stringValue = fmt.Sprintf("%s", colData)
	case strings.Contains(dataType, "TEXT"): // mssql
		// https://stackoverflow.com/a/18615786/10245
		bytes := colData.([]uint8)
		stringValue = fmt.Sprintf("%s", bytes)
	case dataType == "varbinary": // mssql varbinary
		stringValue = "[binary]"
	default:
		stringValue = fmt.Sprintf("%v", colData)
	}
	return &stringValue
}
