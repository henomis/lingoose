package summarizepipeline

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/document"
	"github.com/henomis/lingoose/legacy/pipeline"
	"github.com/henomis/lingoose/prompt"
	"github.com/henomis/lingoose/types"
)

type Loader interface {
	Load(ctx context.Context) ([]document.Document, error)
}

func New(llmEngine pipeline.LlmEngine, loader Loader) *pipeline.Pipeline {
	docs := []document.Document{}
	iterator := 0
	remainigDocs := 0

	summarizeLLM := pipeline.Llm{
		LlmEngine: llmEngine,
		LlmMode:   pipeline.LlmModeCompletion,
		Prompt:    prompt.NewPromptTemplate(summaryPrompt),
	}

	refineLLM := pipeline.Llm{
		LlmEngine: llmEngine,
		LlmMode:   pipeline.LlmModeCompletion,
		Prompt:    prompt.NewPromptTemplate(refinePrompt),
	}

	summary := pipeline.NewTube(summarizeLLM)
	preSummaryCB := pipeline.Callback(func(ctx context.Context, input types.M) (types.M, error) {
		var err error
		docs, err = loader.Load(ctx)
		if err != nil {
			return nil, err
		}

		if len(docs) == 0 {
			return nil, fmt.Errorf("no documents to summarize")
		}
		iterator = 0
		remainigDocs = len(docs)
		return types.M{
			"text": docs[iterator].Content,
		}, nil
	})
	postSummaryCB := pipeline.Callback(func(ctx context.Context, output types.M) (types.M, error) {
		remainigDocs--
		iterator++
		if remainigDocs == 0 {
			output[pipeline.NextTubeKey] = pipeline.NextTubeExit
		}

		return output, nil
	})

	refine := pipeline.NewTube(refineLLM)
	preRefineCB := pipeline.Callback(func(ctx context.Context, input types.M) (types.M, error) {
		input["text"] = docs[iterator].Content
		return input, nil
	})

	postRefineCB := pipeline.Callback(func(ctx context.Context, output types.M) (types.M, error) {
		remainigDocs--
		iterator++
		if remainigDocs == 0 {
			output[pipeline.NextTubeKey] = pipeline.NextTubeExit
		} else {
			output[pipeline.NextTubeKey] = 1
		}
		return output, nil
	})

	summarizePipeline := pipeline.New(summary, refine).
		WithPreCallbacks(preSummaryCB, preRefineCB).WithPostCallbacks(postSummaryCB, postRefineCB)

	return summarizePipeline
}
