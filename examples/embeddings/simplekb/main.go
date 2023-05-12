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
	docs, _ := loader.NewPDFToTextLoader("./kb").WithTextSplitter(textsplitter.NewRecursiveCharacterTextSplitter(2000, 200)).Load(context.Background())
	openaiEmbedder := openaiembedder.New(openaiembedder.AdaEmbeddingV2)
	index.NewSimpleVectorIndex("db", ".", openaiEmbedder).LoadFromDocuments(context.Background(), docs)
	similarities, _ := index.NewSimpleVectorIndex("db", ".", openaiEmbedder).SimilaritySearch(context.Background(), query, index.WithTopK(3))
	pipeline.NewQATube(openai.NewChat().WithVerbose(true)).Run(context.Background(), query, similarities.ToDocuments())
}
