package main

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/vertexai/genai"
	"github.com/henomis/lingoose/llm/gemini"
	"google.golang.org/api/option"
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
		streamCallBack).WithTools(buildFuncTool())

	err = geminiLLM.BindFunction(
		getAnswer,
		"getAnswer",
		"use this function when pirate finishes his answer")

	if err != nil {
		panic(err)
	}

	model := client.GenerativeModel(gemini.Gemini1Pro.String())

	// Before initiating a conversation, we tell the model which tools it has
	// at its disposal.
	model.Tools = []*genai.Tool{buildFuncTool()[0]}
	sess := model.StartChat()

	res, err := sess.SendMessage(ctx, genai.Text("Hello, I'm a user"),
		genai.Text("Can you greet me?"),
		genai.Text("Have you to act like a pirate."),
		genai.Text("please greet me as a pirate."),
	)
	if err != nil {
		panic(err)
	}

	part := res.Candidates[0].Content.Parts[0]
	funcall, ok := part.(genai.FunctionCall)
	if ok {
		fmt.Println("Received func call: ", funcall.Name)
		if funcall.Name == "getAnswer" {
			ans := Answer{Answer: funcall.Args["answer"].(string)}
			fmt.Println(getAnswer(ans))
		}
	}

	//_, err = sess.SendMessage(ctx, genai.FunctionResponse{
	//	Name:     "getAnswer",
	//	Response: nil},
	//)
	//if err != nil {
	//	panic(err)
	//}

	res, err = sess.SendMessage(ctx, genai.Text("Where are you sailing today?"))
	if err != nil {
		panic(err)
	}

	part = res.Candidates[0].Content.Parts[0]
	funcall, ok = part.(genai.FunctionCall)
	if ok {
		fmt.Println("Received func call: ", funcall.Name)
		if funcall.Name == "getAnswer" {
			ans := Answer{Answer: funcall.Args["answer"].(string)}
			fmt.Println(getAnswer(ans))
		}
	}

	res, err = sess.SendMessage(ctx, genai.Text("How long will it take you to sail from europe to india?"))
	if err != nil {
		panic(err)
	}

	part = res.Candidates[0].Content.Parts[0]
	funcall, ok = part.(genai.FunctionCall)
	if ok {
		fmt.Println("Received func call: ", funcall.Name)
		if funcall.Name == "getAnswer" {
			ans := Answer{Answer: funcall.Args["answer"].(string)}
			fmt.Println(getAnswer(ans))
		}
	}

	PrintHistory(sess.History)
}
