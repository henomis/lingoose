package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/thread"
	"github.com/henomis/lingoose/tools/dalle"
	"github.com/henomis/lingoose/transformer"
)

type Image struct {
	Description string `json:"description" jsonschema:"description=the description of the image that should be created"`
}

func crateImage(i Image) string {
	d := transformer.NewDallE().WithImageSize(transformer.DallEImageSize512x512)
	imageURL, err := d.Transform(context.Background(), i.Description)
	if err != nil {
		return fmt.Errorf("error creating image: %w", err).Error()
	}

	fmt.Println("Image created with url:", imageURL)

	return imageURL.(string)
}

func newStr(str string) *string {
	return &str
}

func main() {
	openaillm := openai.New().WithModel(openai.GPT4o)
	openaillm.WithToolChoice(newStr("auto")).WithTools(dalle.New())

	t := thread.New().AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent("Please, create an image that inspires you"),
		),
	)

	err := openaillm.Generate(context.Background(), t)
	if err != nil {
		panic(err)
	}

	if t.LastMessage().Role == thread.RoleTool {
		var output dalle.Output

		err = json.Unmarshal([]byte(t.LastMessage().Contents[0].AsToolResponseData().Result), &output)
		if err != nil {
			panic(err)
		}

		t.AddMessage(thread.NewUserMessage().AddContent(
			thread.NewImageContentFromURL(
				output.ImageURL,
			),
		).AddContent(
			thread.NewTextContent("can you describe the image?"),
		))

		err = openaillm.Generate(context.Background(), t)
		if err != nil {
			panic(err)
		}
	}

	fmt.Println(t)
}
