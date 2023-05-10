package main

import (
	"context"
	"fmt"

	openaiembedder "github.com/henomis/lingoose/embedder/openai"
	"github.com/henomis/lingoose/index"
	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/loader"
	"github.com/henomis/lingoose/pipeline"
	"github.com/henomis/lingoose/textsplitter"
)

func main() {
	docs, _ := loader.NewPDFToTextLoader("/usr/bin/pdftotext", "./kb").Load()
	chunks := textsplitter.NewRecursiveCharacterTextSplitter(2000, 200, nil, nil).SplitDocuments(docs)
	openaiEmbedder := openaiembedder.New(openaiembedder.AdaEmbeddingV2)
	index.NewSimpleVectorIndex("db", ".", openaiEmbedder).LoadFromDocuments(context.Background(), chunks)
	query := "Who is Mario?"
	topk := 3
	similarities, _ := index.NewSimpleVectorIndex("db", ".", openaiEmbedder).SimilaritySearch(context.Background(), query, &topk)
	qa := pipeline.NewQATube(openai.NewChat().WithVerbose(true))
	resp, _ := qa.Run(context.Background(), query, similarities)
	fmt.Println(resp)
}
