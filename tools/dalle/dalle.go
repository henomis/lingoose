package dalle

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/transformer"
)

type Tool struct {
}

type Input struct {
	Description string `json:"description" jsonschema:"description=the description of the image that should be created"`
}

type Output struct {
	Error    string `json:"error,omitempty"`
	ImageURL string `json:"imageURL,omitempty"`
}

type FnPrototype func(Input) Output

func New() *Tool {
	return &Tool{}
}

func (t *Tool) Name() string {
	return "dalle"
}

func (t *Tool) Description() string {
	return "A tool that creates an image from a description."
}

func (t *Tool) Fn() any {
	return t.fn
}

func (t *Tool) fn(i Input) Output {
	d := transformer.NewDallE().WithImageSize(transformer.DallEImageSize512x512)
	imageURL, err := d.Transform(context.Background(), i.Description)
	if err != nil {
		return Output{Error: fmt.Sprintf("error creating image: %v", err)}
	}

	return Output{ImageURL: imageURL.(string)}
}
