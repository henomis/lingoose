package main

import (
	"fmt"

	"github.com/henomis/lingoose/prompt"
)

func main() {

	promptTemplate, err := prompt.New(nil, nil, "Hello World")
	if err != nil {
		panic(err)
	}
	output, _ := promptTemplate.Format()
	fmt.Println(output)

}
