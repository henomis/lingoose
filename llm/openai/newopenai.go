package openai

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/thread"
	openai "github.com/sashabaranov/go-openai"
)

var threadRoleToOpenAIRole = map[thread.Role]string{
	thread.RoleUser:      "user",
	thread.RoleAssistant: "assistant",
	thread.RoleTool:      "tool",
}

func threadToChatCompletionMessages(t *thread.Thread) []openai.ChatCompletionMessage {
	chatCompletionMessages := make([]openai.ChatCompletionMessage, len(t.Messages))
	for i, message := range t.Messages {
		chatCompletionMessages[i] = openai.ChatCompletionMessage{
			Role: threadRoleToOpenAIRole[message.Role],
		}

		if len(message.Contents) > 1 {
			chatCompletionMessages[i].MultiContent = threadContentsToChatMessageParts(message)
			continue
		}

		switch message.Role {
		case thread.RoleUser, thread.RoleAssistant:
			if data, ok := message.Contents[0].Data.(string); ok {
				chatCompletionMessages[i].Content = data
			} else {
				continue
			}
		case thread.RoleTool:
			if data, ok := message.Contents[0].Data.(thread.ToolData); ok {
				chatCompletionMessages[i].ToolCallID = data.ID
				chatCompletionMessages[i].Content = data.Result
			} else {
				continue
			}
		}
	}

	return chatCompletionMessages
}

func threadContentsToChatMessageParts(m *thread.Message) []openai.ChatMessagePart {
	chatMessageParts := make([]openai.ChatMessagePart, len(m.Contents))

	for i, content := range m.Contents {
		var chatMessagePart *openai.ChatMessagePart

		switch content.Type {
		case thread.ContentTypeText:
			chatMessagePart = &openai.ChatMessagePart{
				Type: openai.ChatMessagePartTypeText,
				Text: content.Data.(string),
			}
		case thread.ContentTypeImage:
			chatMessagePart = &openai.ChatMessagePart{
				Type: openai.ChatMessagePartTypeImageURL,
				ImageURL: &openai.ChatMessageImageURL{
					URL:    content.Data.(string),
					Detail: openai.ImageURLDetailAuto,
				},
			}
		default:
			continue
		}

		chatMessageParts[i] = *chatMessagePart
	}

	return chatMessageParts
}

func (o *OpenAI) Generate(ctx context.Context, t *thread.Thread) (*thread.Thread, error) {
	if t == nil {
		return nil, nil
	}

	chatCompletionRequest := openai.ChatCompletionRequest{
		Model:       string(o.model),
		Messages:    threadToChatCompletionMessages(t),
		MaxTokens:   o.maxTokens,
		Temperature: o.temperature,
		N:           DefaultOpenAINumResults,
		TopP:        DefaultOpenAITopP,
		Stop:        o.stop,
	}

	if len(o.functions) > 0 {
		chatCompletionRequest.Tools = o.getChatCompletionRequestTools()
		chatCompletionRequest.ToolChoice = o.getChatCompletionRequestToolChoice()
	}

	response, err := o.openAIClient.CreateChatCompletion(
		ctx,
		chatCompletionRequest,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrOpenAIChat, err)
	}

	if o.usageCallback != nil {
		o.setUsageMetadata(response.Usage)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("%w: no choices returned", ErrOpenAIChat)
	}

	var messages []*thread.Message
	if response.Choices[0].FinishReason == "tool_calls" {
		messages = o.callTools(response)
	} else {
		if o.verbose {
			//TODO
		}
		messages = []*thread.Message{
			textContentToThreadMessage(response.Choices[0].Message.Content),
		}
	}

	t.Messages = append(t.Messages, messages...)

	return t, nil
}

func (o *OpenAI) getChatCompletionRequestTools() []openai.Tool {
	return o.getFunctions()
}

func (o *OpenAI) getChatCompletionRequestToolChoice() any {
	if o.toolChoice != nil {
		return openai.ToolChoice{
			Type: openai.ToolTypeFunction,
			Function: openai.ToolFunction{
				Name: *o.toolChoice,
			},
		}
	}

	return "auto"
}

func (o *OpenAI) callTool(toolCall openai.ToolCall) (string, error) {
	fn, ok := o.functions[toolCall.Function.Name]
	if !ok {
		return "", fmt.Errorf("unknown function %s", toolCall.Function.Name)
	}

	resultAsJSON, err := callFnWithArgumentAsJSON(fn.Fn, toolCall.Function.Arguments)
	if err != nil {
		return "", err
	}

	o.calledFunctionName = &fn.Name

	return resultAsJSON, nil
}

func (o *OpenAI) callTools(response openai.ChatCompletionResponse) []*thread.Message {
	if len(o.functions) == 0 || len(response.Choices[0].Message.ToolCalls) == 0 {
		return nil
	}

	var messages []*thread.Message
	for _, toolCall := range response.Choices[0].Message.ToolCalls {
		if o.verbose {
			fmt.Printf("Calling function %s\n", toolCall.Function.Name)
			fmt.Printf("Function call arguments: %s\n", toolCall.Function.Arguments)
		}

		result, err := o.callTool(toolCall)
		if err != nil {
			result = fmt.Sprintf("error: %s", err)
		}

		messages = append(messages, toolCallResultToThreadMessage(toolCall, result))
	}

	return messages
}

func toolCallResultToThreadMessage(toolCall openai.ToolCall, result string) *thread.Message {
	return &thread.Message{
		Role: thread.RoleTool,
		Contents: []*thread.Content{
			thread.NewToolContent(
				&thread.ToolData{
					ID:     toolCall.ID,
					Name:   toolCall.Function.Name,
					Result: result,
				},
			),
		},
	}
}

func textContentToThreadMessage(content string) *thread.Message {
	return &thread.Message{
		Role: thread.RoleAssistant,
		Contents: []*thread.Content{
			thread.NewTextContent(content),
		},
	}
}
