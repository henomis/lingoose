package main

import (
	"context"

	openaiembedder "github.com/henomis/lingoose/embedder/openai"
	"github.com/henomis/lingoose/index"
	"github.com/henomis/lingoose/index/option"
	"github.com/henomis/lingoose/index/vectordb/jsondb"
	qapipeline "github.com/henomis/lingoose/legacy/pipeline/qa"
	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/loader"
	"github.com/henomis/lingoose/textsplitter"
)

func main() {
	docs, _ := loader.NewPDFToTextLoader("./kb").WithTextSplitter(textsplitter.NewRecursiveCharacterTextSplitter(2000, 200)).Load(context.Background())
	index := index.New(jsondb.New(), openaiembedder.New(openaiembedder.AdaEmbeddingV2)).WithIncludeContents(true)
	index.LoadFromDocuments(context.Background(), docs)
	qapipeline.New(openai.NewChat().WithVerbose(true)).WithIndex(index).Query(context.Background(), "What is the NATO purpose?", option.WithTopK(1))
}
