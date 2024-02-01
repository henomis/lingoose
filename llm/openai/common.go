package openai

import (
	"fmt"

	"github.com/henomis/lingoose/types"
	"github.com/sashabaranov/go-openai"
)

var (
	ErrOpenAICompletion = fmt.Errorf("openai completion error")
	ErrOpenAIChat       = fmt.Errorf("openai chat error")
)

const (
	DefaultOpenAIMaxTokens   = 256
	DefaultOpenAITemperature = 0.7
	DefaultOpenAINumResults  = 1
	DefaultOpenAITopP        = 1.0
	DefaultMaxIterations     = 3
)

type Model string

const (
	GPT432K0613           Model = openai.GPT432K0613
	GPT432K0314           Model = openai.GPT432K0314
	GPT432K               Model = openai.GPT432K
	GPT40613              Model = openai.GPT40613
	GPT40314              Model = openai.GPT40314
	GPT4Turbo0125         Model = openai.GPT4Turbo0125
	GPT4Turbo1106         Model = openai.GPT4Turbo1106
	GPT4TurboPreview      Model = openai.GPT4TurboPreview
	GPT4VisionPreview     Model = openai.GPT4VisionPreview
	GPT4                  Model = openai.GPT4
	GPT3Dot5Turbo1106     Model = openai.GPT3Dot5Turbo1106
	GPT3Dot5Turbo0613     Model = openai.GPT3Dot5Turbo0613
	GPT3Dot5Turbo0301     Model = openai.GPT3Dot5Turbo0301
	GPT3Dot5Turbo16K      Model = openai.GPT3Dot5Turbo16K
	GPT3Dot5Turbo16K0613  Model = openai.GPT3Dot5Turbo16K0613
	GPT3Dot5Turbo         Model = openai.GPT3Dot5Turbo
	GPT3Dot5TurboInstruct Model = openai.GPT3Dot5TurboInstruct
	GPT3Davinci           Model = openai.GPT3Davinci
	GPT3Davinci002        Model = openai.GPT3Davinci002
	GPT3Curie             Model = openai.GPT3Curie
	GPT3Curie002          Model = openai.GPT3Curie002
	GPT3Ada               Model = openai.GPT3Ada
	GPT3Ada002            Model = openai.GPT3Ada002
	GPT3Babbage           Model = openai.GPT3Babbage
	GPT3Babbage002        Model = openai.GPT3Babbage002
)

type UsageCallback func(types.Meta)
type StreamCallback func(string)
