package main

import (
	"context"

	openaiembedder "github.com/henomis/lingoose/embedder/openai"
	"github.com/henomis/lingoose/index"
	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/loader"
	"github.com/henomis/lingoose/pipeline"
	"github.com/henomis/lingoose/textsplitter"
)

func main() {
	query := "Who is Mario?"
	topk := 3
	docs, _ := loader.NewPDFToTextLoader("./kb").WithTextSplitter(textsplitter.NewRecursiveCharacterTextSplitter(2000, 200)).Load()
	openaiEmbedder := openaiembedder.New(openaiembedder.AdaEmbeddingV2)
	index.NewSimpleVectorIndex("db", ".", openaiEmbedder).LoadFromDocuments(context.Background(), docs)
	similarities, _ := index.NewSimpleVectorIndex("db", ".", openaiEmbedder).SimilaritySearch(context.Background(), query, &topk)
	pipeline.NewQATube(openai.NewChat().WithVerbose(true)).Run(context.Background(), query, similarities)
}
