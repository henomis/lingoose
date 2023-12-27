package main

import (
	"context"
	"fmt"
	"os"

	"github.com/henomis/lingoose/assistant"
	openaiembedder "github.com/henomis/lingoose/embedder/openai"
	"github.com/henomis/lingoose/index"
	"github.com/henomis/lingoose/index/vectordb/jsondb"
	"github.com/henomis/lingoose/llm/cache"
	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/rag"
	"github.com/henomis/lingoose/thread"
)

// download https://raw.githubusercontent.com/hwchase17/chat-your-data/master/state_of_the_union.txt

func main() {
	openaiEmbedder := openaiembedder.New(openaiembedder.AdaEmbeddingV2)
	cache := cache.New(
		openaiEmbedder,
		index.New(
			jsondb.New().WithPersist("cache.json"),
			openaiEmbedder,
		),
	)

	r := rag.New(
		index.New(
			jsondb.New().WithPersist("db.json"),
			openaiEmbedder,
		),
	)

	_, err := os.Stat("db.json")
	if os.IsNotExist(err) {
		r.AddFiles(context.Background(), "state_of_the_union.txt")
	}

	a := assistant.New(
		openai.New().WithCache(cache),
	).WithRAG(r)

	questions := []string{
		"what is NATO?",
		"what is the purpose of NATO?",
		"what is the purpose of the NATO Alliance?",
		"what is the meaning of NATO?",
	}

	for _, question := range questions {
		a = a.WithThread(thread.NewThread())
		err := a.Run(context.Background(), question)
		if err != nil {
			panic(err)
		}

		fmt.Println("----")
		fmt.Println(a.Thread())
		fmt.Println("----")
	}
}
