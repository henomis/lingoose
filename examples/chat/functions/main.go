package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/henomis/lingoose/chat"
	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/prompt"
	// uncomment this line to use the sqlite3 driver
	// _ "github.com/mattn/go-sqlite3"
)

const (
	database = "Chinook_Sqlite.sqlite"
)

// Downdload the Chinook database:
// wget https://github.com/lerocha/chinook-database/raw/master/ChinookDatabase/DataSources/Chinook_Sqlite.sqlite

func main() {

	databaseSchemaString, err := getDatabaseSchema(database)
	if err != nil {
		panic(err)
	}

	chat := chat.New(
		chat.PromptMessage{
			Type:   chat.MessageTypeSystem,
			Prompt: prompt.New("Answer user questions by generating SQL queries against the Chinook Music Database."),
		},
		chat.PromptMessage{
			Type:   chat.MessageTypeUser,
			Prompt: prompt.New("Hi, who are the top 5 artists by number of tracks?"),
		},
	).WithOption(
		func(o *chat.Options) {
			o.OpenAIFunctionsEnabled = true
			o.OpenAIFunctionsMaxIterations = 1
		},
	)

	llmOpenAI := openai.New(openai.GPT3Dot5Turbo0613, openai.DefaultOpenAITemperature, openai.DefaultOpenAIMaxTokens, true)

	llmOpenAI.BindFunction(
		"ask_database",
		"Use this function to answer user questions about music. Output should be a fully formed SQL query.",
		askDatabase,
		func(m map[string]interface{}) error {
			m["properties"].(map[string]interface{})["query"].(map[string]interface{})["description"] = fmt.Sprintf(
				"SQL query extracting info to answer the user's question.\nSQL should be written using this database schema:\n%s\nThe query should be returned in plain text, not in JSON.",
				databaseSchemaString,
			)
			return nil
		},
	)

	response, err := llmOpenAI.Chat(context.Background(), chat)
	if err != nil {
		panic(err)
	}

	fmt.Printf("\n%#v", response)

}

type Query struct {
	Query string `json:"query"`
}

func askDatabase(query Query) ([]map[string]interface{}, error) {

	conn, err := sql.Open("sqlite3", database)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer conn.Close()

	rows, err := conn.Query(query.Query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	results := make([]map[string]interface{}, 0)
	for rows.Next() {
		err := rows.Scan(valuePtrs...)
		if err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}
		results = append(results, row)
	}

	return results, nil
}
