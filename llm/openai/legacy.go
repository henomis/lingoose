// Package openai provides a wrapper around the OpenAI API.
package openai

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/rsest/lingoose/legacy/chat"
	"github.com/rsest/lingoose/llm/cache"
	"github.com/rsest/lingoose/types"
	"github.com/sashabaranov/go-openai"
)

type Legacy struct {
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
func (o *Legacy) WithModel(model Model) *Legacy {
	o.model = model
	return o
}

// WithTemperature sets the temperature to use for the OpenAI instance.
func (o *Legacy) WithTemperature(temperature float32) *Legacy {
	o.temperature = temperature
	return o
}

// WithMaxTokens sets the max tokens to use for the OpenAI instance.
func (o *Legacy) WithMaxTokens(maxTokens int) *Legacy {
	o.maxTokens = maxTokens
	return o
}

// WithUsageCallback sets the usage callback to use for the OpenAI instance.
func (o *Legacy) WithCallback(callback UsageCallback) *Legacy {
	o.usageCallback = callback
	return o
}

// WithStop sets the stop sequences to use for the OpenAI instance.
func (o *Legacy) WithStop(stop []string) *Legacy {
	o.stop = stop
	return o
}

// WithClient sets the client to use for the OpenAI instance.
func (o *Legacy) WithClient(client *openai.Client) *Legacy {
	o.openAIClient = client
	return o
}

// WithVerbose sets the verbose flag to use for the OpenAI instance.
func (o *Legacy) WithVerbose(verbose bool) *Legacy {
	o.verbose = verbose
	return o
}

// WithCache sets the cache to use for the OpenAI instance.
func (o *Legacy) WithCompletionCache(cache *cache.Cache) *Legacy {
	o.cache = cache
	return o
}

// SetStop sets the stop sequences for the completion.
func (o *Legacy) SetStop(stop []string) {
	o.stop = stop
}

func (o *Legacy) setUsageMetadata(usage openai.Usage) {
	callbackMetadata := make(types.Meta)

	err := mapstructure.Decode(usage, &callbackMetadata)
	if err != nil {
		return
	}

	o.usageCallback(callbackMetadata)
}

func NewLegacy(model Model, temperature float32, maxTokens int, verbose bool) *Legacy {
	openAIKey := os.Getenv("OPENAI_API_KEY")

	return &Legacy{
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
func (o *Legacy) CalledFunctionName() *string {
	return o.calledFunctionName
}

// FinishReason returns the LLM finish reason.
func (o *Legacy) FinishReason() string {
	return o.finishReason
}

func NewCompletion() *Legacy {
	return NewLegacy(
		GPT3Dot5TurboInstruct,
		DefaultOpenAITemperature,
		DefaultOpenAIMaxTokens,
		false,
	)
}

func NewChat() *Legacy {
	return NewLegacy(
		GPT3Dot5Turbo,
		DefaultOpenAITemperature,
		DefaultOpenAIMaxTokens,
		false,
	)
}

// Completion returns a single completion for the given prompt.
func (o *Legacy) Completion(ctx context.Context, prompt string) (string, error) {
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
func (o *Legacy) BatchCompletion(ctx context.Context, prompts []string) ([]string, error) {
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
func (o *Legacy) CompletionStream(ctx context.Context, callbackFn StreamCallback, prompt string) error {
	return o.BatchCompletionStream(ctx, []StreamCallback{callbackFn}, []string{prompt})
}

// BatchCompletionStream returns multiple completion streams for the given prompts.
func (o *Legacy) BatchCompletionStream(ctx context.Context, callbackFn []StreamCallback, prompts []string) error {
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
func (o *Legacy) Chat(ctx context.Context, prompt *chat.Chat) (string, error) {
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
		//nolint:staticcheck
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
func (o *Legacy) ChatStream(ctx context.Context, callbackFn StreamCallback, prompt *chat.Chat) error {
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
