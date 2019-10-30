package reader

import (
	"github.com/timabell/schema-explorer/driver_interface"
	"github.com/timabell/schema-explorer/drivers"
	"github.com/timabell/schema-explorer/options"
	"github.com/timabell/schema-explorer/params"
	"github.com/timabell/schema-explorer/resources"
	"github.com/timabell/schema-explorer/schema"
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
)

type SchemaCache map[string]*schema.Database

// Global in-memory cache of database structures, keyed on database name.
// If multiple databases aren't supported then ignore the name and just use index zero for storage.
var Databases = make(map[string]*schema.Database)

// Single row of data
type RowData []interface{}

// This is how implementations for reading different RDBMS systems can register themselves.
// They should call this in their init() function
func RegisterReader(driver *drivers.Driver) {
	drivers.Drivers[driver.Name] = driver
	//group, err := options.ArgParser.AddGroup(driver.Name, fmt.Sprintf("Options for %s database", driver.Name), driver.Options)
	//if err != nil {
	//	panic(err)
	//}
	//group.Namespace = driver.Name
	//group.EnvNamespace = driver.Name
}

func InitializeDatabase(databaseName string) (err error) {
	dbReader := GetDbReader()
	log.Println("Checking database connection...")
	err = dbReader.CheckConnection(databaseName)
	if err != nil {
		err = errors.New("check connection failed: " + err.Error())
		return
	}

	log.Print("Reading schema, this may take a while...")
	Databases[databaseName], err = dbReader.ReadSchema(databaseName)
	if err != nil {
		err = errors.New("error reading schema: " + err.Error())
		return
	}
	Databases[databaseName].Name = databaseName
	setupPeekList(Databases[databaseName])
	return
}

func setupPeekList(database *schema.Database) {
	if options.Options == nil {
		panic("options is nil")
	}
	var peekFilename string
	if (*options.Options).PeekConfigPath == "" {
		peekFilename = path.Join(resources.BasePath, "config/peek-config.txt")
	} else {
		peekFilename = options.Options.PeekConfigPath
	}
	log.Printf("Loading peek config from %s ...", peekFilename)
	file, err := os.Open(peekFilename)
	if err != nil {
		log.Printf("Failed to load %s, disabling peek feature, check peek-config-path configuration. %s", peekFilename, err)
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var regexes []regexp.Regexp
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // skip blanks and comments
		}
		regexes = append(regexes, *regexp.MustCompile(line))
	}
	for _, tbl := range database.Tables {
		for _, col := range tbl.Columns {
			for _, regex := range regexes {
				fullName := tbl.String() + "." + col.Name
				fullNameLower := strings.ToLower(fullName)
				if regex.MatchString(fullNameLower) {
					tbl.PeekColumns = append(tbl.PeekColumns, col)
					log.Printf(" - peek configured for %s", fullName)
				}
			}
		}
	}
}

func GetDbReader() driver_interface.DbReader {
	if options.Options == nil || (*options.Options).Driver == "" {
		panic("driver option missing")
	}
	driver := drivers.Drivers[options.Options.Driver]
	if driver == nil {
		log.Printf("Unknown reader '%s'", options.Options.Driver)
		os.Exit(1)
	}
	return driver.CreateReader()
}

func GetRows(reader driver_interface.DbReader, databaseName string, table *schema.Table, params *params.TableParams) (rowsData []RowData, peekFinder *driver_interface.PeekLookup, err error) {
	// load up all the fks that we have peek info for
	peekFinder = &driver_interface.PeekLookup{}
	inboundPeekCount := 0
	for _, fk := range table.Fks {
		if len(fk.DestinationTable.PeekColumns) == 0 {
			continue
		}
		peekFinder.Fks = append(peekFinder.Fks, fk)
		inboundPeekCount += len(fk.DestinationTable.PeekColumns)
	}
	peekFinder.OutboundPeekStartIndex = len(table.Columns)
	peekFinder.InboundPeekStartIndex = peekFinder.OutboundPeekStartIndex + inboundPeekCount
	peekFinder.PeekColumnCount = inboundPeekCount + len(table.InboundFks)
	peekFinder.Table = table

	rows, err := reader.GetSqlRows(databaseName, table, params, peekFinder)
	if rows == nil {
		panic("GetSqlRows() returned nil")
	}
	defer rows.Close()
	if len(table.Columns) == 0 {
		panic("No columns found when reading table data table")
	}
	rowsData, err = getAllData(len(table.Columns)+peekFinder.PeekColumnCount, rows)
	if err != nil {
		return nil, nil, err
	}
	return
}

func getAllData(colCount int, rows *sql.Rows) (rowsData []RowData, err error) {
	for rows.Next() {
		row, err := getRow(colCount, rows)
		if err != nil {
			return nil, err
		}
		rowsData = append(rowsData, row)
	}
	return
}

func getRow(colCount int, rows *sql.Rows) (rowsData RowData, err error) {
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
	// todo: check type of colData matches type of dataTyoe - sqlite will let you insert anything into anything
	var stringValue string
	dataType = strings.ToLower(dataType)
	uuidLen := 16
	// todo: optimise order for speed, also consider possible fuzzy clashes and which one would win
	switch {
	// === // NULLs ...
	case colData == nil:
		return nil
	// === // exact matches only ...
	case dataType == "uniqueidentifier": // mssql guid
		bytes := colData.([]byte)
		if len(bytes) != uuidLen {
			panic(fmt.Sprintf("Unexpected byte-count for uniqueidentifier, expected %d, got %d. Value: %+v", uuidLen, len(bytes), colData))
		}
		stringValue = fmt.Sprintf("%x%x%x%x-%x%x-%x%x-%x%x-%x%x%x%x%x%x",
			bytes[3], bytes[2], bytes[1], bytes[0], bytes[5], bytes[4], bytes[7], bytes[6], bytes[8], bytes[9], bytes[10], bytes[11], bytes[12], bytes[13], bytes[14], bytes[15])
	// === //
	case dataType == "numeric": // sqlite - best type for number. needs casting for pg
		stringValue = fmt.Sprintf("%v", colData.(float64))
	case dataType == "varbinary":
		fallthrough
	case dataType == "blob":
		fallthrough
	// === //
	case dataType == "boolean":
		fallthrough
	// === //
	case dataType == "date":
		fallthrough
	case dataType == "datetime":
		fallthrough
	// === // more expensive fuzzy type name matches from here ...
	case dataType == "money":
		fallthrough
	case dataType == "real":
		fallthrough
	case dataType == "float":
		fallthrough
	case strings.HasPrefix(dataType, "double"):
		fallthrough
	case strings.HasPrefix(dataType, "decimal"): // todo: don't allow "decimal(10,5)"
		fallthrough
	case strings.Contains(dataType, "int"): // todo: expensive, optimise for supported values
		stringValue = fmt.Sprintf("%v", colData)
	// === //
	case dataType == "text": // sqlite
		fallthrough
	case dataType == "jsonb":
		fallthrough
	case dataType == "json":
		fallthrough
	case dataType == "clob": // sqlite - char large object
		fallthrough
	case strings.Contains(dataType, "char"): // See test sql files for things this should cover. // todo: expensive, optimise for supported values
		stringValue = fmt.Sprintf("%s", colData)
	// === //
	case strings.Contains(dataType, "text"): // mssql // todo: expensive, optimise for supported values
		// https://stackoverflow.com/a/18615786/10245
		bytes := colData.([]uint8)
		stringValue = fmt.Sprintf("%s", bytes)
	// === // unknown ...
	default:
		//log.Printf("unknown data type %s", dataType)
		//panic(fmt.Sprintf("unknown data type %s", dataType))
		stringValue = fmt.Sprintf("%v", colData) // fallback, hope for the best, but don't use this for ones we know to make it clear what works
	}
	return &stringValue
}
