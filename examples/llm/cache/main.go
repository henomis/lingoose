package main

import (
	"context"
	"fmt"

	openaiembedder "github.com/henomis/lingoose/embedder/openai"
	"github.com/henomis/lingoose/index"
	"github.com/henomis/lingoose/index/vectordb/jsondb"
	"github.com/henomis/lingoose/llm/cache"
	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/thread"
)

func main() {

	embedder := openaiembedder.New(openaiembedder.AdaEmbeddingV2)
	index := index.New(
		jsondb.New().WithPersist("index.json"),
		embedder,
	)
	llm := openai.New().WithCache(cache.New(embedder, index).WithTopK(3))

	questions := []string{
		"what's github",
		"can you explain what GitHub is",
		"can you tell me more about GitHub",
		"what is the purpose of GitHub",
	}

	for _, question := range questions {
		t := thread.NewThread().AddMessage(
			thread.NewUserMessage().AddContent(
				thread.NewTextContent(question),
			),
		)

		err := llm.Generate(context.Background(), t)
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Println(t)
	}
}
