package sql

const (
	sqlPromptTemplate = `	
Use the following table schema info to create your SQL query:
{{.schema}}

Question: {{.question}}
SQLQuery: `

	sqlRefinePromptTemplate = `
SQLQuery: {{.sql_query}}

The SQLResult has the following error: "{{.sql_error}}"
The SQLQuery you produced is syntactically incorrect. Please fix.
Answer with the SQL query, do not add any extra information, the query must be usable as-is.
SQLQuery: `

	sqlFinalPromptTemplate = `
You are an helpful assistant and your goal is to answer to the user's question.
Your answer must be comprehensible by non-tech users.
You will be provided with additional context information as result of an SQL query.

Question: {{.question}}
SQL result: {{.sql_result}}
Answer: `
)
