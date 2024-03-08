package main

import (
	"context"
	"fmt"
	"os"

	openaiembedder "github.com/henomis/lingoose/embedder/openai"
	"github.com/henomis/lingoose/index"
	"github.com/henomis/lingoose/index/vectordb/jsondb"
	"github.com/henomis/lingoose/linglet/qa"
	"github.com/henomis/lingoose/llm/openai"
)

// download https://raw.githubusercontent.com/hwchase17/chat-your-data/master/state_of_the_union.txt

func main() {
	qa := qa.New(
		openai.New().WithTemperature(0),
		index.New(
			jsondb.New().WithPersist("db.json"),
			openaiembedder.New(openaiembedder.AdaEmbeddingV2),
		),
	)

	_, err := os.Stat("db.json")
	if os.IsNotExist(err) {
		err = qa.AddSource(context.Background(), "state_of_the_union.txt")
		if err != nil {
			panic(err)
		}
	}

	response, err := qa.Run(context.Background(), "What is the NATO purpose?")
	if err != nil {
		panic(err)
	}

	fmt.Println(response)
}
