package reader

import (
	"bitbucket.org/timabell/sql-data-viewer/params"
	"bitbucket.org/timabell/sql-data-viewer/schema"
	"database/sql"
	"fmt"
	"github.com/jessevdk/go-flags"
	"log"
	"strings"
	"os"
)

type SdvOptions struct {
	Driver                *string `short:"d" long:"driver" required:"true" description:"Driver to use (mssql, pg or sqlite)" env:"schemaexplorer_driver"`
	Live                  *bool   `short:"l" long:"live" description:"update html templates & schema information from disk on every page load" env:"schemaexplorer_live"`
	ConnectionDisplayName *string `short:"n" long:"display-name" description:"A display name for this connection" env:"schemaexplorer_displayname"`
	ListenOnAddress       *string `short:"a" long:"listen-on-address" description:"address to listen on" default:"localhost" env:"schemaexplorer_listenonaddress"` // localhost so that it's secure by default, only listen for local connections
	ListenOnPort          *int    `short:"p" long:"listen-on-port" description:"port to listen on" default:"8080" env:"schemaexplorer_listenonport"`
}

// todo: arg parsing and options shouldn't be here
var Options = SdvOptions{}
var ArgParser = flags.NewParser(&Options, flags.Default)

type DbReader interface {
	CheckConnection() (err error)
	ReadSchema() (database *schema.Database, err error)
	GetSqlRows(table *schema.Table, params *params.TableParams) (rows *sql.Rows, err error)
}

type DbReaderOptions interface{}

type CreateReader func() DbReader

// Single row of data
type RowData []interface{}

var creators = make(map[string]CreateReader)

func RegisterReader(name string, opt interface{}, creator CreateReader) {
	creators[name] = creator
	ArgParser.AddGroup(name, fmt.Sprintf("Options for %s database", name), opt)
	log.Printf("%s capability locked and loaded", name)
}

func GetDbReader() DbReader {
	log.Printf("Initializing %s reader", *Options.Driver)
	createReader := creators[*Options.Driver]
	if createReader == nil{
		log.Printf("Unknown reader '%s'", *Options.Driver)
		os.Exit(1)
	}
	return createReader()
}

func GetRows(reader DbReader, table *schema.Table, params *params.TableParams) (rowsData []RowData, err error) {
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
		return nil, err
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
