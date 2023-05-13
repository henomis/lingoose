package main

import (
	"context"
	"fmt"

	llamacppembedder "github.com/henomis/lingoose/embedder/llamacpp"
)

func main() {
	llamacppEmbedder := llamacppembedder.New().
		WithModel("/home/simone/GIT/lingoose/llama.cpp/models/7B/ggml-model-q4_0.bin").
		WithLlamaCppPath("/home/simone/GIT/lingoose/llama.cpp/embedding")

	embeddings, err := llamacppEmbedder.Embed(context.Background(), []string{"hello", "world"})
	if err != nil {
		panic(err)
	}

	for _, embedding := range embeddings {
		fmt.Printf("%#v\n", embedding)
		fmt.Println(len(embedding))
	}

}
