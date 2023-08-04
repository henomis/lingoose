package main

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/loader"
)

func main() {

	l := loader.NewYoutubeDLLoader("https://www.youtube.com/watch?v=--khbXchTeE").WithYoutubeDLPath("/opt/homebrew/bin/youtube-dl")

	docs, err := l.Load(context.Background())
	if err != nil {
		panic(err)
	}

	fmt.Println("Transcription:")
	fmt.Println(docs[0].Content)

	llm := openai.NewCompletion()

	summary, err := llm.Completion(
		context.Background(),
		fmt.Sprintf("Summarize the following text:\n\nTranscription:\n%s", docs[0].Content),
	)

	if err != nil {
		panic(err)
	}

	fmt.Println("Summary:")
	fmt.Println(summary)

}
