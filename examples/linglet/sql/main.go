package main

import (
	"context"
	"database/sql"
	"fmt"

	lingletsql "github.com/henomis/lingoose/linglet/sql"
	"github.com/henomis/lingoose/llm/openai"
	// enable sqlite3 driver
	// _ "github.com/mattn/go-sqlite3"
)

// sqlite https://raw.githubusercontent.com/lerocha/chinook-database/master/ChinookDatabase/DataSources/Chinook_Sqlite.sqlite

func main() {
	db, err := sql.Open("sqlite3", "Chinook_Sqlite.sqlite")
	if err != nil {
		panic(err)
	}

	lingletSQL := lingletsql.New(
		openai.New().WithMaxTokens(2000).WithTemperature(0).WithModel(openai.GPT3Dot5Turbo16K0613),
		db,
	)

	result, err := lingletSQL.Run(
		context.Background(),
		"list the top 3 albums that are most frequently present in playlists.",
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("SQL Query\n-----\n%s\n\n", result.SQLQuery)
	fmt.Printf("Answer\n-------\n%s\n", result.Answer)
}
