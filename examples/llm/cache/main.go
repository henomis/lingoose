package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	openaiembedder "github.com/henomis/lingoose/embedder/openai"
	"github.com/henomis/lingoose/index"
	"github.com/henomis/lingoose/index/vectordb/jsondb"
	"github.com/henomis/lingoose/llm/cache"
	"github.com/henomis/lingoose/llm/openai"
)

func main() {

	embedder := openaiembedder.New(openaiembedder.AdaEmbeddingV2)
	index := index.New(
		jsondb.New("db.json"),
		embedder,
	)
	llm := openai.NewCompletion().WithCompletionCache(cache.New(embedder, index).WithTopK(3))

	for {
		text := askUserInput("What is your question?")

		response, err := llm.Completion(context.Background(), text)
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Println(response)
	}
}

func askUserInput(question string) string {
	fmt.Printf("%s > ", question)
	reader := bufio.NewReader(os.Stdin)
	name, _ := reader.ReadString('\n')
	name = strings.TrimSuffix(name, "\n")
	return name
}
