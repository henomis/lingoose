package main

import (
	"database/sql"
	"fmt"
	"strings"
)

func getDatabaseSchema(database string) (string, error) {
	conn, err := sql.Open("sqlite3", database)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	databaseSchemaDict := getDatabaseInfo(conn)
	databaseSchemaString := ""
	for _, table := range databaseSchemaDict {
		databaseSchemaString += fmt.Sprintf("Table: %s\nColumns: %s\n", table["table_name"], table["column_names"])
	}

	return databaseSchemaString, nil
}

func getTableNames(conn *sql.DB) []string {
	tableNames := []string{}
	tables, err := conn.Query("SELECT name FROM sqlite_master WHERE type='table';")
	if err != nil {
		fmt.Println(err)
		return tableNames
	}
	defer tables.Close()
	for tables.Next() {
		var tableName string
		err := tables.Scan(&tableName)
		if err != nil {
			fmt.Println(err)
			return tableNames
		}
		tableNames = append(tableNames, tableName)
	}
	return tableNames
}

func getColumnNames(conn *sql.DB, tableName string) []string {
	columnNames := []string{}
	columns, err := conn.Query(fmt.Sprintf("PRAGMA table_info('%s');", tableName))
	if err != nil {
		fmt.Println(err)
		return columnNames
	}
	defer columns.Close()
	for columns.Next() {
		var cid int
		var name string
		var dataType string
		var notNull int
		var dfltValue sql.NullString
		var pk int
		err := columns.Scan(&cid, &name, &dataType, &notNull, &dfltValue, &pk)

		if err != nil {
			fmt.Println(err)
			return columnNames
		}
		columnNames = append(columnNames, name)
	}
	return columnNames
}

func getDatabaseInfo(conn *sql.DB) []map[string]interface{} {
	tableDicts := []map[string]interface{}{}
	for _, tableName := range getTableNames(conn) {
		columnNames := getColumnNames(conn, tableName)
		tableDict := map[string]interface{}{
			"table_name":   tableName,
			"column_names": strings.Join(columnNames, ", "),
		}
		tableDicts = append(tableDicts, tableDict)
	}
	return tableDicts
}
