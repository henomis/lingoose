package sqlpipeline

import (
	"database/sql"

	"github.com/henomis/lingoose/pipeline"
	"github.com/henomis/lingoose/prompt"
)

type SQLDataSourceType string

const (
	SQLDataSourceSqlite SQLDataSourceType = "sqlite"
)

var sqlDataSourceTypePromptTemplate = map[SQLDataSourceType]string{
	SQLDataSourceSqlite: sqliteDataSourcePromptTemplate,
}

type SqlDDLFn func() (string, error)

func New(llmEngine pipeline.LlmEngine, db *sql.DB, dataSourceType SQLDataSourceType, sqlDDLFn SqlDDLFn) *pipeline.Pipeline {

	dataSourcePromptTemplate := sqlDataSourceTypePromptTemplate[dataSourceType]

	sqlPrompt := prompt.NewPromptTemplate(dataSourcePromptTemplate + sqlPromptTemplate)
	llm := pipeline.Llm{
		LlmEngine: llmEngine,
		LlmMode:   pipeline.LlmModeCompletion,
		Prompt:    sqlPrompt,
	}

	tube := pipeline.NewTube(llm)

	return &SQLTube{
		tube:           tube,
		db:             db,
		dataSourceType: dataSourceType,
		getDDLFn:       sqlDDLFn,
		topK:           sqlDefaultTopK,
		maxIterations:  sqlDefaultMaxIterations,
	}

}
