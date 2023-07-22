package main

import (
	"context"
	"fmt"

	openaiembedder "github.com/henomis/lingoose/embedder/openai"
	"github.com/henomis/lingoose/index"
	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/loader"
	"github.com/henomis/lingoose/prompt"
	"github.com/henomis/lingoose/textsplitter"
)

// download https://raw.githubusercontent.com/hwchase17/chat-your-data/master/state_of_the_union.txt
// run qdrant docker run -p 6333:6333 qdrant/qdrant

func main() {

	openaiEmbedder := openaiembedder.New(openaiembedder.AdaEmbeddingV2)

	qdrantIndex := index.NewQdrant(
		index.QdrantOptions{
			CollectionName: "test",
			IncludeContent: true,
			CreateCollection: &index.QdrantCreateCollectionOptions{
				Dimension: 1536,
				Distance:  index.QdrantDistanceCosine,
			},
		},
		openaiEmbedder,
	).WithAPIKeyAndEdpoint("", "http://localhost:6333")

	indexIsEmpty, err := qdrantIndex.IsEmpty(context.Background())
	if err != nil {
		panic(err)
	}

	if indexIsEmpty {
		err = ingestData(qdrantIndex)
		if err != nil {
			panic(err)
		}
	}

	query := "What is the purpose of the NATO Alliance?"
	similarities, err := qdrantIndex.SimilaritySearch(
		context.Background(),
		query,
		index.WithTopK(3),
	)
	if err != nil {
		panic(err)
	}

	content := ""
	for _, similarity := range similarities {
		fmt.Printf("Similarity: %f\n", similarity.Score)
		fmt.Printf("Document: %s\n", similarity.Document.Content)
		fmt.Println("Metadata: ", similarity.Document.Metadata)
		fmt.Println("ID: ", similarity.ID)
		fmt.Println("----------")
		content += similarity.Document.Content + "\n"
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

func ingestData(qdrantIndex *index.Qdrant) error {

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

	return qdrantIndex.LoadFromDocuments(context.Background(), documentChunks)

}
