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
	O1Mini                Model = openai.O1Mini
	O1Mini20240912        Model = openai.O1Mini20240912
	O1Preview             Model = openai.O1Preview
	O1Preview20240912     Model = openai.O1Preview20240912
	GPT432K0613           Model = openai.GPT432K0613
	GPT432K0314           Model = openai.GPT432K0314
	GPT432K               Model = openai.GPT432K
	GPT40613              Model = openai.GPT40613
	GPT40314              Model = openai.GPT40314
	GPT4o                 Model = openai.GPT4o
	GPT4o20240513         Model = openai.GPT4o20240513
	GPT4o20240806         Model = openai.GPT4o20240806
	GPT4o20241120         Model = openai.GPT4o20241120
	GPT4oLatest           Model = openai.GPT4oLatest
	GPT4oMini             Model = openai.GPT4oMini
	GPT4oMini20240718     Model = openai.GPT4oMini20240718
	GPT4Turbo             Model = openai.GPT4Turbo
	GPT4Turbo20240409     Model = openai.GPT4Turbo20240409
	GPT4Turbo0125         Model = openai.GPT4Turbo0125
	GPT4Turbo1106         Model = openai.GPT4Turbo1106
	GPT4TurboPreview      Model = openai.GPT4TurboPreview
	GPT4VisionPreview     Model = openai.GPT4VisionPreview
	GPT4                  Model = openai.GPT4
	GPT3Dot5Turbo0125     Model = openai.GPT3Dot5Turbo0125
	GPT3Dot5Turbo1106     Model = openai.GPT3Dot5Turbo1106
	GPT3Dot5Turbo0613     Model = openai.GPT3Dot5Turbo0613
	GPT3Dot5Turbo0301     Model = openai.GPT3Dot5Turbo0301
	GPT3Dot5Turbo16K      Model = openai.GPT3Dot5Turbo16K
	GPT3Dot5Turbo16K0613  Model = openai.GPT3Dot5Turbo16K0613
	GPT3Dot5Turbo         Model = openai.GPT3Dot5Turbo
	GPT3Dot5TurboInstruct Model = openai.GPT3Dot5TurboInstruct
	// Deprecated: Model is shutdown. Use gpt-3.5-turbo-instruct instead.
	//lint:ignore SA1019 retained until removed by go-openai for backwards compatibility
	//nolint:staticcheck
	GPT3TextDavinci003 Model = openai.GPT3TextDavinci003
	// Deprecated: Model is shutdown. Use gpt-3.5-turbo-instruct instead.
	//lint:ignore SA1019 retained until removed by go-openai for backwards compatibility
	//nolint:staticcheck
	GPT3TextDavinci002 Model = openai.GPT3TextDavinci002
	// Deprecated: Model is shutdown. Use gpt-3.5-turbo-instruct instead.
	//lint:ignore SA1019 retained until removed by go-openai for backwards compatibility
	//nolint:staticcheck
	GPT3TextCurie001 Model = openai.GPT3TextCurie001
	// Deprecated: Model is shutdown. Use gpt-3.5-turbo-instruct instead.
	//lint:ignore SA1019 retained until removed by go-openai for backwards compatibility
	//nolint:staticcheck
	GPT3TextBabbage001 Model = openai.GPT3TextBabbage001
	// Deprecated: Model is shutdown. Use gpt-3.5-turbo-instruct instead.
	//lint:ignore SA1019 retained until removed by go-openai for backwards compatibility
	//nolint:staticcheck
	GPT3TextAda001 Model = openai.GPT3TextAda001
	// Deprecated: Model is shutdown. Use gpt-3.5-turbo-instruct instead.
	//lint:ignore SA1019 retained until removed by go-openai for backwards compatibility
	//nolint:staticcheck
	GPT3TextDavinci001 Model = openai.GPT3TextDavinci001
	// Deprecated: Model is shutdown. Use gpt-3.5-turbo-instruct instead.
	//lint:ignore SA1019 retained until removed by go-openai for backwards compatibility
	//nolint:staticcheck
	GPT3DavinciInstructBeta Model = openai.GPT3DavinciInstructBeta
	// Deprecated: Model is shutdown. Use davinci-002 instead.
	//lint:ignore SA1019 retained until removed by go-openai for backwards compatibility
	//nolint:staticcheck
	GPT3Davinci    Model = openai.GPT3Davinci
	GPT3Davinci002 Model = openai.GPT3Davinci002
	// Deprecated: Model is shutdown. Use gpt-3.5-turbo-instruct instead.
	//lint:ignore SA1019 retained until removed by go-openai for backwards compatibility
	//nolint:staticcheck
	GPT3CurieInstructBeta Model = openai.GPT3CurieInstructBeta
	GPT3Curie             Model = openai.GPT3Curie
	GPT3Curie002          Model = openai.GPT3Curie002
	// Deprecated: Model is shutdown. Use babbage-002 instead.
	//lint:ignore SA1019 retained until removed by go-openai for backwards compatibility
	//nolint:staticcheck
	GPT3Ada    Model = openai.GPT3Ada
	GPT3Ada002 Model = openai.GPT3Ada002
	// Deprecated: Model is shutdown. Use babbage-002 instead.
	//lint:ignore SA1019 retained until removed by go-openai for backwards compatibility
	//nolint:staticcheck
	GPT3Babbage    Model = openai.GPT3Babbage
	GPT3Babbage002 Model = openai.GPT3Babbage002
)

type UsageCallback func(types.Meta)
type StreamCallback func(string)

type ResponseFormat = openai.ChatCompletionResponseFormatType

const (
	ResponseFormatJSONObject ResponseFormat = openai.ChatCompletionResponseFormatTypeJSONObject
	ResponseFormatText       ResponseFormat = openai.ChatCompletionResponseFormatTypeText
)
