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
			if data, isTollResponseData := m.Contents[0].Data.(thread.ToolResponseData); isTollResponseData && !m.Contents[0].Processed {
				var funcResponses genai.FunctionResponse
				funcResponses.Name = data.Name
				funcResponses.Response = map[string]any{
					"result": data.Result,
				}
				chatMessages = append(chatMessages, funcResponses)
			}
		}
	}
	return chatMessages
}

func threadToChatPartMessage(t *thread.Thread) ([]genai.Part, error) {
	var (
		chatMessages []genai.Part
		//chatHistory  []*genai.Content
	)

	for _, m := range t.Messages {
		switch m.Role {
		case thread.RoleUser, thread.RoleSystem:
			for _, content := range m.Contents {
				contentData, ok := content.Data.(string)
				if !ok || content.Processed {
					continue
				}
				chatMessages = append(chatMessages, genai.Text(contentData))
				content.Processed = true
			}

		case thread.RoleAssistant:
			continue
		case thread.RoleTool:
			if data, isTollResponseData := m.Contents[0].Data.(thread.ToolResponseData); isTollResponseData && !m.Contents[0].Processed {
				var funcResponses genai.FunctionResponse
				funcResponses.Name = data.Name
				funcResponses.Response = map[string]any{
					"result": data.Result,
				}
				chatMessages = append(chatMessages, funcResponses)
				m.Contents[0].Processed = true
			}
		}
	}

	//msgToModel = system prompts + user utterance
	//for _, m := range t.Messages {
	//
	//	//process last msg as chat message
	//	if m == t.LastMessage() && (m.Role == thread.RoleUser || m.Role == thread.RoleSystem) {
	//		break
	//	}
	//
	//	switch m.Role {
	//	case thread.RoleUser, thread.RoleSystem:
	//		role := threadRoleToGeminiRole[thread.RoleUser]
	//		formChatHistory(role, m, chatHistory)
	//
	//	case thread.RoleAssistant:
	//		assistantRole := threadRoleToGeminiRole[thread.RoleAssistant]
	//		formChatHistory(assistantRole, m, chatHistory)
	//
	//	case thread.RoleTool:
	//		continue
	//	}
	//}

	//userMsg := LastUserMessage(t)
	//if userMsg != nil {
	//	for _, content := range userMsg.Contents {
	//		contentData, ok := content.Data.(string)
	//		if !ok {
	//			continue
	//		}
	//		chatMessages = append(chatMessages, genai.Text(contentData))
	//	}
	//}

	if len(chatMessages) == 0 {
		return nil, fmt.Errorf("%w", ErrGeminiNoChat)
	}
	return chatMessages, nil
}

func PartsTostring(parts []genai.Part) string {
	var msg strings.Builder
	size := len(parts) - 1
	for i := 0; i < len(parts); i++ {
		switch parts[i].(type) {
		case genai.Text:
			msg.WriteString(fmt.Sprintf("%v", parts[i]))
			if i != size {
				msg.WriteString(" ")
			}
		case genai.FunctionCall:
			fp := parts[i].(genai.FunctionCall)
			msg.WriteString(fmt.Sprintf("FunctionCall: %+v ", fp))

		case genai.FunctionResponse:
			fp := parts[i].(genai.FunctionResponse)
			msg.WriteString(fmt.Sprintf("FunctionResponse: %+v ", fp))
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

func formChatHistory(role string, m *thread.Message, ch []*genai.Content) {
	chatContent := &genai.Content{
		Role: role,
	}
	for _, content := range m.Contents {
		contentData, ok := content.Data.(string)
		if !ok {
			continue
		}
		chatContent.Parts = append(chatContent.Parts, genai.Text(contentData))
	}
	ch = append(ch, chatContent)
}

// LastUserMessage returns last user or assistant message in the thread
func LastUserMessage(t *thread.Thread) *thread.Message {
	for i := len(t.Messages) - 1; i >= 0; i-- {
		if t.Messages[i].Role == thread.RoleUser || t.Messages[i].Role == thread.RoleAssistant {
			return t.Messages[i]
		}
	}
	return nil
}
