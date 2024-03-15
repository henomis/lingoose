package main

import (
	"context"
	"fmt"

	openaiembedder "github.com/henomis/lingoose/embedder/openai"
	"github.com/henomis/lingoose/index"
	indexoption "github.com/henomis/lingoose/index/option"
	milvusdb "github.com/henomis/lingoose/index/vectordb/milvus"
	"github.com/henomis/lingoose/legacy/prompt"
	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/loader"
	"github.com/henomis/lingoose/textsplitter"
)

// download https://raw.githubusercontent.com/hwchase17/chat-your-data/master/state_of_the_union.txt
// run milvus docker run --rm -p 6333:6333 milvus/milvus

func main() {

	index := index.New(
		milvusdb.New(
			milvusdb.Options{
				CollectionName: "test",
				CreateCollection: &milvusdb.CreateCollectionOptions{
					Dimension: 1536,
					Metric:    milvusdb.DistanceL2,
				},
			},
		).WithCredentialsAndEndpoint("root", "Milvus", "http://localhost:19530"),
		openaiembedder.New(openaiembedder.AdaEmbeddingV2),
	)

	indexIsEmpty, err := index.IsEmpty(context.Background())
	if err != nil {
		panic(err)
	}

	if indexIsEmpty {
		err = ingestData(index)
		if err != nil {
			panic(err)
		}
	}

	query := "What is the purpose of the NATO Alliance?"
	similarities, err := index.Query(
		context.Background(),
		query,
		indexoption.WithTopK(3),
	)
	if err != nil {
		panic(err)
	}

	content := ""
	for _, similarity := range similarities {
		fmt.Printf("Similarity: %f\n", similarity.Score)
		fmt.Printf("Document: %s\n", similarity.Content())
		fmt.Println("Metadata: ", similarity.Metadata)
		fmt.Println("ID: ", similarity.ID)
		fmt.Println("Values: ", similarity.Values)
		fmt.Println("----------")
		content += similarity.Content() + "\n"
	}

	llmOpenAI := openai.NewCompletion().WithVerbose(true)

	prompt1 := prompt.NewPromptTemplate(
		"Based on the following context answer to the question.\n\nContext:\n{{.context}}\n\nQuestion: {{.query}}").WithInputs(
		map[string]string{
			"query":   query,
			"context": content,
		},
	)

	err = prompt1.Format(nil)
	if err != nil {
		panic(err)
	}

	_, err = llmOpenAI.Completion(context.Background(), prompt1.String())
	if err != nil {
		panic(err)
	}

}

func ingestData(index *index.Index) error {

	documents, err := loader.NewDirectoryLoader(".", ".txt").Load(context.Background())
	if err != nil {
		return err
	}

	textSplitter := textsplitter.NewRecursiveCharacterTextSplitter(1000, 20)

	documentChunks := textSplitter.SplitDocuments(documents)

	for _, doc := range documentChunks {
		fmt.Println(doc.Content)
		fmt.Println("----------")
		fmt.Println(doc.Metadata)
		fmt.Println("----------")
		fmt.Println()

	}

	return index.LoadFromDocuments(context.Background(), documentChunks)

}
