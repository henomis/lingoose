package sqlpipeline

import (
	"database/sql"
	"fmt"
	"strings"
)

//nolint:lll
const sqliteDataSourcePromptTemplate = `
You are a SQLite expert. Given an input question, first create a syntactically correct SQLite query to run, then look at the results of the query and return the answer to the input question.
Unless the user specifies in the question a specific number of examples to obtain, query for at most {{.top_k}} results using the LIMIT clause as per SQLite. You can order the results to return the most informative data in the database.
Never query for all columns from a table. You must query only the columns that are needed to answer the question. Wrap each column name in double quotes (") to denote them as delimited identifiers.
Pay attention to use only the column names you can see in the tables below. Be careful to not query for columns that do not exist. Also, pay attention to which column is in which table.
Pay attention to use date('now') function to get the current date, if the question involves "today".`

//nolint:funlen,gocognit
func getSqliteSchema(db *sql.DB) (string, error) {
	var schema string

	// Retrieve table names
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table'")
	if err != nil {
		return "", err
	}
	err = rows.Err()
	if err != nil {
		return "", err
	}
	defer rows.Close()

	// Loop through tables and retrieve schema
	for rows.Next() {
		var tableName string
		if errScan := rows.Scan(&tableName); errScan != nil {
			return "", errScan
		}

		// Retrieve column information
		cols, errQuery := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
		if errQuery != nil {
			return "", errQuery
		}
		if errRows := cols.Err(); errRows != nil {
			return "", errRows
		}
		defer cols.Close()

		// Build CREATE TABLE statement
		var createTableStmt strings.Builder
		createTableStmt.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", tableName))

		for cols.Next() {
			var (
				colNum     int
				colName    string
				colType    string
				notNull    int
				defaultVal sql.NullString
				primaryKey int
			)
			if errScan := cols.Scan(&colNum, &colName, &colType, &notNull, &defaultVal, &primaryKey); errScan != nil {
				return "", errScan
			}

			// Build column definition
			colDef := fmt.Sprintf("  %s %s", colName, colType)
			if notNull == 1 {
				colDef += " NOT NULL"
			}
			if defaultVal.Valid {
				colDef += fmt.Sprintf(" DEFAULT '%s'", defaultVal.String)
			}
			if primaryKey == 1 {
				colDef += " PRIMARY KEY"
			}

			createTableStmt.WriteString(colDef)
			createTableStmt.WriteString(",\n")
		}

		// Retrieve foreign key information
		fks, errQuery := db.Query(fmt.Sprintf("PRAGMA foreign_key_list(%s)", tableName))
		if errQuery != nil {
			return "", errQuery
		}
		if errRows := fks.Err(); errRows != nil {
			return "", errRows
		}
		defer fks.Close()

		// Build foreign key definitions
		var fkDefs []string
		for fks.Next() {
			var (
				id            int
				seq           int
				table         string
				from          string
				to            string
				onUpdate      string
				onDelete      string
				match         string
				foreignKeyDef string
			)
			if errScan := fks.Scan(&id, &seq, &table, &from, &to, &onUpdate, &onDelete, &match); errScan != nil {
				return "", errScan
			}

			foreignKeyDef = fmt.Sprintf("  FOREIGN KEY (%s) REFERENCES %s(%s)", from, table, to)
			if onUpdate != "" {
				foreignKeyDef += fmt.Sprintf(" ON UPDATE %s", onUpdate)
			}
			if onDelete != "" {
				foreignKeyDef += fmt.Sprintf(" ON DELETE %s", onDelete)
			}
			if match != "" {
				foreignKeyDef += fmt.Sprintf(" MATCH %s", match)
			}

			fkDefs = append(fkDefs, foreignKeyDef)
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
