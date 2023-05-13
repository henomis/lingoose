package main

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"

	_ "github.com/mattn/go-sqlite3"

	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/pipeline"
	"github.com/henomis/lingoose/types"
)

func main() {

	sqliteDB := "/tmp/Chinook_Sqlite.sqlite"

	db, err := sql.Open("sqlite3", sqliteDB)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	s := pipeline.NewSQLTube(
		openai.NewCompletion().WithMaxTokens(1000),
		db,
		pipeline.SQLDataSourceSqlite,
		func() (string, error) {
			output, err := exec.Command("sqlite3", sqliteDB, ".schema").Output()
			if err != nil {
				return "", err
			}

			return string(output), nil
		},
	)

	for {

		fmt.Printf("> ")
		reader := bufio.NewReader(os.Stdin)
		query, _ := reader.ReadString('\n')

		query = query[:len(query)-1]

		if query == "quit" {
			break
		}

		output, err := s.Run(context.Background(), query)
		if err != nil {
			panic(err)
		}
		_ = output

		results := output["output"].(types.M)

		fmt.Println(results["answer"])
		fmt.Println()
		fmt.Println(results["query"])
		fmt.Println()
		fmt.Println(results["result"])
	}

}
