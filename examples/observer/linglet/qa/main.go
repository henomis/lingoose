package main

import (
	"context"
	"fmt"
	"os"

	openaiembedder "github.com/rsest/lingoose/embedder/openai"
	"github.com/rsest/lingoose/index"
	"github.com/rsest/lingoose/index/vectordb/jsondb"
	"github.com/rsest/lingoose/linglet/qa"
	"github.com/rsest/lingoose/llm/openai"
	"github.com/rsest/lingoose/observer"
	"github.com/rsest/lingoose/observer/langfuse"
)

// download https://raw.githubusercontent.com/hwchase17/chat-your-data/master/state_of_the_union.txt

func main() {
	ctx := context.Background()

	o := langfuse.New(ctx)
	trace, err := o.Trace(&observer.Trace{Name: "state of the union"})
	if err != nil {
		panic(err)
	}

	ctx = observer.ContextWithObserverInstance(ctx, o)
	ctx = observer.ContextWithTraceID(ctx, trace.ID)

	qa := qa.New(
		openai.New().WithTemperature(0),
		index.New(
			jsondb.New().WithPersist("db.json"),
			openaiembedder.New(openaiembedder.AdaEmbeddingV2),
		),
	)

	_, err = os.Stat("db.json")
	if os.IsNotExist(err) {
		err = qa.AddSource(ctx, "state_of_the_union.txt")
		if err != nil {
			panic(err)
		}
	}

	response, err := qa.Run(ctx, "What is the NATO purpose?")
	if err != nil {
		panic(err)
	}

	fmt.Println(response)
}
