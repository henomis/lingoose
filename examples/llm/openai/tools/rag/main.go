package main

import (
	"context"
	"fmt"
	"os"

	openaiembedder "github.com/henomis/lingoose/embedder/openai"
	"github.com/henomis/lingoose/index"
	"github.com/henomis/lingoose/index/vectordb/jsondb"
	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/rag"
	"github.com/henomis/lingoose/thread"
	ragtool "github.com/henomis/lingoose/tools/rag"
	"github.com/henomis/lingoose/tools/serpapi"
)

func main() {

	rag := rag.New(
		index.New(
			jsondb.New().WithPersist("index.json"),
			openaiembedder.New(openaiembedder.AdaEmbeddingV2),
		),
	).WithChunkSize(1000).WithChunkOverlap(0)

	_, err := os.Stat("index.json")
	if os.IsNotExist(err) {
		err = rag.AddSources(context.Background(), "state_of_the_union.txt")
		if err != nil {
			panic(err)
		}
	}

	newStr := func(str string) *string {
		return &str
	}
	llm := openai.New().WithModel(openai.GPT4o).WithToolChoice(newStr("auto")).WithTools(
		ragtool.New(rag, "US covid vaccines"),
		serpapi.New(),
	)

	topics := []string{
		"how many covid vaccine doses US has donated to other countries",
		"who's the author of LinGoose github project",
	}

	for _, topic := range topics {
		t := thread.New().AddMessage(
			thread.NewUserMessage().AddContent(
				thread.NewTextContent("Please tell me " + topic + "."),
			),
		)

		llm.Generate(context.Background(), t)
		if t.LastMessage().Role == thread.RoleTool {
			llm.Generate(context.Background(), t)
		}

		fmt.Println(t)
	}

}
