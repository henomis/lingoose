package main

import (
	"context"
	"fmt"
	"os"

	"github.com/henomis/lingoose/embedding"
	"github.com/henomis/lingoose/index"
	"github.com/henomis/lingoose/loader"
	"github.com/henomis/lingoose/textsplitter"

	"github.com/pkoukk/tiktoken-go"
	"github.com/sashabaranov/go-openai"
)

func NumTokensFromMessages(messages []openai.ChatCompletionMessage, model string) (num_tokens int) {
	tkm, err := tiktoken.EncodingForModel(model)
	if err != nil {
		err = fmt.Errorf("EncodingForModel: %v", err)
		fmt.Println(err)
		return
	}

	var tokens_per_message int
	var tokens_per_name int
	if model == "gpt-3.5-turbo-0301" || model == "gpt-3.5-turbo" {
		tokens_per_message = 4
		tokens_per_name = -1
	} else if model == "gpt-4-0314" || model == "gpt-4" {
		tokens_per_message = 3
		tokens_per_name = 1
	} else {
		fmt.Println("Warning: model not found. Using cl100k_base encoding.")
		tokens_per_message = 3
		tokens_per_name = 1
	}

	for _, message := range messages {
		num_tokens += tokens_per_message
		num_tokens += len(tkm.Encode(message.Content, nil, nil))
		num_tokens += len(tkm.Encode(message.Role, nil, nil))
		if message.Name != "" {
			num_tokens += tokens_per_name
		}
	}
	num_tokens += 3
	return num_tokens
}

func main() {

	loader, err := loader.NewDirLoader(".", ".md")
	if err != nil {
		panic(err)
	}

	documents, err := loader.Load()
	if err != nil {
		panic(err)
	}

	text_splitter := textsplitter.NewRecursiveCharacterTextSplitter(nil, 1000, 0, nil)

	docs := text_splitter.SplitDocuments(documents)

	for _, doc := range docs {
		fmt.Println(doc.Content)
		fmt.Println("----------")
		fmt.Println(doc.Metadata)
		fmt.Println("----------")
		fmt.Println()

	}

	embed, err := embedding.NewOpenAIEmbeddings(embedding.AdaSimilarity)
	if err != nil {
		panic(err)
	}

	vector := new(index.VectorIndex)
	_, err = os.Stat("vector.json")
	if err != nil {

		objs, err := embed.Embed(context.Background(), docs)
		if err != nil {
			panic(err)
		}

		vector := index.NewVectorIndex(objs)
		vector.Save("vector.json")
	}
	vector.Load("vector.json")

	objs, err := embed.Embed(context.Background(), docs)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", objs)

	// model := "davinci"

	// b, err := os.ReadFile("README.md")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// tkm, err := tiktoken.EncodingForModel(model)
	// if err != nil {
	// 	err = fmt.Errorf("getEncoding: %v", err)
	// 	return
	// }

	// // encode
	// token := tkm.Encode(string(b), nil, nil)

	// // num_tokens
	// fmt.Println(len(token))
	// fmt.Printf("%#v", token)
}
