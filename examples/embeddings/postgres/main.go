package main

import (
	"context"
	"database/sql"
	"fmt"

	openaiembedder "github.com/henomis/lingoose/embedder/openai"
	"github.com/henomis/lingoose/index"
	indexoption "github.com/henomis/lingoose/index/option"
	"github.com/henomis/lingoose/index/vectordb/postgres"
	"github.com/henomis/lingoose/legacy/prompt"
	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/loader"
	"github.com/henomis/lingoose/textsplitter"
	// uncomment to use postgres
	// _ "github.com/lib/pq"
)

// download https://raw.githubusercontent.com/hwchase17/chat-your-data/master/state_of_the_union.txt

func main() {

	connStr := "user=root sslmode=disable password=pass dbname=test host=localhost port=5432"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	index := index.New(
		postgres.New(
			postgres.Options{
				DB:    db,
				Table: "test",
				CreateIndex: &postgres.CreateIndexOptions{
					Dimension: 1536,
					Distance:  postgres.DistanceCosine,
				},
			},
		),
		openaiembedder.New(openaiembedder.AdaEmbeddingV2),
	).WithIncludeContents(true)

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
		fmt.Println("----------")
		content += similarity.Content() + "\n"
	}

	llmOpenAI := openai.NewCompletion().WithVerbose(true)

	prompt1 := prompt.NewPromptTemplate(
		"Based on the following context answer to the question.\n\nContext:\n{{.context}}\n\nQuestion: {{.query}}").
		WithInputs(
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

func ingestData(redisIndex *index.Index) error {
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

	return redisIndex.LoadFromDocuments(context.Background(), documentChunks)
}
