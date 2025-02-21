package main

import (
	"context"
	"fmt"

	openaiembedder "github.com/rsest/lingoose/embedder/openai"
	"github.com/rsest/lingoose/index"
	indexoption "github.com/rsest/lingoose/index/option"
	pineconedb "github.com/rsest/lingoose/index/vectordb/pinecone"
	"github.com/rsest/lingoose/legacy/prompt"
	"github.com/rsest/lingoose/llm/openai"
	"github.com/rsest/lingoose/loader"
	"github.com/rsest/lingoose/textsplitter"
)

// download https://raw.githubusercontent.com/hwchase17/chat-your-data/master/state_of_the_union.txt

func main() {
	replicas := 1
	shards := 1
	index := index.New(
		pineconedb.New(
			pineconedb.Options{
				IndexName: "test",
				Namespace: "test-namespace",
				CreateIndexOptions: &pineconedb.CreateIndexOptions{
					Dimension: 1536,
					Metric:    "cosine",
					Pod: &pineconedb.Pod{
						Replicas:    &replicas,
						Shards:      &shards,
						PodType:     "s1.x1",
						Environment: "gcp-starter",
					},
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
		"Based on the following context answer to the question.\n\n" +
			"Context:\n{{.context}}\n\nQuestion: {{.query}}").WithInputs(
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
