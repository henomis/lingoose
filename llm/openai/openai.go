// Package openai provides a wrapper around the OpenAI API.
package openai

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/henomis/lingoose/chat"
	"github.com/henomis/lingoose/types"
	"github.com/mitchellh/mapstructure"
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
)

type Model string

const (
	GPT4               Model = openai.GPT4
	GPT3Dot5Turbo      Model = openai.GPT3Dot5Turbo
	GPT3TextDavinci003 Model = openai.GPT3TextDavinci003
	GPT3TextDavinci002 Model = openai.GPT3TextDavinci002
	GPT3TextCurie001   Model = openai.GPT3TextCurie001
	GPT3TextBabbage001 Model = openai.GPT3TextBabbage001
	GPT3TextAda001     Model = openai.GPT3TextAda001
	GPT3TextDavinci001 Model = openai.GPT3TextDavinci001
	GPT3Davinci        Model = openai.GPT3Davinci
	GPT3Curie          Model = openai.GPT3Curie
	GPT3Ada            Model = openai.GPT3Ada
	GPT3Babbage        Model = openai.GPT3Babbage
)

type OpenAICallback func(types.Meta)

type openAI struct {
	openAIClient *openai.Client
	model        Model
	temperature  float32
	maxTokens    int
	stop         []string
	verbose      bool
	callback     OpenAICallback
}

func New(model Model, temperature float32, maxTokens int, verbose bool) (*openAI, error) {

	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not set")
	}

	return &openAI{
		openAIClient: openai.NewClient(openAIKey),
		model:        model,
		temperature:  temperature,
		maxTokens:    maxTokens,
		verbose:      verbose,
	}, nil
}

func (o *openAI) WithStop(stop []string) *openAI {
	o.stop = stop
	return o
}

func NewCompletion() (*openAI, error) {
	return New(
		GPT3TextDavinci003,
		DefaultOpenAITemperature,
		DefaultOpenAIMaxTokens,
		false,
	)
}

func NewChat() (*openAI, error) {
	return New(
		GPT3Dot5Turbo,
		DefaultOpenAITemperature,
		DefaultOpenAIMaxTokens,
		false,
	)
}

func (o *openAI) Completion(ctx context.Context, prompt string) (string, error) {

	completionRequest := openai.CompletionRequest{
		Model:       string(o.model),
		Prompt:      prompt,
		MaxTokens:   o.maxTokens,
		Temperature: o.temperature,
		N:           DefaultOpenAINumResults,
		TopP:        DefaultOpenAITopP,
	}

	if len(o.stop) > 0 {
		completionRequest.Stop = o.stop
	}

	response, err := o.openAIClient.CreateCompletion(ctx, completionRequest)

	if err != nil {
		return "", fmt.Errorf("%s: %w", ErrOpenAICompletion, err)
	}

	if o.callback != nil {
		o.setMetadata(response.Usage)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("%s: no choices returned", ErrOpenAICompletion)
	}

	output := strings.TrimSpace(response.Choices[0].Text)
	if o.verbose {
		debugCompletion(prompt, output)
	}

	return output, nil
}

func (o *openAI) Chat(ctx context.Context, prompt *chat.Chat) (string, error) {

	messages, err := buildMessages(prompt)
	if err != nil {
		return "", fmt.Errorf("%s: %w", ErrOpenAIChat, err)
	}

	chatRequest := openai.ChatCompletionRequest{
		Model:       string(o.model),
		Messages:    messages,
		MaxTokens:   o.maxTokens,
		Temperature: o.temperature,
		N:           DefaultOpenAINumResults,
		TopP:        DefaultOpenAITopP,
	}

	if len(o.stop) > 0 {
		chatRequest.Stop = o.stop
	}

	response, err := o.openAIClient.CreateChatCompletion(ctx, chatRequest)

	if err != nil {
		return "", fmt.Errorf("%s: %w", ErrOpenAIChat, err)
	}

	if o.callback != nil {
		o.setMetadata(response.Usage)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("%s: no choices returned", ErrOpenAIChat)
	}

	content := response.Choices[0].Message.Content

	if o.verbose {
		debugChat(prompt, content)
	}

	return content, nil
}

func (o *openAI) SetCallback(callback OpenAICallback) {
	o.callback = callback
}

func (o *openAI) setMetadata(usage openai.Usage) {

	callbackMetadata := make(types.Meta)

	err := mapstructure.Decode(usage, &callbackMetadata)
	if err != nil {
		return
	}

	o.callback(callbackMetadata)
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
		}
	}
	fmt.Printf("---AI---\n%s\n", content)
}

func debugCompletion(prompt string, content string) {
	fmt.Printf("---USER---\n%s\n", prompt)
	fmt.Printf("---AI---\n%s\n", content)
}
