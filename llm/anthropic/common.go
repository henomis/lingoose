package anthropic

import (
	"fmt"

	anthropicsdk "github.com/anthropics/anthropic-sdk-go"
	"github.com/henomis/lingoose/types"
)

var (
	// ErrAnthropicCompletion is returned when there's an error with the Anthropic completion API
	ErrAnthropicCompletion = fmt.Errorf("anthropic completion error")

	// ErrAnthropicChat is returned when there's an error with the Anthropic chat API
	ErrAnthropicChat = fmt.Errorf("anthropic chat error")
)

const (
	// DefaultAnthropicMaxTokens is the default value for max tokens in Anthropic requests
	DefaultAnthropicMaxTokens = 256

	// DefaultAnthropicTemperature is the default temperature value for Anthropic requests
	DefaultAnthropicTemperature = 0.7

	// DefaultAnthropicNumResults is the default number of results to return
	DefaultAnthropicNumResults = 1

	// DefaultAnthropicTopP is the default top_p value for Anthropic requests
	DefaultAnthropicTopP = 1.0

	// DefaultMaxIterations is the default maximum number of iterations for function calling
	DefaultMaxIterations = 3
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

type ResponseFormat string

const (
	// ResponseFormatJSONObject tells the model to output a JSON object
	ResponseFormatJSONObject ResponseFormat = "json_object"

	// ResponseFormatText tells the model to output plain text (default)
	ResponseFormatText ResponseFormat = "text"
)
