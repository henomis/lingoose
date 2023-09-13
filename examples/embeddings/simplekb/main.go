package main

import (
	"context"

	openaiembedder "github.com/henomis/lingoose/embedder/openai"
	"github.com/henomis/lingoose/index/retriever"
	simplevectorindex "github.com/henomis/lingoose/index/simpleVectorIndex"
	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/loader"
	qapipeline "github.com/henomis/lingoose/pipeline/qa"
	"github.com/henomis/lingoose/textsplitter"
)

func main() {
	docs, _ := loader.NewPDFToTextLoader("./kb").WithPDFToTextPath("/opt/homebrew/bin/pdftotext").WithTextSplitter(textsplitter.NewRecursiveCharacterTextSplitter(2000, 200)).Load(context.Background())
	qapipeline.New(openai.NewChat().WithVerbose(true)).WithDocuments(&docs).WithRetriever(retriever.New(simplevectorindex.New("db", ".", openaiembedder.New(openaiembedder.AdaEmbeddingV2)))).Query(context.Background(), "What is the NATO purpose?")
}
