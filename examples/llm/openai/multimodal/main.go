package main

import (
	"context"
	"fmt"

	"github.com/rsest/lingoose/llm/openai"
	"github.com/rsest/lingoose/thread"
)

func main() {
	openaillm := openai.New().WithModel(openai.GPT4VisionPreview)

	t := thread.New().AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent("Can you describe the image?"),
		).AddContent(
			thread.NewImageContentFromURL("https://upload.wikimedia.org/wikipedia/commons/thumb/3/34/Anser_anser_1_%28Piotr_Kuczynski%29.jpg/1280px-Anser_anser_1_%28Piotr_Kuczynski%29.jpg"),
		),
	)

	fmt.Println(t)

	err := openaillm.Generate(context.Background(), t)
	if err != nil {
		panic(err)
	}

	fmt.Println(t)
}
