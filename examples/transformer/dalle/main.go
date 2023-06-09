package main

import (
	"context"

	"github.com/henomis/lingoose/transformer"
)

func main() {

	d := transformer.NewDallE().WithImageSize(transformer.DallEImageSize1024).AsFile("test.png")

	_, err := d.Transform(context.Background(), "a goose working with pipelines")
	if err != nil {
		panic(err)
	}
}
