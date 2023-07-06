package main

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/chat"
	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/prompt"
)

func main() {

	chat := chat.New(
		chat.PromptMessage{
			Type: chat.MessageTypeUser,
			Prompt: prompt.New(
				`

There once was an ambitious man who enjoyed cars. He had a very special collection of automobiles, each one with a unique appeal that spoke to him.

He had an old Volvo S80, known for its nice exterior and superior safety features. He also had a brand new Audi S4, complete with all the fancy bells and whistles and sleek lines. Then there was his pride and joy, an Aston Martin DB11, built to impress and ignite his passion every time he took it out on a drive.

The man loved to take long joy rides with his beloved cars. He had the most memorable trip when he drove his Volvo to the mountains to take long hikes and admire the natural beauty of his surroundings. When he was feeling a bit fancier, he'd take the Audi out for a spin around town, showing off its power and performance. But when he really wanted an adventure, he'd take the Aston Martin on a road trip, exploring some of the most breathtaking landscapes around the world.

He cherished every moment spent behind the wheel of his special cars and loved to reminisce of the great memories they created. He felt a strong connection with each of them, like he was part of something truly unique. And even though each one had its own character, he felt like they were all pieces of his life's story.
			`),
		},
	)

	llmOpenAI := openai.New(openai.GPT3Dot5Turbo0613, openai.DefaultOpenAITemperature, openai.DefaultOpenAIMaxTokens, true).WithFunctionCallOptions(true)

	var carInText Car

	err := llmOpenAI.BindFunction(
		ListCars,
		"get_cars_list",
		"Extract cars brand and model from a text",
	)
	if err != nil {
		panic(err)
	}

	response, err := llmOpenAI.Chat(context.Background(), chat)
	if err != nil {
		panic(err)
	}

	fmt.Printf("\n%s", response)

	fmt.Printf("\n\nCar in text: %+v", carInText)

}

type Car struct {
	Brand string `json:"brand" jsonschema:"description=The brand of the car"`
	Model string `json:"model" jsonschema:"description=The model of the car"`
}

type Cars struct {
	Cars []Car `json:"name" jsonschema:"description=The list of cars"`
}

type Country struct {
	CountryID   string  `json:"country_id"`
	Probability float64 `json:"probability"`
}

type NationalizeResponse struct {
	Name    string    `json:"name"`
	Country []Country `json:"country"`
}

func ListCars(car Cars) (Cars, error) {
	fmt.Printf("\n\nListCars: %+v", car)
	return car, nil
}
