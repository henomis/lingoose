package sql

import (
	"database/sql"
	"fmt"
)

//nolint:lll
var psqlSystemPromptTemplate = `
You are a Postgresql expert. Given an input question, create a syntactically correct psql query to run. Do not add any extra information to the query. The query must be usable as-is.
Unless the user specifies in the question a specific number of examples to obtain, query for at most {{.top_k}} results using the LIMIT clause as per Postgresql. You can order the results to return the most informative data in the database.
Never query for all columns from a table. You must query only the columns that are needed to answer the question. Wrap each column name in double quotes (") to denote them as delimited identifiers.
Pay attention to use only the column names you can see in the tables below. Be careful to not query for columns that do not exist. Also, pay attention to which column is in which table.
Pay attention to use date('now') function to get the current date, if the question involves "today". Do not use markdown to format the query.`

// psqlSchema retrieves the schema information for all tables in a PostgreSQL database.
//
//nolint:funlen,gocognit
func (s *SQL) psqlSchema() (*string, error) {
	rows, err := s.db.Query(`
	SELECT
			c.table_name,
			c.column_name,
			c.data_type,
			c.column_default,
			c.is_nullable,
			tc.constraint_type,
			kcu.constraint_name,
			ccu.table_name AS foreign_table_name,
			ccu.column_name AS foreign_column_name
	FROM
			information_schema.columns c
	LEFT JOIN
			information_schema.key_column_usage kcu ON c.table_name = kcu.table_name AND c.column_name = kcu.column_name
	LEFT JOIN
			information_schema.table_constraints tc ON kcu.constraint_name = tc.constraint_name
	LEFT JOIN
			information_schema.constraint_column_usage ccu ON tc.constraint_name = ccu.constraint_name
	WHERE
			c.table_schema = 'public'
	ORDER BY
			c.table_name, c.ordinal_position;
`)
	if err != nil {
		return nil, fmt.Errorf("querying schema: %w", err)
	}
	defer rows.Close()

	schema := ""
	currentTable := ""

	for rows.Next() {
		//nolint:lll
		var tableName, columnName, dataType, columnDefault, isNullable, constraintType, constraintName, foreignTableName, foreignColumnName sql.NullString
		//nolint:lll
		if rowsErr := rows.Scan(&tableName, &columnName, &dataType, &columnDefault, &isNullable, &constraintType, &constraintName, &foreignTableName, &foreignColumnName); rowsErr != nil {
			return nil, fmt.Errorf("scanning row: %w", rowsErr)
		}

		//nolin:nestif
		if tableName.Valid && tableName.String != currentTable {
			if currentTable != "" {
				schema += "\n" // Add a newline before a new table
			}
			schema += fmt.Sprintf("Table: %s\n", tableName.String)
			currentTable = tableName.String
		}

		//nolint:nestif
		if columnName.Valid {
			schema += fmt.Sprintf("  Column: %s, Type: %s", columnName.String, dataType.String)

			if columnDefault.Valid {
				schema += fmt.Sprintf(", Default: %s", columnDefault.String)
			}

			if isNullable.Valid {
				schema += fmt.Sprintf(", Nullable: %s", isNullable.String)
			}

			if constraintType.Valid {
				schema += fmt.Sprintf(", Constraint: %s (%s)", constraintType.String, constraintName.String)
				if foreignTableName.Valid && foreignColumnName.Valid {
					schema += fmt.Sprintf(", References: %s(%s)", foreignTableName.String, foreignColumnName.String)
				}
			}

			schema += "\n"
		}
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, fmt.Errorf("rows error: %w", rowsErr)
	}

	return &schema, nil
}
