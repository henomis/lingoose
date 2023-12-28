package main

import (
	"context"
	"fmt"
	"os"

	"github.com/henomis/lingoose/assistant"
	openaiembedder "github.com/henomis/lingoose/embedder/openai"
	"github.com/henomis/lingoose/index"
	"github.com/henomis/lingoose/index/vectordb/jsondb"
	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/rag"
	"github.com/henomis/lingoose/thread"
)

// download https://raw.githubusercontent.com/hwchase17/chat-your-data/master/state_of_the_union.txt

func main() {
	r := rag.NewRAGFusion(
		index.New(
			jsondb.New().WithPersist("db.json"),
			openaiembedder.New(openaiembedder.AdaEmbeddingV2),
		),
		openai.New().WithTemperature(0),
	)

	_, err := os.Stat("db.json")
	if os.IsNotExist(err) {
		r.AddFiles(context.Background(), "state_of_the_union.txt")
	}

	a := assistant.New(
		openai.New().WithTemperature(0),
	).WithRAG(r)

	a = a.WithThread(thread.NewThread())
	err = a.Run(context.Background(), "what is the purpose of NATO?")
	if err != nil {
		panic(err)
	}

	fmt.Println("----")
	fmt.Println(a.Thread())
	fmt.Println("----")
}
