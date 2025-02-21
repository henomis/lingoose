package main

import (
	"context"
	"fmt"
	"os"

	openaiembedder "github.com/rsest/lingoose/embedder/openai"
	"github.com/rsest/lingoose/index"
	"github.com/rsest/lingoose/index/vectordb/jsondb"
	"github.com/rsest/lingoose/llm/openai"
	"github.com/rsest/lingoose/rag"
	"github.com/rsest/lingoose/thread"
	ragtool "github.com/rsest/lingoose/tool/rag"
	"github.com/rsest/lingoose/tool/serpapi"
	"github.com/rsest/lingoose/tool/shell"
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
		shell.New(),
	)

	topics := []string{
		"how many covid vaccine doses US has donated to other countries.",
		"who's the author of LinGoose github project.",
		"which process is consuming the most memory.",
	}

	for _, topic := range topics {
		t := thread.New().AddMessage(
			thread.NewUserMessage().AddContent(
				thread.NewTextContent("Please tell me " + topic),
			),
		)

		llm.Generate(context.Background(), t)
		if t.LastMessage().Role == thread.RoleTool {
			llm.Generate(context.Background(), t)
		}

		fmt.Println(t)
	}

}
