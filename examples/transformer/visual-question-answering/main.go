package main

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/transformer"
)

func main() {

	d := transformer.NewHFVisualQuestionAnswering("test.png")

	response, err := d.Transform(context.Background(), "is it wearing glasses?", true)
	if err != nil {
		panic(err)
	}

	fmt.Println(response)
}
