package anthropic

import (
	"encoding/json"

	anthropicsdk "github.com/anthropics/anthropic-sdk-go"
	"github.com/henomis/lingoose/thread"
)

//nolint:gocognit
func threadToChatCompletionMessages(t *thread.Thread) []anthropicsdk.Message {
	chatCompletionMessages := make([]anthropicsdk.Message, len(t.Messages))
	for i, message := range t.Messages {
		chatCompletionMessages[i] = anthropicsdk.Message{
			Role: anthropicsdk.MessageRole(message.Role),
		}

		if len(message.Contents) > 1 {
			continue
		}

		switch message.Role {
		case thread.RoleUser, thread.RoleSystem:
			if data, isUserTextData := message.Contents[0].Data.(string); isUserTextData {
				chatCompletionMessages[i].Content = []anthropicsdk.ContentBlock{{Text: data}}
			} else {
				continue
			}
		case thread.RoleAssistant:
			if data, isAssistantTextData := message.Contents[0].Data.(string); isAssistantTextData {
				chatCompletionMessages[i].Content = []anthropicsdk.ContentBlock{{Text: data}}
			} else if toolCallDataSlice, isToolCallData := message.Contents[0].Data.([]thread.ToolCallData); isToolCallData {
				var toolUses []anthropicsdk.ContentBlock
				for _, toolCallData := range toolCallDataSlice {
					toolUses = append(toolUses, anthropicsdk.ContentBlock{
						Type:  "tool_use",
						Input: json.RawMessage(toolCallData.Arguments),
					})
				}
				chatCompletionMessages[i].Content = toolUses
			} else {
				continue
			}
		case thread.RoleTool:
			if data, isToolResponseData := message.Contents[0].Data.(thread.ToolResponseData); isToolResponseData {
				chatCompletionMessages[i].Content = []anthropicsdk.ContentBlock{{Text: data.Result}}
			} else {
				continue
			}
		}
	}

	return chatCompletionMessages
}

func toolCallResultToThreadMessage(toolCall anthropicsdk.ToolParam, result string) *thread.Message {
	return thread.NewToolMessage().AddContent(
		thread.NewToolResponseContent(
			thread.ToolResponseData{
				Name:   toolCall.Name.Value,
				Result: result,
			},
		),
	)
}
