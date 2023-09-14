package sqlpipeline

import (
	"database/sql"
	"fmt"
	"strings"
)

//nolint:lll
const mysqlDataSourcePromptTemplate = "\n" +
	"You are a MySQL expert. Given an input question, first create a syntactically correct MySQL query to run, then look at the results of the query and return the answer to the input question.\n" +
	"Unless the user specifies in the question a specific number of examples to obtain, query for at most {top_k} results using the LIMIT clause as per MySQL. You can order the results to return the most informative data in the database.\n" +
	"Never query for all columns from a table. You must query only the columns that are needed to answer the question. Wrap each column name in backticks (`) to denote them as delimited identifiers.\n" +
	"Pay attention to use only the column names you can see in the tables below. Be careful to not query for columns that do not exist. Also, pay attention to which column is in which table.\n" +
	"Pay attention to use CURDATE() function to get the current date, if the question involves \"today\"."

func getMySQLSchema(db *sql.DB, dbName string) (string, error) {
	var schema string

	// Retrieve table names
	//nolint:lll
	rows, err := db.Query(fmt.Sprintf("SELECT table_name FROM information_schema.tables WHERE table_schema = '%s'", dbName))
	if err != nil {
		return "", err
	}
	defer rows.Close()

	// Loop through tables and retrieve schema
	for rows.Next() {
		var tableName string
		if errQuery := rows.Scan(&tableName); errQuery != nil {
			return "", errQuery
		}

		// Retrieve column information
		cols, errQuery := db.Query(fmt.Sprintf("SHOW COLUMNS FROM %s", tableName))
		if errQuery != nil {
			return "", errQuery
		}
		defer cols.Close()

		// Build CREATE TABLE statement
		var createTableStmt strings.Builder
		createTableStmt.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", tableName))

		for cols.Next() {
			var (
				field sql.NullString
				typ   sql.NullString
				null  sql.NullString
				key   sql.NullString
				def   sql.NullString
				extra sql.NullString
			)
			if errScan := cols.Scan(&field, &typ, &null, &key, &def, &extra); errScan != nil {
				return "", errScan
			}

			// Build column definition
			colDef := fmt.Sprintf("  %s %s", field.String, typ.String)
			if null.Valid && null.String == "YES" {
				colDef += " NULL"
			} else if null.Valid {
				colDef += " NOT NULL"
			}
			if def.Valid {
				colDef += fmt.Sprintf(" DEFAULT '%s'", def.String)
			}
			if key.Valid && key.String == "PRI" {
				colDef += " PRIMARY KEY"
			}
			if key.Valid && key.String == "MUL" {
				colDef += " KEY"
			}
			if extra.Valid && extra.String != "" {
				colDef += fmt.Sprintf(" %s", extra.String)
			}

			createTableStmt.WriteString(colDef)
			createTableStmt.WriteString(",\n")
		}

		// Retrieve foreign key information
		//nolint:lll
		fks, errQuery := db.Query(fmt.Sprintf("SELECT constraint_name, column_name, referenced_table_name, referenced_column_name FROM information_schema.key_column_usage WHERE table_schema = '%s' AND table_name = '%s' AND referenced_table_name IS NOT NULL", dbName, tableName))
		if errQuery != nil {
			return "", errQuery
		}
		defer fks.Close()

		// Build foreign key definitions
		var fkDefs []string
		for fks.Next() {
			var (
				constraintName       string
				columnName           string
				referencedTableName  string
				referencedColumnName string
			)
			if errScan := fks.Scan(&constraintName, &columnName, &referencedTableName, &referencedColumnName); errScan != nil {
				return "", errScan
			}

			fkDef := fmt.Sprintf("  CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s)",
				constraintName, columnName, referencedTableName, referencedColumnName)
			fkDefs = append(fkDefs, fkDef)
		}

		// Add foreign key definitions to CREATE TABLE statement
		if len(fkDefs) > 0 {
			createTableStmt.WriteString(strings.Join(fkDefs, ",\n"))
			createTableStmt.WriteString(",\n")
		}

		// Remove trailing comma and add closing parenthesis
		content := createTableStmt.String()
		content = content[:len(content)-2]
		createTableStmt.Reset()
		createTableStmt.WriteString(content)

		createTableStmt.WriteString("\n);")

		schema += createTableStmt.String() + "\n\n"
	}

	return schema, nil
}
