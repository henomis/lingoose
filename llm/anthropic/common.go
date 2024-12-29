package anthropic

import (
	"fmt"

	anthropicsdk "github.com/anthropics/anthropic-sdk-go"
	"github.com/henomis/lingoose/types"
	"github.com/sashabaranov/go-openai"
)

var (
	ErrAnthropicCompletion = fmt.Errorf("anthropic completion error")
	ErrAnthropicChat       = fmt.Errorf("anthropic chat error")
)

const (
	DefaultAnthropicMaxTokens   = 256
	DefaultAnthropicTemperature = 0.7
	DefaultAnthropicNumResults  = 1
	DefaultAnthropicTopP        = 1.0
	DefaultMaxIterations        = 3
)

type Model string

const (
	ModelClaude3_5HaikuLatest       Model = Model(anthropicsdk.ModelClaude3_5HaikuLatest)
	ModelClaude3_5Haiku20241022     Model = Model(anthropicsdk.ModelClaude3_5Haiku20241022)
	ModelClaude3_5SonnetLatest      Model = Model(anthropicsdk.ModelClaude3_5SonnetLatest)
	ModelClaude3_5Sonnet20241022    Model = Model(anthropicsdk.ModelClaude3_5Sonnet20241022)
	ModelClaude_3_5_Sonnet_20240620 Model = Model(anthropicsdk.ModelClaude_3_5_Sonnet_20240620)
	ModelClaude3OpusLatest          Model = Model(anthropicsdk.ModelClaude3OpusLatest)
	ModelClaude_3_Opus_20240229     Model = Model(anthropicsdk.ModelClaude_3_Opus_20240229)
	ModelClaude_3_Sonnet_20240229   Model = Model(anthropicsdk.ModelClaude_3_Sonnet_20240229)
	ModelClaude_3_Haiku_20240307    Model = Model(anthropicsdk.ModelClaude_3_Haiku_20240307)
	ModelClaude_2_1                 Model = Model(anthropicsdk.ModelClaude_2_1)
	ModelClaude_2_0                 Model = Model(anthropicsdk.ModelClaude_2_0)
)

type UsageCallback func(types.Meta)
type StreamCallback func(string)

type ResponseFormat = openai.ChatCompletionResponseFormatType

const (
	ResponseFormatJSONObject ResponseFormat = openai.ChatCompletionResponseFormatTypeJSONObject
	ResponseFormatText       ResponseFormat = openai.ChatCompletionResponseFormatTypeText
)
