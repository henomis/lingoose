package main

import (
	"cloud.google.com/go/vertexai/genai"
	"context"
	"fmt"
	"github.com/henomis/lingoose/llm/gemini"
	"google.golang.org/api/option"
	"os"

	"github.com/henomis/lingoose/thread"
)

var (
	PROJECT      = "conversenow-dev"
	REGION       = "us-central1"
	GCP_KEY_PATH string
)

func init() {
	GCP_KEY_PATH = os.Getenv("GCP_KEY_PATH")
}

func main() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, PROJECT, REGION, option.WithCredentialsFile(GCP_KEY_PATH))
	if err != nil {
		return
	}
	geminiLLM := gemini.New(client, gemini.Gemini1Pro001).WithStream(true, func(string) {})

	t := thread.New().AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent("Hello, I'm a user"),
		).AddContent(
			thread.NewTextContent("Can you greet me?"),
		),
	).AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent("please greet me as a pirate."),
		),
	)
	fmt.Println("INPUT THREAD ::")
	fmt.Println(t.String())

	err = geminiLLM.Generate(context.Background(), t)
	if err != nil {
		panic(err)
	}

	fmt.Println("PREDICTION THREAD ::")
	fmt.Println(t.String())

	t.AddMessage(thread.NewUserMessage().AddContent(
		thread.NewTextContent("now translate to italian as a poem"),
	))

	fmt.Println("INPUT THREAD ::")
	fmt.Println(t.String())

	err = geminiLLM.Generate(context.Background(), t)
	if err != nil {
		panic(err)
	}

	fmt.Println("PREDICTION THREAD ::")
	fmt.Println(t.String())

}
