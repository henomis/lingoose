package main

import (
	"context"

	"github.com/henomis/lingoose/transformer"
)

func main() {

	d := transformer.NewHFTextToImage().WithPersistImage("test.png")

	_, err := d.Transform(context.Background(), "A cat over the table.")
	if err != nil {
		panic(err)
	}

}
