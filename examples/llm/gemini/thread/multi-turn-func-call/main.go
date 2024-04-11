package main

import (
	"cloud.google.com/go/vertexai/genai"
	"context"
	"fmt"
	"github.com/henomis/lingoose/llm/gemini"
	"google.golang.org/api/option"
	"log"
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

	//err := ExampleTool()
	//if err == nil {
	//	return
	//}

	client, err := genai.NewClient(ctx, PROJECT, REGION, option.WithCredentialsFile(GCP_KEY_PATH))
	if err != nil {
		return
	}
	geminiLLM := gemini.New(client, gemini.Gemini1Pro001).WithStream(true,
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

func ExampleTool() error {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, PROJECT, REGION, option.WithCredentialsFile(GCP_KEY_PATH))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	//currentWeather := func(city string) string {
	//	switch city {
	//	case "New York, NY":
	//		return "cold"
	//	case "Miami, FL":
	//		return "warm"
	//	default:
	//		return "unknown"
	//	}
	//}

	// To use functions / tools, we have to first define a schema that describes
	// the function to the model. The schema is similar to OpenAPI 3.0.
	//
	// In this example, we create a single function that provides the model with
	// a weather forecast in a given location.
	//schema := &genai.Schema{
	//	Type: genai.TypeObject,
	//	Properties: map[string]*genai.Schema{
	//		"location": {
	//			Type:        genai.TypeString,
	//			Description: "The city and state, e.g. San Francisco, CA",
	//		},
	//		"unit": {
	//			Type: genai.TypeString,
	//			Enum: []string{"celsius", "fahrenheit"},
	//		},
	//	},
	//	Required: []string{"location"},
	//}
	//
	//weatherTool := &genai.Tool{
	//	FunctionDeclarations: []*genai.FunctionDeclaration{{
	//		Name:        "CurrentWeather",
	//		Description: "Get the current weather in a given location",
	//		Parameters:  schema,
	//	}},
	//}

	model := client.GenerativeModel("gemini-1.0-pro")

	// Before initiating a conversation, we tell the model which tools it has
	// at its disposal.
	model.Tools = []*genai.Tool{buildFuncTool()[0]}

	// For using tools, the chat mode is useful because it provides the required
	// chat context. A model needs to have tools supplied to it in the chat
	// history so it can use them in subsequent conversations.
	//
	// The flow of message expected here is:
	//
	// 1. We send a question to the model
	// 2. The model recognizes that it needs to use a tool to answer the question,
	//    an returns a FunctionCall response asking to use the CurrentWeather
	//    tool.
	// 3. We send a FunctionResponse message, simulating the return value of
	//    CurrentWeather for the model's query.
	// 4. The model provides its text answer in response to this message.
	session := model.StartChat()

	res, err := session.SendMessage(ctx, genai.Text("Hello, I'm a user"), genai.Text("Can you greet me?"), genai.Text("please greet me as a pirate."))
	if err != nil {
		log.Fatal(err)
	}

	part := res.Candidates[0].Content.Parts[0]
	funcall, ok := part.(genai.FunctionCall)
	if ok {
		log.Println(funcall.Name, " ", funcall.Args)
	}
	//
	//if funcall.Name != "CurrentWeather" {
	//	log.Fatalf("expected CurrentWeather: %v", funcall.Name)
	//}

	// Expect the model to pass a proper string "location" argument to the tool.
	//locArg, ok := funcall.Args["location"].(string)
	//if !ok {
	//	log.Fatalf("expected string: %v", funcall.Args["location"])
	//}

	res, err = session.SendMessage(ctx, genai.Text("Where are you?"))
	if err != nil {
		return err
	}

	part = res.Candidates[0].Content.Parts[0]
	funcall, ok = part.(genai.FunctionCall)
	if ok {
		log.Println(funcall.Name, " ", funcall.Args)
	}
	//weatherData := currentWeather(locArg)
	//res, err = session.SendMessage(ctx, genai.FunctionResponse{
	//	Name: weatherTool.FunctionDeclarations[0].Name,
	//	Response: map[string]any{
	//		"weather": weatherData,
	//	},
	//})
	//if err != nil {
	//	log.Fatal(err)
	//}

	printResponse(res)
	return nil
}

func printResponse(resp *genai.GenerateContentResponse) {
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				fmt.Println(part)
			}
		}
	}
	fmt.Println("---")
}
