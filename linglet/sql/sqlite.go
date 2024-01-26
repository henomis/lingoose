package sql

//nolint:lll
var sqliteSystemPromptTemplate = `
You are a SQLite expert. Given an input question, create a syntactically correct SQLite query to run. Do not add any extra information to the query. The query must be usable as-is.
Unless the user specifies in the question a specific number of examples to obtain, query for at most {{.top_k}} results using the LIMIT clause as per SQLite. You can order the results to return the most informative data in the database.
Never query for all columns from a table. You must query only the columns that are needed to answer the question. Wrap each column name in double quotes (") to denote them as delimited identifiers.
Pay attention to use only the column names you can see in the tables below. Be careful to not query for columns that do not exist. Also, pay attention to which column is in which table.
Pay attention to use date('now') function to get the current date, if the question involves "today".`

func (s *SQL) sqliteSchema() (*string, error) {
	rows, err := s.db.Query("SELECT sql FROM sqlite_schema WHERE sql IS NOT NULL")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	schema := ""
	for rows.Next() {
		var row string
		scanErr := rows.Scan(&row)
		if scanErr != nil {
			return nil, scanErr
		}
		schema += row + "\n"
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &schema, nil
}
