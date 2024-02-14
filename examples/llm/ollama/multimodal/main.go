package main

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/llm/ollama"
	"github.com/henomis/lingoose/thread"
)

func main() {
	ollamallm := ollama.New().WithModel("llama2").WithVisionModel("llava").WithConvertImageContentToText(true).WithTemperature(0)

	t := thread.New().AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewImageContentFromURL("https://upload.wikimedia.org/wikipedia/commons/thumb/3/34/Anser_anser_1_%28Piotr_Kuczynski%29.jpg/1280px-Anser_anser_1_%28Piotr_Kuczynski%29.jpg"),
		).AddContent(
			thread.NewTextContent("Can you describe the image?"),
		),
	)

	err := ollamallm.Generate(context.Background(), t)
	if err != nil {
		panic(err)
	}

	fmt.Println(t)
}
