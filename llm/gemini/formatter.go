package gemini

import (
	"cloud.google.com/go/vertexai/genai"
	"fmt"
	"github.com/henomis/lingoose/thread"
	"strings"
)

func threadToPartMessage(t *thread.Thread) []genai.Part {
	var chatMessages []genai.Part

	//msgToModel = system prompts + user utterance
	for _, m := range t.Messages {
		switch m.Role {
		case thread.RoleUser, thread.RoleSystem:
			for _, content := range m.Contents {
				contentData, ok := content.Data.(string)
				if !ok {
					continue
				}
				chatMessages = append(chatMessages, genai.Text(contentData))
			}

		case thread.RoleAssistant:
			continue
		case thread.RoleTool:
			continue
		}
	}
	return chatMessages
}

func partsTostring(parts []genai.Part) string {
	var msg strings.Builder
	size := len(parts) - 1
	for i := 0; i < len(parts); i++ {
		switch parts[i].(type) {
		case genai.Text:
			msg.WriteString(fmt.Sprintf("%v", parts[i]))
			if i != size {
				msg.WriteString(" ")
			}
		}
	}
	return msg.String()
}

func functionToolCallsToToolCallMessage(toolCalls []genai.FunctionCall) *thread.Message {
	if len(toolCalls) == 0 {
		return nil
	}

	var toolCallData []thread.ToolCallData
	for _, toolCall := range toolCalls {
		toolCallData = append(toolCallData, thread.ToolCallData{
			Name: toolCall.Name,
		})
	}

	return thread.NewAssistantMessage().AddContent(
		thread.NewToolCallContent(
			toolCallData,
		),
	)
}

func toolCallResultToThreadMessage(fnCall genai.FunctionCall, result string) *thread.Message {
	return thread.NewToolMessage().AddContent(
		thread.NewToolResponseContent(
			thread.ToolResponseData{
				Name:   fnCall.Name,
				Result: result,
			},
		),
	)
}
