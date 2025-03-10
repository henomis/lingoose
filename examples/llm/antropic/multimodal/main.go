package main

import (
	"context"
	"fmt"
	"os"

	"github.com/henomis/lingoose/llm/anthropic"
	"github.com/henomis/lingoose/thread"
)

func main() {
	anthropicllm := anthropic.NewAnthropic(os.Getenv("ANTHROPIC_API_KEY")).WithModel(anthropic.ModelClaude_3_Opus_20240229)

	t := thread.New().AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent("Can you describe the image?"),
		).AddContent(
			thread.NewImageContentFromURL("https://upload.wikimedia.org/wikipedia/commons/thumb/3/34/Anser_anser_1_%28Piotr_Kuczynski%29.jpg/1280px-Anser_anser_1_%28Piotr_Kuczynski%29.jpg"),
		),
	)

	err := anthropicllm.Chat(context.Background(), t)
	if err != nil {
		panic(err)
	}

	fmt.Println(t)
}
