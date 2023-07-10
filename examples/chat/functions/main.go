package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/henomis/lingoose/chat"
	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/prompt"
)

func main() {
	fmt.Printf("What's your name?\n> ")
	reader := bufio.NewReader(os.Stdin)
	name, _ := reader.ReadString('\n')

	llmChat := chat.New(

		chat.PromptMessage{
			Type: chat.MessageTypeUser,
			Prompt: prompt.New(
				fmt.Sprintf("My name is %s, can you guess my nationality?", name),
			),
		},
	)

	llmOpenAI := openai.New(openai.GPT3Dot5Turbo0613, openai.DefaultOpenAITemperature, openai.DefaultOpenAIMaxTokens, true)

	llmOpenAI.BindFunction(
		GetNationalitiesForName,
		"GetNationalitiesForName",
		"Use this function to get the nationalities for a given name.",
	)

	response, err := llmOpenAI.Chat(context.Background(), llmChat)
	if err != nil {
		panic(err)
	}

	if llmOpenAI.CalledFunctionName() == nil {
		fmt.Printf("expected called function name to be set")
		return
	}

	llmChat.AddPromptMessages(
		[]chat.PromptMessage{
			{
				Type:   chat.MessageTypeFunction,
				Prompt: prompt.New(response),
				Name:   llmOpenAI.CalledFunctionName(),
			},
		},
	)

	_, err = llmOpenAI.Chat(context.Background(), llmChat)
	if err != nil {
		panic(err)
	}

}

type Query struct {
	Name string `json:"name" jsonschema:"description=The name to get the nationalities for"`
}

type Country struct {
	CountryID   string  `json:"country_id"`
	Probability float64 `json:"probability"`
}

type NationalizeResponse struct {
	Name    string    `json:"name"`
	Country []Country `json:"country"`
}

func GetNationalitiesForName(query Query) ([]Country, error) {
	url := fmt.Sprintf("https://api.nationalize.io/?name=%s", query.Name)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response NationalizeResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return response.Country, nil
}
