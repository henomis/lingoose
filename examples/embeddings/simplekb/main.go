package main

import (
	"context"

	openaiembedder "github.com/rsest/lingoose/embedder/openai"
	"github.com/rsest/lingoose/index"
	"github.com/rsest/lingoose/index/option"
	"github.com/rsest/lingoose/index/vectordb/jsondb"
	qapipeline "github.com/rsest/lingoose/legacy/pipeline/qa"
	"github.com/rsest/lingoose/llm/openai"
	"github.com/rsest/lingoose/loader"
	"github.com/rsest/lingoose/textsplitter"
)

func main() {
	docs, _ := loader.NewPDFToTextLoader("./kb").WithTextSplitter(textsplitter.NewRecursiveCharacterTextSplitter(2000, 200)).Load(context.Background())
	index := index.New(jsondb.New(), openaiembedder.New(openaiembedder.AdaEmbeddingV2)).WithIncludeContents(true)
	index.LoadFromDocuments(context.Background(), docs)
	qapipeline.New(openai.NewChat().WithVerbose(true)).WithIndex(index).Query(context.Background(), "What is the NATO purpose?", option.WithTopK(1))
}
