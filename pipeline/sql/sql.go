package sqlpipeline

import (
	"database/sql"
	"fmt"

	"github.com/henomis/lingoose/decoder"
	"github.com/henomis/lingoose/pipeline"
	"github.com/henomis/lingoose/prompt"
	"github.com/henomis/lingoose/types"
)

type SQLDataSourceType string

const (
	SQLDataSourceSqlite SQLDataSourceType = "sqlite"
)

const (
	sqlQueryRegexExpr       = `(?s)SQLQuery: (.*)[;\n]*`
	sqlDefaultTopK          = 5
	sqlDefaultMaxIterations = 2
)

const (
	tubeQueryIndex         = 0
	tubeRefineQueryIndex   = 1
	tubeDescribeQueryIndex = 2
)

const (
	questionKey  = "question"
	sqlQueryKey  = "sql_query"
	sqlResultKey = "sql_result"
	tableInfoKey = "table_info"
	topKKey      = "top_k"
)

var (
	sqlStopSequence = []string{"SQLResult:"}
)

var sqlDataSourceTypePromptTemplate = map[SQLDataSourceType]string{
	SQLDataSourceSqlite: sqliteDataSourcePromptTemplate,
}

type llmWithStop interface {
	SetStop([]string)
}

type SqlDDLFn func() (string, error)

func New(llmEngine pipeline.LlmEngine, db *sql.DB, dataSourceType SQLDataSourceType, sqlDDLFn SqlDDLFn) (*pipeline.Pipeline, error) {

	memory := types.M{}

	if !llmImplementsSetStop(llmEngine) {
		return nil, fmt.Errorf("llmEngine does not implement SetStop([]string)")
	}

	sqlDDL, err := sqlDDLFn()
	if err != nil {
		return nil, err
	}

	llmEngine.(llmWithStop).SetStop(sqlStopSequence)

	dataSourcePromptTemplate := sqlDataSourceTypePromptTemplate[dataSourceType]

	sqlPrompt := prompt.NewPromptTemplate(dataSourcePromptTemplate + sqlPromptTemplate)
	queryLLM := pipeline.Llm{
		LlmEngine: llmEngine,
		LlmMode:   pipeline.LlmModeCompletion,
		Prompt:    sqlPrompt,
	}

	// ********** QUERY TUBE ************//
	query := pipeline.NewTube(queryLLM).WithDecoder(decoder.NewRegExDecoder(sqlQueryRegexExpr))

	preQueryCB := pipeline.PipelineCallback(func(input types.M) (types.M, error) {
		if q, ok := input[questionKey].(string); ok {
			memory[questionKey] = q
		}

		return preQueryCBFn(input, sqlDDL)
	})

	postQueryCB := pipeline.PipelineCallback(func(output types.M) (types.M, error) {
		return postQueryCBFn(output, db, sqlDDL, memory)
	})
	// ********** END QUERY TUBE ************//

	// ********** REFINE QUERY TUBE ************//

	refinePrompt := prompt.NewPromptTemplate(dataSourcePromptTemplate + sqlPromptTemplate + sqlRefinePromptTemplate)
	refineLLM := pipeline.Llm{
		LlmEngine: llmEngine,
		LlmMode:   pipeline.LlmModeCompletion,
		Prompt:    refinePrompt,
	}

	refine := pipeline.NewTube(refineLLM).WithDecoder(decoder.NewRegExDecoder(sqlQueryRegexExpr))

	preRefineCB := pipeline.PipelineCallback(func(input types.M) (types.M, error) {
		return preRefineCBFn(input, sqlDDL, memory)
	})

	postRefineCBFn := pipeline.PipelineCallback(func(output types.M) (types.M, error) {
		return postRefineCBFn(output, db, sqlDDL, memory)
	})

	// ********** END REFINE QUERY TUBE ************//

	// ********** DESCRIBE ************//

	describePrompt := prompt.NewPromptTemplate(dataSourcePromptTemplate + sqlPromptTemplate + sqlFinalPromptTemplate)

	describeLLM := pipeline.Llm{
		LlmEngine: llmEngine,
		LlmMode:   pipeline.LlmModeCompletion,
		Prompt:    describePrompt,
	}

	describe := pipeline.NewTube(describeLLM)

	preDescribeCB := pipeline.PipelineCallback(func(input types.M) (types.M, error) {
		return preDescribeCBFn(input, sqlDDL, memory)
	})

	postDescribeCB := pipeline.PipelineCallback(func(output types.M) (types.M, error) {
		output[sqlQueryKey] = memory[sqlQueryKey]
		output[sqlResultKey] = memory[sqlResultKey]
		return output, nil
	})

	// ********** END DESCRIBE ************//

	sqlPipeline := pipeline.New(query, refine, describe).WithPreCallbacks(preQueryCB, preRefineCB, preDescribeCB).WithPostCallbacks(postQueryCB, postRefineCBFn, postDescribeCB)

	return sqlPipeline, nil

}

func llmImplementsSetStop(llmEngine pipeline.LlmEngine) bool {
	var i interface{} = llmEngine
	_, ok := i.(llmWithStop)
	return ok
}

func preQueryCBFn(input types.M, sqlDDL string) (types.M, error) {
	input[tableInfoKey] = sqlDDL
	input[topKKey] = sqlDefaultTopK
	return input, nil
}

func preRefineCBFn(input types.M, sqlDDL string, memory types.M) (types.M, error) {
	input[tableInfoKey] = sqlDDL
	input[topKKey] = sqlDefaultTopK
	input[questionKey] = memory[questionKey]
	return input, nil
}

func preDescribeCBFn(input types.M, sqlDDL string, memory types.M) (types.M, error) {
	input[questionKey] = memory[questionKey]
	input[sqlQueryKey] = memory[sqlQueryKey]
	input[tableInfoKey] = sqlDDL
	input[topKKey] = sqlDefaultTopK
	return input, nil
}

func postQueryCBFn(output types.M, db *sql.DB, sqlDDL string, memory types.M) (types.M, error) {

	sqlQueryMatches, ok := output[types.DefaultOutputKey].([]string)
	if !ok || len(sqlQueryMatches) != 1 {
		return output, nil
	}

	sqlQuery := sqlQueryMatches[0]

	output[sqlQueryKey] = sqlQuery
	memory[sqlQueryKey] = sqlQuery

	sqlResult, err := getSqlResult(db, sqlQuery)

	memory[sqlResultKey] = sqlResult

	if err == nil {
		output[sqlResultKey] = sqlResult
		pipeline.SetNextTube(output, tubeDescribeQueryIndex)
	} else {
		output[sqlResultKey] = err.Error()
		pipeline.SetNextTube(output, tubeRefineQueryIndex)
	}

	return output, nil
}

func postRefineCBFn(output types.M, db *sql.DB, sqlDDL string, memory types.M) (types.M, error) {

	sqlQueryMatches, ok := output[types.DefaultOutputKey].([]string)
	if !ok || len(sqlQueryMatches) != 1 {
		return output, nil
	}
	sqlQuery := sqlQueryMatches[0]

	sqlResult, err := getSqlResult(db, sqlQuery)

	output[sqlResultKey] = sqlResult
	output[sqlQueryKey] = sqlQuery
	memory[sqlQueryKey] = sqlQuery
	memory[sqlResultKey] = sqlResult

	if err == nil {
		pipeline.SetNextTube(output, tubeDescribeQueryIndex)
	} else {
		output[types.DefaultOutputKey] = err.Error()
		pipeline.SetNextTubeExit(output)
	}

	return output, nil
}

func getSqlResult(db *sql.DB, query string) (string, error) {

	rows, err := db.Query(query)
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
