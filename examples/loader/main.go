package main

import (
	"fmt"

	"github.com/henomis/lingoose/loader"
)

func main() {

	p := loader.NewPubmedLoader([]string{"33024307", "32265180"})

	docs, err := p.Load()
	if err != nil {
		panic(err)
	}

	for _, doc := range docs {
		fmt.Println(doc.Content)
		fmt.Println("------")
		fmt.Println(doc.Metadata)
		fmt.Println("------")
	}

}
