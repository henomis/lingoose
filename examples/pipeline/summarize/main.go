package main

import (
	"context"

	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/loader"
	summarizepipeline "github.com/henomis/lingoose/pipeline/summarize"
	"github.com/henomis/lingoose/textsplitter"
)

// download https://raw.githubusercontent.com/hwchase17/chat-your-data/master/state_of_the_union.txt

func main() {

	summarize := summarizepipeline.New(
		openai.NewCompletion().WithMaxTokens(1000).WithVerbose(true).WithModel(openai.GPT3Dot5TurboInstruct),
		loader.NewTextLoader("state_of_the_union.txt", nil).
			WithTextSplitter(textsplitter.NewRecursiveCharacterTextSplitter(2000, 0)),
	)

	_, err := summarize.Run(context.Background(), nil)
	if err != nil {
		panic(err)
	}
}
