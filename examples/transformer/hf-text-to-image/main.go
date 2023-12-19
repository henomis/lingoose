package main

import (
	"context"

	"github.com/henomis/lingoose/transformer"
)

func main() {
	t := transformer.NewHFTextToImage().WithPersistImage("test.png")

	_, err := t.Transform(context.Background(), "A cat over the table.")
	if err != nil {
		panic(err)
	}
}
