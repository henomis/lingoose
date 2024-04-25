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

type Answer struct {
	Answer string `json:"answer" jsonschema:"description=the pirate answer"`
}

func getAnswer(a Answer) string {
	return "🦜 ☠️ " + a.Answer
}

func buildFuncTool() []*genai.Tool {
	var tools []*genai.Tool

	schema := &genai.Schema{
		Type: genai.TypeObject,

		Properties: map[string]*genai.Schema{
			"answer": {
				Type:        genai.TypeString,
				Description: "the pirate answer",
			},
		},
		Required: []string{"answer"},
	}

	answerTool := &genai.Tool{
		FunctionDeclarations: []*genai.FunctionDeclaration{{
			Name:        "getAnswer",
			Description: "run this function to get pirate answer",
			Parameters:  schema,
		}},
	}

	tools = append(tools, answerTool)
	return tools
}

func streamCallBack(s string) {
	if s == gemini.EOS {
		fmt.Printf("\n")
		return
	}
	fmt.Printf("%s \n", s)
}

func main() {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, PROJECT, REGION, option.WithCredentialsFile(GCP_KEY_PATH))
	if err != nil {
		return
	}
	geminiLLM := gemini.New(ctx, client, gemini.Gemini1Pro001).WithStream(true,
		streamCallBack).WithTools(buildFuncTool())

	err = geminiLLM.BindFunction(
		getAnswer,
		"getAnswer",
		"use this function when pirate finishes his answer")

	if err != nil {
		panic(err)
	}

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

	//Clear thread
	t.ClearMessages()
	t.AddMessage(thread.NewUserMessage().AddContent(
		thread.NewTextContent("Have you ever looted any ship?"),
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
