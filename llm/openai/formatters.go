package openai

import (
	"github.com/rsest/lingoose/thread"
	"github.com/sashabaranov/go-openai"
)

//nolint:gocognit
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
		case thread.RoleUser, thread.RoleSystem:
			if data, isUserTextData := message.Contents[0].Data.(string); isUserTextData {
				chatCompletionMessages[i].Content = data
			} else {
				continue
			}
		case thread.RoleAssistant:
			if data, isAssistantTextData := message.Contents[0].Data.(string); isAssistantTextData {
				chatCompletionMessages[i].Content = data
			} else if data, isTollCallData := message.Contents[0].Data.([]thread.ToolCallData); isTollCallData {
				var toolCalls []openai.ToolCall
				for _, toolCallData := range data {
					toolCalls = append(toolCalls, openai.ToolCall{
						ID:   toolCallData.ID,
						Type: "function",
						Function: openai.FunctionCall{
							Name:      toolCallData.Name,
							Arguments: toolCallData.Arguments,
						},
					})
				}
				chatCompletionMessages[i].ToolCalls = toolCalls
			} else {
				continue
			}
		case thread.RoleTool:
			if data, isTollResponseData := message.Contents[0].Data.(thread.ToolResponseData); isTollResponseData {
				chatCompletionMessages[i].ToolCallID = data.ID
				chatCompletionMessages[i].Name = data.Name
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
			contentAsString, ok := content.Data.(string)
			if !ok {
				continue
			}

			chatMessagePart = &openai.ChatMessagePart{
				Type: openai.ChatMessagePartTypeText,
				Text: contentAsString,
			}
		case thread.ContentTypeImage:
			contentAsString, ok := content.Data.(string)
			if !ok {
				continue
			}

			chatMessagePart = &openai.ChatMessagePart{
				Type: openai.ChatMessagePartTypeImageURL,
				ImageURL: &openai.ChatMessageImageURL{
					URL:    contentAsString,
					Detail: openai.ImageURLDetailAuto,
				},
			}
		case thread.ContentTypeToolCall, thread.ContentTypeToolResponse:
			continue
		default:
			continue
		}

		chatMessageParts[i] = *chatMessagePart
	}

	return chatMessageParts
}

func toolCallResultToThreadMessage(toolCall openai.ToolCall, result string) *thread.Message {
	return thread.NewToolMessage().AddContent(
		thread.NewToolResponseContent(
			thread.ToolResponseData{
				ID:     toolCall.ID,
				Name:   toolCall.Function.Name,
				Result: result,
			},
		),
	)
}

func toolCallsToToolCallMessage(toolCalls []openai.ToolCall) *thread.Message {
	if len(toolCalls) == 0 {
		return nil
	}

	var toolCallData []thread.ToolCallData
	for _, toolCall := range toolCalls {
		toolCallData = append(toolCallData, thread.ToolCallData{
			ID:        toolCall.ID,
			Name:      toolCall.Function.Name,
			Arguments: toolCall.Function.Arguments,
		})
	}

	return thread.NewAssistantMessage().AddContent(
		thread.NewToolCallContent(
			toolCallData,
		),
	)
}
