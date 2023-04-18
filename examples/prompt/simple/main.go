package main

import (
	"fmt"

	"github.com/henomis/lingoose/prompt"
)

func main() {

	promptTemplate, err := prompt.New("Hello World", nil, nil, nil)
	if err != nil {
		panic(err)
	}
	output, _ := promptTemplate.Format()
	fmt.Println(output)

}

func newString(s string) *string {
	return &s
}
