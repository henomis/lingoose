package main

import (
	"context"
	"fmt"

	"github.com/rsest/lingoose/transformer"
)

func main() {

	d := transformer.NewDallE().WithImageSize(transformer.DallEImageSize1024x1024)

	imageURL, err := d.Transform(context.Background(), "a goose working with pipelines")
	if err != nil {
		panic(err)
	}

	fmt.Println("Image created:", imageURL)
}
