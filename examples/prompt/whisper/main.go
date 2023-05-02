package main

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/prompt"
)

func main() {

	prompt, err := prompt.NewPromptFromAudioFile(
		context.Background(),
		"hello.mp3",
		prompt.AudioResponseFormatVTT,
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(prompt)

}
