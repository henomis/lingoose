package main

import (
	"context"

	"github.com/henomis/lingoose/legacy/pipeline"
	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/prompt"
	"github.com/henomis/lingoose/types"
)

func main() {

	iterator := 0
	languages := []string{"italian", "spanish"}
	sentence := "hello world"

	translate := pipeline.NewTube(
		pipeline.Llm{
			LlmMode:   pipeline.LlmModeCompletion,
			LlmEngine: openai.NewCompletion().WithVerbose(true),
			Prompt:    prompt.NewPromptTemplate("translate hello in {{.language}} the following sentence: \"{{.sentence}}\""),
		},
	)

	expand := pipeline.NewTube(
		pipeline.Llm{
			LlmMode:   pipeline.LlmModeCompletion,
			LlmEngine: openai.NewCompletion().WithVerbose(true),
			Prompt:    prompt.NewPromptTemplate("expand the sentence \"{{.output}}\" adding more details."),
		},
	)

	translatePreCallback := pipeline.Callback(func(ctx context.Context, input types.M) (types.M, error) {
		input["language"] = languages[iterator]
		input["sentence"] = sentence
		return input, nil
	})

	expandPostCallback := pipeline.Callback(func(ctx context.Context, output types.M) (types.M, error) {
		iterator++
		if iterator >= len(languages) {
			pipeline.SetNextTubeExit(output)
		} else {
			pipeline.SetNextTube(output, 0)
		}
		return output, nil
	})

	pipeLine := pipeline.New(translate, expand).WithPreCallbacks(translatePreCallback, nil).WithPostCallbacks(nil, expandPostCallback)

	pipeLine.Run(context.Background(), nil)

}
