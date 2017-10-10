package sdv

import (
	"bitbucket.org/timabell/sql-data-viewer/schema"
	"database/sql"
	"fmt"
	"log"
	"strings"
)

type dbReader interface {
	CheckConnection() (err error)
	GetTables() (tables []schema.Table, err error)
	AllFks() (allFks schema.GlobalFkList, err error)
	GetSqlRows(query schema.RowFilter, table schema.Table, rowLimit int) (rows *sql.Rows, err error)
	GetColumns(table schema.Table) (cols []schema.Column, err error)
}

// Single row of data
type RowData []interface{}

func GetRows(reader dbReader, query schema.RowFilter, table schema.Table, colCount int, rowLimit int) (rowsData []RowData, err error) {
	rows, err := reader.GetSqlRows(query, table, rowLimit)
	if rows == nil {
		panic("GetSqlRows() returned nil")
	}
	defer rows.Close()

	rowsData, err = GetAllData(colCount, rows)
	if err != nil {
		return nil, err
	}
	return
}

func GetAllData(colCount int, rows *sql.Rows) (rowsData []RowData, err error) {
	// http://stackoverflow.com/a/23507765/10245 - getting ad-hoc column data
	singleRow := make([]interface{}, colCount)
	rowDataPointers := make([]interface{}, colCount)
	for i := 0; i < colCount; i++ {
		rowDataPointers[i] = &singleRow[i]
	}
	for rows.Next() {
		err := rows.Scan(rowDataPointers...)
		if err != nil {
			log.Println("error reading row data", err)
			return nil, err
		}
		rowsData = append(rowsData, singleRow)
	}
	return
}

func DbValueToString(colData interface{}, dataType string) *string {
	var stringValue string
	switch {
	case colData == nil:
		return nil
	case dataType == "integer":
		stringValue = fmt.Sprintf("%d", colData)
	case dataType == "float":
		stringValue = fmt.Sprintf("%f", colData)
	case dataType == "text": // sqlite
		fallthrough
	case strings.Contains(strings.ToLower(dataType), "varchar"): // sqlite is [N]VARCHAR sqlserver is [n]varchar
		stringValue = fmt.Sprintf("%s", colData)
	case strings.Contains(dataType, "TEXT"): // mssql
		// https://stackoverflow.com/a/18615786/10245
		bytes := colData.([]uint8)
		stringValue = fmt.Sprintf("%s", bytes)
	default:
		stringValue = fmt.Sprintf("%v", colData)
	}
	return &stringValue
}
