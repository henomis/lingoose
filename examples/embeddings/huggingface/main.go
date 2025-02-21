package main

import (
	"context"
	"fmt"

	huggingfaceembedder "github.com/rsest/lingoose/embedder/huggingface"
)

func main() {
	hfEmbedder := huggingfaceembedder.New()

	embeddings, err := hfEmbedder.Embed(context.Background(), []string{"hello", "world"})
	if err != nil {
		panic(err)
	}

	for _, embedding := range embeddings {
		fmt.Printf("%#v\n", embedding)
		fmt.Println(len(embedding))
	}

}
