package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/thread"
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
	openaillm.WithToolChoice(newStr("auto"))
	err := openaillm.BindFunction(
		crateImage,
		"createImage",
		"use this function to create an image from a description",
	)
	if err != nil {
		panic(err)
	}

	t := thread.New().AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent("Please, create an image that inspires you"),
		),
	)

	err = openaillm.Generate(context.Background(), t)
	if err != nil {
		panic(err)
	}

	if t.LastMessage().Role == thread.RoleTool {
		t.AddMessage(thread.NewUserMessage().AddContent(
			thread.NewImageContentFromURL(
				strings.ReplaceAll(t.LastMessage().Contents[0].AsToolResponseData().Result, `"`, ""),
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
