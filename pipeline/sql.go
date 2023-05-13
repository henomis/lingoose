package pipeline

import (
	"context"
	"database/sql"
	"regexp"
	"strings"

	"github.com/henomis/lingoose/prompt"
	"github.com/henomis/lingoose/types"
)

type SQLDataSourceType string

const (
	SQLDataSourceSqlite SQLDataSourceType = "sqlite"
)

var sqlDataSourceTypePromptTemplate = map[SQLDataSourceType]string{
	SQLDataSourceSqlite: `
You are a SQLite expert. Given an input question, first create a syntactically correct SQLite query to run, then look at the results of the query and return the answer to the input question.
Unless the user specifies in the question a specific number of examples to obtain, query for at most {{.top_k}} results using the LIMIT clause as per SQLite. You can order the results to return the most informative data in the database.
Never query for all columns from a table. You must query only the columns that are needed to answer the question. Wrap each column name in double quotes (") to denote them as delimited identifiers.
Pay attention to use only the column names you can see in the tables below. Be careful to not query for columns that do not exist. Also, pay attention to which column is in which table.
Pay attention to use date('now') function to get the current date, if the question involves "today".`,
}

const (
	sqlPromptTemplate = `

Use the following format:

Question: Question here
SQLQuery: SQL Query to run.
SQLResult: Result of the SQLQuery
Answer: Final answer here
	
Use the following table schema info to create your SQL query:
{{.table_info}}

Question: {{.question}}`

	sqlRefinePromptTemplate = `{{.ai_sql_query}}

The SQLResult has the following error: "{{.sql_error}}"
The SQLQuery you produced is syntactically incorrect. Please fix.`

	sqlFinalPromptTemplate = `
SQLQuery: {{.ai_sql_query}}
SQLResult: {{.sql_result}}
Answer: `
)

const (
	sqlQueryRegexExpr       = `(?s)SQLQuery: (.*);`
	sqlDefaultTopK          = 5
	sqlDefaultMaxIterations = 2
)

type SqlDDLFn func() (string, error)

type SQLTube struct {
	tube           *Tube
	db             *sql.DB
	dataSourceType SQLDataSourceType
	getDDLFn       SqlDDLFn
	topK           int
	maxIterations  int
}

func NewSQLTube(llmEngine LlmEngine, db *sql.DB, dataSourceType SQLDataSourceType, sqlDDLFn SqlDDLFn) *SQLTube {

	// TODO: check if llmEngine can call WithStop([]string). If so, add "SQLResult:" to the list of stop words.
	//else add reference to simulateStop to the object and call after completion.

	dataSourcePromptTemplate := sqlDataSourceTypePromptTemplate[dataSourceType]

	sqlPrompt := prompt.NewPromptTemplate(dataSourcePromptTemplate + sqlPromptTemplate)
	llm := Llm{
		LlmEngine: llmEngine,
		LlmMode:   LlmModeCompletion,
		Prompt:    sqlPrompt,
	}

	tube := NewTube(llm)

	return &SQLTube{
		tube:           tube,
		db:             db,
		dataSourceType: dataSourceType,
		getDDLFn:       sqlDDLFn,
		topK:           sqlDefaultTopK,
		maxIterations:  sqlDefaultMaxIterations,
	}
}

func (s *SQLTube) WithPrompt(prompt Prompt) *SQLTube {
	s.tube.llm.Prompt = prompt
	return s
}

func (s *SQLTube) WithTopK(topK int) *SQLTube {
	s.topK = topK
	return s
}

func (s *SQLTube) WithMaxIterations(maxIterations int) *SQLTube {
	s.maxIterations = maxIterations
	return s
}

func (s *SQLTube) Run(ctx context.Context, query string) (types.M, error) {

	var output types.M
	ddl, err := s.getDDLFn()
	if err != nil {
		return nil, err
	}

	sqlQueryRegexExpr, err := regexp.Compile(sqlQueryRegexExpr)
	if err != nil {
		return nil, err
	}

	iteration := 1
	aiSqlQuery := ""
	sqlResult := ""
	for {

		output, err = s.tube.Run(
			ctx,
			types.M{
				"table_info": ddl,
				"question":   query,
				"top_k":      s.topK,
			},
		)
		if err != nil {
			return nil, err
		}

		aiSqlQuery = simulateLlmStop(output[types.DefaultOutputKey].(string))

		sqlQuery := sqlQueryRegexExpr.FindStringSubmatch(aiSqlQuery)

		if len(sqlQuery) != 2 {
			return nil, ErrLLMExecution
		}

		sqlResult, err = s.getSqlResult(sqlQuery[1])

		if err == nil {
			output[types.DefaultOutputKey] = sqlResult
			break
		}

		if iteration == s.maxIterations {
			return nil, err
		}

		refinePrompt := prompt.NewPromptTemplate(sqlRefinePromptTemplate)
		refinePrompt.Format(
			types.M{
				"ai_sql_query": aiSqlQuery,
				"sql_error":    err.Error(),
			},
		)
		s.rebuildPrompt(refinePrompt.String())

		iteration++
	}

	finalPrompt := prompt.NewPromptTemplate(sqlFinalPromptTemplate)
	finalPrompt.Format(
		types.M{
			"ai_sql_query": aiSqlQuery,
			"sql_result":   sqlResult,
		},
	)
	s.rebuildPrompt(finalPrompt.String())

	output, err = s.tube.Run(
		ctx,
		types.M{
			"table_info": ddl,
			"question":   query,
			"top_k":      s.topK,
		},
	)
	if err != nil {
		return nil, err
	}

	output = types.M{
		types.DefaultOutputKey: types.M{
			"description": output[types.DefaultOutputKey],
			"sql_query":   aiSqlQuery,
			"sql_result":  sqlResult,
		},
	}

	return output, nil
}

func (s *SQLTube) getSqlResult(query string) (string, error) {

	rows, err := s.db.Query(query)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	content := ""
	columns, err := rows.Columns()
	if err != nil {
		return "", err
	}

	for _, col := range columns {
		if content != "" {
			content += "|" + col
		} else {
			content += col
		}
	}

	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return "", err
		}

		row := ""
		for _, col := range values {

			if row != "" {
				row += "|" + string(col)
			} else {
				row += string(col)
			}

		}

		content += "\n" + row
	}

	return content, nil
}

func (s *SQLTube) rebuildPrompt(refinePrompt string) {
	dataSourcePromptTemplate := sqlDataSourceTypePromptTemplate[s.dataSourceType]
	s.tube.llm.Prompt = prompt.NewPromptTemplate(dataSourcePromptTemplate + sqlPromptTemplate + refinePrompt)
}

func simulateLlmStop(input string) string {
	index := strings.Index(input, "SQLResult:")
	if index == -1 {
		return input
	}
	return strings.TrimSpace(input[:index])
}
