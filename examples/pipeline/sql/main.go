package main

import (
	"context"
	"database/sql"
	"fmt"
	"os/exec"

	"github.com/henomis/lingoose/llm/openai"
	sqlpipeline "github.com/henomis/lingoose/pipeline/sql"
	"github.com/henomis/lingoose/types"

	_ "github.com/mattn/go-sqlite3"
)

func main() {

	// https://github.com/lerocha/chinook-database/raw/master/ChinookDatabase/DataSources/Chinook_Sqlite.sqlite
	sqliteDB := "/tmp/Chinook_Sqlite.sqlite"

	db, err := sql.Open("sqlite3", sqliteDB)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	s, err := sqlpipeline.New(
		openai.NewCompletion().WithMaxTokens(1000).WithVerbose(true),
		db,
		sqlpipeline.SQLDataSourceSqlite,
		func() (string, error) {
			output, err := exec.Command("sqlite3", sqliteDB, ".schema").Output()
			if err != nil {
				return "", err
			}

			return string(output), nil
		},
	)
	if err != nil {
		panic(err)
	}

	output, err := s.Run(context.Background(), types.M{"question": "list the top 3 playlists and count how many tracks they have."})
	if err != nil {
		panic(err)
	}

	fmt.Println(output)

}
