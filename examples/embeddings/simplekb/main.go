package main

import (
	"context"

	openaiembedder "github.com/henomis/lingoose/embedder/openai"
	indexoption "github.com/henomis/lingoose/index/option"
	simplevectorindex "github.com/henomis/lingoose/index/simpleVectorIndex"
	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/loader"
	qapipeline "github.com/henomis/lingoose/pipeline/qa"
	"github.com/henomis/lingoose/textsplitter"
)

func main() {
	query := "What is the NATO purpose?"
	docs, _ := loader.NewPDFToTextLoader("./kb").WithTextSplitter(textsplitter.NewRecursiveCharacterTextSplitter(2000, 200)).Load(context.Background())
	openaiEmbedder := openaiembedder.New(openaiembedder.AdaEmbeddingV2)
	simplevectorindex.New("db", ".", openaiEmbedder).LoadFromDocuments(context.Background(), docs)
	results, _ := simplevectorindex.New("db", ".", openaiEmbedder).Query(context.Background(), query, indexoption.WithTopK(3))
	qapipeline.New(openai.NewChat().WithVerbose(true)).Run(context.Background(), query, results.ToDocuments())
}
