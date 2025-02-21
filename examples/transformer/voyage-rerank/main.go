package main

import (
	"context"
	"fmt"

	"github.com/rsest/lingoose/document"
	"github.com/rsest/lingoose/transformer"
)

func main() {

	r := transformer.NewVoyageRerank()

	documents, err := r.Rerank(
		context.Background(),
		"What is the capital of the United States?",
		[]document.Document{
			{
				Content: "Carson City is the capital city of the American state of Nevada.",
			}, {
				Content: "Washington, D.C. (also known as simply Washington or D.C., and officially as the District of Columbia) is the capital of the United States. It is a federal district.",
			}, {
				Content: "Capital punishment (the death penalty) has existed in the United States since beforethe United States was a country. As of 2017, capital punishment is legal in 30 of the 50 states.",
			},
		},
	)
	if err != nil {
		panic(err)
	}

	for _, doc := range documents {
		fmt.Println(doc.GetEnrichedContent())
		fmt.Println("-----")
	}
}
