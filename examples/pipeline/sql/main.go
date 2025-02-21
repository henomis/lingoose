package main

import (
	"context"
	"fmt"

	sqlpipeline "github.com/rsest/lingoose/legacy/pipeline/sql"
	"github.com/rsest/lingoose/llm/openai"
	"github.com/rsest/lingoose/types"
	// enable sqlite3 driver
	// _ "github.com/mattn/go-sqlite3"
	// enable mysql driver
	// _ "github.com/go-sql-driver/mysql"
)

// mysql https://raw.githubusercontent.com/lerocha/chinook-database/master/ChinookDatabase/DataSources/Chinook_MySql.sql
// sqlite https://raw.githubusercontent.com/lerocha/chinook-database/master/ChinookDatabase/DataSources/Chinook_Sqlite.sqlite

func main() {

	// SQLite
	s, err := sqlpipeline.New(
		openai.NewCompletion().WithMaxTokens(1000).WithVerbose(true),
		sqlpipeline.DataSourceSqlite,
		"/tmp/Chinook_Sqlite.sqlite",
	)

	// MySQL
	// s, err := sqlpipeline.New(
	// 	openai.NewCompletion().WithMaxTokens(1000).WithVerbose(true),
	// 	sqlpipeline.DataSourceMySQL,
	// 	"root:password@tcp(localhost:3306)/Chinook",
	// )

	if err != nil {
		panic(err)
	}

	output, err := s.Run(context.Background(), types.M{"question": "list the top 3 playlists and count how many tracks they have."})
	if err != nil {
		panic(err)
	}

	fmt.Println(output)

}
