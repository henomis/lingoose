// Package openai provides a wrapper around the OpenAI API.
package openai

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/henomis/lingoose/chat"
	"github.com/henomis/lingoose/llm/cache"
	"github.com/henomis/lingoose/types"
	"github.com/mitchellh/mapstructure"
	"github.com/sashabaranov/go-openai"
)

type OpenAILegacy struct {
	openAIClient           *openai.Client
	model                  Model
	temperature            float32
	maxTokens              int
	stop                   []string
	verbose                bool
	usageCallback          UsageCallback
	functions              map[string]Function
	functionsMaxIterations uint
	calledFunctionName     *string
	finishReason           string
	cache                  *cache.Cache
}

// WithModel sets the model to use for the OpenAI instance.
func (o *OpenAILegacy) WithModel(model Model) *OpenAILegacy {
	o.model = model
	return o
}

// WithTemperature sets the temperature to use for the OpenAI instance.
func (o *OpenAILegacy) WithTemperature(temperature float32) *OpenAILegacy {
	o.temperature = temperature
	return o
}

// WithMaxTokens sets the max tokens to use for the OpenAI instance.
func (o *OpenAILegacy) WithMaxTokens(maxTokens int) *OpenAILegacy {
	o.maxTokens = maxTokens
	return o
}

// WithUsageCallback sets the usage callback to use for the OpenAI instance.
func (o *OpenAILegacy) WithCallback(callback UsageCallback) *OpenAILegacy {
	o.usageCallback = callback
	return o
}

// WithStop sets the stop sequences to use for the OpenAI instance.
func (o *OpenAILegacy) WithStop(stop []string) *OpenAILegacy {
	o.stop = stop
	return o
}

// WithClient sets the client to use for the OpenAI instance.
func (o *OpenAILegacy) WithClient(client *openai.Client) *OpenAILegacy {
	o.openAIClient = client
	return o
}

// WithVerbose sets the verbose flag to use for the OpenAI instance.
func (o *OpenAILegacy) WithVerbose(verbose bool) *OpenAILegacy {
	o.verbose = verbose
	return o
}

// WithCache sets the cache to use for the OpenAI instance.
func (o *OpenAILegacy) WithCompletionCache(cache *cache.Cache) *OpenAILegacy {
	o.cache = cache
	return o
}

// SetStop sets the stop sequences for the completion.
func (o *OpenAILegacy) SetStop(stop []string) {
	o.stop = stop
}

func (o *OpenAILegacy) setUsageMetadata(usage openai.Usage) {
	callbackMetadata := make(types.Meta)

	err := mapstructure.Decode(usage, &callbackMetadata)
	if err != nil {
		return
	}

	o.usageCallback(callbackMetadata)
}

func NewLegacy(model Model, temperature float32, maxTokens int, verbose bool) *OpenAILegacy {
	openAIKey := os.Getenv("OPENAI_API_KEY")

	return &OpenAILegacy{
		openAIClient:           openai.NewClient(openAIKey),
		model:                  model,
		temperature:            temperature,
		maxTokens:              maxTokens,
		verbose:                verbose,
		functions:              make(map[string]Function),
		functionsMaxIterations: DefaultMaxIterations,
	}
}

// CalledFunctionName returns the name of the function that was called.
func (o *OpenAILegacy) CalledFunctionName() *string {
	return o.calledFunctionName
}

// FinishReason returns the LLM finish reason.
func (o *OpenAILegacy) FinishReason() string {
	return o.finishReason
}

func NewCompletion() *OpenAILegacy {
	return NewLegacy(
		GPT3Dot5TurboInstruct,
		DefaultOpenAITemperature,
		DefaultOpenAIMaxTokens,
		false,
	)
}

func NewChat() *OpenAILegacy {
	return NewLegacy(
		GPT3Dot5Turbo,
		DefaultOpenAITemperature,
		DefaultOpenAIMaxTokens,
		false,
	)
}

// Completion returns a single completion for the given prompt.
func (o *OpenAILegacy) Completion(ctx context.Context, prompt string) (string, error) {
	var cacheResult *cache.Result
	var err error

	if o.cache != nil {
		cacheResult, err = o.cache.Get(ctx, prompt)
		if err == nil {
			return strings.Join(cacheResult.Answer, "\n"), nil
		} else if !errors.Is(err, cache.ErrCacheMiss) {
			return "", fmt.Errorf("%w: %w", ErrOpenAICompletion, err)
		}
	}

	outputs, err := o.BatchCompletion(ctx, []string{prompt})
	if err != nil {
		return "", err
	}

	if o.cache != nil {
		err = o.cache.Set(ctx, cacheResult.Embedding, outputs[0])
		if err != nil {
			return "", fmt.Errorf("%w: %w", ErrOpenAICompletion, err)
		}
	}

	return outputs[0], nil
}

// BatchCompletion returns multiple completions for the given prompts.
func (o *OpenAILegacy) BatchCompletion(ctx context.Context, prompts []string) ([]string, error) {
	response, err := o.openAIClient.CreateCompletion(
		ctx,
		openai.CompletionRequest{
			Model:       string(o.model),
			Prompt:      prompts,
			MaxTokens:   o.maxTokens,
			Temperature: o.temperature,
			N:           DefaultOpenAINumResults,
			TopP:        DefaultOpenAITopP,
			Stop:        o.stop,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrOpenAICompletion, err)
	}

	if o.usageCallback != nil {
		o.setUsageMetadata(response.Usage)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("%w: no choices returned", ErrOpenAICompletion)
	}

	var outputs []string
	for _, choice := range response.Choices {
		index := choice.Index
		outputs = append(outputs, strings.TrimSpace(choice.Text))
		if o.verbose {
			debugCompletion(prompts[index], choice.Text)
		}
	}

	return outputs, nil
}

// CompletionStream returns a single completion stream for the given prompt.
func (o *OpenAILegacy) CompletionStream(ctx context.Context, callbackFn StreamCallback, prompt string) error {
	return o.BatchCompletionStream(ctx, []StreamCallback{callbackFn}, []string{prompt})
}

// BatchCompletionStream returns multiple completion streams for the given prompts.
func (o *OpenAILegacy) BatchCompletionStream(ctx context.Context, callbackFn []StreamCallback, prompts []string) error {
	stream, err := o.openAIClient.CreateCompletionStream(
		ctx,
		openai.CompletionRequest{
			Model:       string(o.model),
			Prompt:      prompts,
			MaxTokens:   o.maxTokens,
			Temperature: o.temperature,
			N:           DefaultOpenAINumResults,
			TopP:        DefaultOpenAITopP,
			Stop:        o.stop,
		},
	)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrOpenAICompletion, err)
	}

	defer stream.Close()

	for {
		response, errRecv := stream.Recv()
		if errors.Is(errRecv, io.EOF) {
			break
		}

		if errRecv != nil {
			return fmt.Errorf("%w: %w", ErrOpenAICompletion, errRecv)
		}

		if o.usageCallback != nil {
			o.setUsageMetadata(response.Usage)
		}

		if len(response.Choices) == 0 {
			return fmt.Errorf("%w: no choices returned", ErrOpenAICompletion)
		}

		for _, choice := range response.Choices {
			index := choice.Index
			output := choice.Text
			if o.verbose {
				debugCompletion(prompts[index], output)
			}

			callbackFn[index](output)
		}
	}

	return nil
}

// Chat returns a single chat completion for the given prompt.
func (o *OpenAILegacy) Chat(ctx context.Context, prompt *chat.Chat) (string, error) {
	messages, err := buildMessages(prompt)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrOpenAIChat, err)
	}

	chatCompletionRequest := openai.ChatCompletionRequest{
		Model:       string(o.model),
		Messages:    messages,
		MaxTokens:   o.maxTokens,
		Temperature: o.temperature,
		N:           DefaultOpenAINumResults,
		TopP:        DefaultOpenAITopP,
		Stop:        o.stop,
	}

	if len(o.functions) > 0 {
		chatCompletionRequest.Functions = o.getFunctions()
	}

	response, err := o.openAIClient.CreateChatCompletion(
		ctx,
		chatCompletionRequest,
	)

	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrOpenAIChat, err)
	}

	if o.usageCallback != nil {
		o.setUsageMetadata(response.Usage)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("%w: no choices returned", ErrOpenAIChat)
	}

	content := response.Choices[0].Message.Content

	o.finishReason = string(response.Choices[0].FinishReason)
	o.calledFunctionName = nil
	if response.Choices[0].FinishReason == "function_call" && len(o.functions) > 0 {
		if o.verbose {
			fmt.Printf("Calling function %s\n", response.Choices[0].Message.FunctionCall.Name)
			fmt.Printf("Function call arguments: %s\n", response.Choices[0].Message.FunctionCall.Arguments)
		}

		content, err = o.functionCall(response)
		if err != nil {
			return "", fmt.Errorf("%w: %w", ErrOpenAIChat, err)
		}
	}

	if o.verbose {
		debugChat(prompt, content)
	}

	return content, nil
}

// ChatStream returns a single chat stream for the given prompt.
func (o *OpenAILegacy) ChatStream(ctx context.Context, callbackFn StreamCallback, prompt *chat.Chat) error {
	messages, err := buildMessages(prompt)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrOpenAIChat, err)
	}

	stream, err := o.openAIClient.CreateChatCompletionStream(
		ctx,
		openai.ChatCompletionRequest{
			Model:       string(o.model),
			Messages:    messages,
			MaxTokens:   o.maxTokens,
			Temperature: o.temperature,
			N:           DefaultOpenAINumResults,
			TopP:        DefaultOpenAITopP,
			Stop:        o.stop,
		},
	)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrOpenAIChat, err)
	}

	for {
		response, errRecv := stream.Recv()
		if errors.Is(errRecv, io.EOF) {
			break
		}

		// oops no usage here?
		// if o.usageCallback != nil {
		// 	o.setUsageMetadata(response.Usage)
		// }

		if len(response.Choices) == 0 {
			return fmt.Errorf("%w: no choices returned", ErrOpenAIChat)
		}

		content := response.Choices[0].Delta.Content

		if o.verbose {
			debugChat(prompt, content)
		}

		callbackFn(content)
	}

	return nil
}

func buildMessages(prompt *chat.Chat) ([]openai.ChatCompletionMessage, error) {
	var messages []openai.ChatCompletionMessage

	promptMessages, err := prompt.ToMessages()
	if err != nil {
		return nil, err
	}

	for _, message := range promptMessages {
		if message.Type == chat.MessageTypeUser {
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: message.Content,
			})
		} else if message.Type == chat.MessageTypeAssistant {
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: message.Content,
			})
		} else if message.Type == chat.MessageTypeSystem {
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleSystem,
				Content: message.Content,
			})
		} else if message.Type == chat.MessageTypeFunction {
			fnmessage := openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleFunction,
				Content: message.Content,
			}
			if message.Name != nil {
				fnmessage.Name = *message.Name
			}

			messages = append(messages, fnmessage)
		}
	}

	return messages, nil
}

func debugChat(prompt *chat.Chat, content string) {
	promptMessages, err := prompt.ToMessages()
	if err != nil {
		return
	}

	for _, message := range promptMessages {
		if message.Type == chat.MessageTypeUser {
			fmt.Printf("---USER---\n%s\n", message.Content)
		} else if message.Type == chat.MessageTypeAssistant {
			fmt.Printf("---AI---\n%s\n", message.Content)
		} else if message.Type == chat.MessageTypeSystem {
			fmt.Printf("---SYSTEM---\n%s\n", message.Content)
		} else if message.Type == chat.MessageTypeFunction {
			fmt.Printf("---FUNCTION---\n%s()\n%s\n", *message.Name, message.Content)
		}
	}
	fmt.Printf("---AI---\n%s\n", content)
}

func debugCompletion(prompt string, content string) {
	fmt.Printf("---USER---\n%s\n", prompt)
	fmt.Printf("---AI---\n%s\n", content)
}
