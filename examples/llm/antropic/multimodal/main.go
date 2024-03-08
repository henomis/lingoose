package main

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/llm/antropic"
	"github.com/henomis/lingoose/thread"
)

func main() {
	antropicllm := antropic.New().WithModel("claude-3-opus-20240229")

	t := thread.New().AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent("Can you describe the image?"),
		).AddContent(
			thread.NewImageContentFromURL("https://upload.wikimedia.org/wikipedia/commons/thumb/3/34/Anser_anser_1_%28Piotr_Kuczynski%29.jpg/1280px-Anser_anser_1_%28Piotr_Kuczynski%29.jpg"),
		),
	)

	err := antropicllm.Generate(context.Background(), t)
	if err != nil {
		panic(err)
	}

	fmt.Println(t)
}
