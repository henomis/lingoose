package main

import (
	"cloud.google.com/go/vertexai/genai"
	"context"
	"fmt"
	"github.com/henomis/lingoose/llm/gemini"
	"github.com/henomis/lingoose/thread"
	"google.golang.org/api/option"
	"os"
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

func doNothing(a Answer) string {
	return ""
}

func answerTool() *genai.Tool {
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
	answerT := &genai.Tool{
		FunctionDeclarations: []*genai.FunctionDeclaration{{
			Name:        "getAnswer",
			Description: "run this function to get pirate answer",
			Parameters:  schema,
		}},
	}

	return answerT
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

	doNothingschema := &genai.Schema{
		Type: genai.TypeObject,

		Properties: map[string]*genai.Schema{},
		Required:   []string{""},
	}
	//
	//doNothingTool := &genai.Tool{
	//	FunctionDeclarations: []*genai.FunctionDeclaration{{
	//		Name:        "doNothing",
	//		Description: "",
	//		Parameters:  doNothingschema,
	//	}},
	//}

	answerTool := &genai.Tool{
		FunctionDeclarations: []*genai.FunctionDeclaration{{
			Name:        "getAnswer",
			Description: "run this function to get pirate answer",
			Parameters:  schema,
		}, {Name: "doNothing",
			//Description: "never call this function anywhere",
			Description: "call this function before pirate answer",
			Parameters:  doNothingschema,
		}},
	}

	tools = append(tools, answerTool)
	//tools = append(tools, doNothingTool)
	return tools
}

func streamCallBack(s string) {
	if s == gemini.EOS {
		fmt.Printf("\n")
		return
	}
	fmt.Printf("%s \n", s)
}

func PrintHistory(history []*genai.Content) {
	for _, h := range history {
		fmt.Println("Role: ", h.Role)
		fmt.Println("Parts: ", gemini.PartsTostring(h.Parts))
		fmt.Println("--------------------------------------")
	}
}

func main() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, PROJECT, REGION, option.WithCredentialsFile(GCP_KEY_PATH))
	if err != nil {
		return
	}
	defer client.Close()

	geminiLLM := gemini.New(ctx, client, gemini.Gemini1Pro001).WithStream(true,
		streamCallBack).WithChatMode()

	err = geminiLLM.BindFunction(
		getAnswer,
		"getAnswer",
		"use this function when pirate finishes his answer")

	err = geminiLLM.BindFunction(
		doNothing,
		"doNothing",
		"never call this function")

	if err != nil {
		panic(err)
	}

	// Before initiating a conversation, we tell the model which tools it has
	// at its disposal.
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

	geminiLLM.WithTools(buildFuncTool())
	err = geminiLLM.Generate(context.Background(), t)
	if err != nil {
		panic(err)
	}

	fmt.Println("PREDICTION THREAD ::")
	fmt.Println(t.String())
	fmt.Println("------------------------------------")
	fmt.Println("SESSION HISTORY :: ")
	PrintHistory(geminiLLM.GetChatHistory())

	if t.LastMessage().Role == thread.RoleTool {
		err = geminiLLM.Generate(context.Background(), t)
		if err != nil {
			panic(err)
		}
	}

	//Clear thread
	//t.ClearMessages()
	t.AddMessage(thread.NewUserMessage().AddContent(
		thread.NewTextContent("Have you ever looted any ship?"),
	))

	geminiLLM.ClearTools()
	geminiLLM.WithTools([]*genai.Tool{answerTool()})

	err = geminiLLM.Generate(context.Background(), t)
	if err != nil {
		panic(err)
	}

	if t.LastMessage().Role == thread.RoleTool {
		err = geminiLLM.Generate(context.Background(), t)
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("PREDICTION THREAD ::")
	fmt.Println(t.String())
	fmt.Println("------------------------------------")
	fmt.Println("SESSION HISTORY ::")
	PrintHistory(geminiLLM.GetChatHistory())

	t.AddMessage(thread.NewUserMessage().AddContent(
		thread.NewTextContent("Thank you! Bye!"),
	))
	err = geminiLLM.Generate(context.Background(), t)
	if err != nil {
		panic(err)
	}

	if t.LastMessage().Role == thread.RoleTool {
		err = geminiLLM.Generate(context.Background(), t)
		if err != nil {
			panic(err)
		}
	}
	fmt.Println("PREDICTION THREAD ::")
	fmt.Println(t.String())
	fmt.Println("------------------------------------")
	fmt.Println("SESSION HISTORY ::")
	PrintHistory(geminiLLM.GetChatHistory())

}
