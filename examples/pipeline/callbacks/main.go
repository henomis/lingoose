package main

import (
	"context"

	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/pipeline"
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

	cbTranslate := pipeline.PipelineCallback(func(output types.M) (types.M, error) {
		iterator++
		return output, nil
	})

	cbExpand := pipeline.PipelineCallback(func(output types.M) (types.M, error) {

		if iterator >= len(languages) {
			pipeline.SetNextTubeExit(output)
		} else {
			pipeline.SetNextTube(output, 0)
			output["language"] = languages[iterator]
			output["sentence"] = sentence
		}

		return output, nil
	})

	pipeLine := pipeline.New(translate, expand).WithCallbacks(cbTranslate, cbExpand)

	pipeLine.Run(context.Background(), types.M{"sentence": sentence, "language": languages[iterator]})

}
