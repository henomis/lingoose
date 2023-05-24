package sqlpipeline

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

	sqlRefinePromptTemplate = `
SQLQuery: {{.sql_query}}

The SQLResult has the following error: "{{.sql_result}}"
The SQLQuery you produced is syntactically incorrect. Please fix.`

	sqlFinalPromptTemplate = `
SQLQuery: {{.sql_query}}
SQLResult: {{.sql_result}}
Answer: `
)
