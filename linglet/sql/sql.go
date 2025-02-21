package sql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/rsest/lingoose/assistant"
	"github.com/rsest/lingoose/thread"
	"github.com/rsest/lingoose/types"
)

type CallbackFn func(t *thread.Thread)

type LLM interface {
	Generate(context.Context, *thread.Thread) error
}

const (
	defaultTopK = 5
)

type SQL struct {
	db         *sql.DB
	topk       int
	assistant  *assistant.Assistant
	callbackFn CallbackFn
}

type Result struct {
	SQLQuery string
	Answer   string
}

func New(llm LLM, db *sql.DB) *SQL {
	return &SQL{
		db:        db,
		topk:      defaultTopK,
		assistant: assistant.New(llm),
	}
}

func (s *SQL) WithTopK(topk int) *SQL {
	s.topk = topk
	return s
}

func (s *SQL) WithCallback(callbackFn CallbackFn) *SQL {
	s.callbackFn = callbackFn
	return s
}

func (s *SQL) schema() (*string, error) {
	driverType := fmt.Sprintf("%T", s.db.Driver())
	if strings.Contains(driverType, "sqlite") {
		return s.sqliteSchema()
	}

	return nil, fmt.Errorf("unsupported database driver %s", driverType)
}

func (s *SQL) systemPrompt() (*string, error) {
	driverType := fmt.Sprintf("%T", s.db.Driver())
	if strings.Contains(driverType, "sqlite") {
		return &sqliteSystemPromptTemplate, nil
	}

	return nil, fmt.Errorf("unsupported database driver %s", driverType)
}

func (s *SQL) Run(ctx context.Context, question string) (*Result, error) {
	sqlQuery, err := s.generateSQLQuery(ctx, question)
	if err != nil {
		return nil, err
	}

	sqlResult, err := s.executeSQLQuery(*sqlQuery)
	if err != nil {
		refinedSQLResult, refineErr := s.generateRefinedSQLQuery(
			ctx,
			question,
			*sqlQuery,
			err,
		)
		if refineErr != nil {
			return nil, refineErr
		}

		if s.callbackFn != nil {
			s.callbackFn(s.assistant.Thread())
		}

		sqlResult, refineErr = s.executeSQLQuery(*refinedSQLResult)
		if refineErr != nil {
			return nil, refineErr
		}
		sqlQuery = refinedSQLResult
	}

	if s.callbackFn != nil {
		s.callbackFn(s.assistant.Thread())
	}

	answer, err := s.generateAnswer(ctx, sqlResult, question)
	if err != nil {
		return nil, err
	}

	if s.callbackFn != nil {
		s.callbackFn(s.assistant.Thread())
	}

	return &Result{
		SQLQuery: *sqlQuery,
		Answer:   *answer,
	}, nil
}

func (s *SQL) generateSQLQuery(ctx context.Context, question string) (*string, error) {
	systemPrompt, err := s.systemPrompt()
	if err != nil {
		return nil, err
	}
	schema, err := s.schema()
	if err != nil {
		return nil, err
	}

	s.assistant.Thread().ClearMessages().AddMessage(
		thread.NewSystemMessage().AddContent(
			thread.NewTextContent(*systemPrompt).Format(
				types.M{
					"top_k": s.topk,
				},
			),
		),
	).AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent(sqlPromptTemplate).Format(
				types.M{
					"schema":   schema,
					"question": question,
				},
			),
		),
	)

	err = s.assistant.Run(ctx)
	if err != nil {
		return nil, err
	}

	if content, ok := s.assistant.Thread().LastMessage().Contents[0].Data.(string); ok {
		return &content, nil
	}

	return nil, fmt.Errorf("no content")
}

func (s *SQL) generateRefinedSQLQuery(ctx context.Context, question, sqlQuery string, sqlError error) (*string, error) {
	systemPrompt, err := s.systemPrompt()
	if err != nil {
		return nil, err
	}
	schema, err := s.schema()
	if err != nil {
		return nil, err
	}

	s.assistant.Thread().ClearMessages().AddMessage(
		thread.NewSystemMessage().AddContent(
			thread.NewTextContent(*systemPrompt).Format(
				types.M{
					"top_k": s.topk,
				},
			),
		),
	).AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent(sqlPromptTemplate).Format(
				types.M{
					"schema":   schema,
					"question": question,
				},
			),
		),
	).AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent(sqlRefinePromptTemplate).Format(
				types.M{
					"sql_query": sqlQuery,
					"sql_error": sqlError.Error(),
				},
			),
		),
	)

	err = s.assistant.Run(ctx)
	if err != nil {
		return nil, err
	}

	if content, ok := s.assistant.Thread().LastMessage().Contents[0].Data.(string); ok {
		return &content, nil
	}

	return nil, fmt.Errorf("no content")
}

func (s *SQL) generateAnswer(ctx context.Context, sqlResult, question string) (*string, error) {
	s.assistant.Thread().ClearMessages().AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent(sqlFinalPromptTemplate).Format(
				types.M{
					"question":   question,
					"sql_result": sqlResult,
				},
			),
		),
	)

	err := s.assistant.Run(ctx)
	if err != nil {
		return nil, err
	}

	if content, ok := s.assistant.Thread().LastMessage().Contents[0].Data.(string); ok {
		return &content, nil
	}

	return nil, fmt.Errorf("no content")
}

func (s *SQL) executeSQLQuery(sqlQuery string) (string, error) {
	rows, err := s.db.Query(sqlQuery)
	if err != nil {
		return "", err
	}
	if err = rows.Err(); err != nil {
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
