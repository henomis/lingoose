package ollama

import "github.com/henomis/lingoose/thread"

func (o *Ollama) buildChatCompletionRequest(t *thread.Thread) *request {
	return &request{
		Model:    o.model,
		Messages: threadToChatMessages(t),
	}
}

func threadToChatMessages(t *thread.Thread) []message {
	chatMessages := make([]message, len(t.Messages))
	for i, m := range t.Messages {
		chatMessages[i] = message{
			Role: threadRoleToOllamaRole[m.Role],
		}

		switch m.Role {
		case thread.RoleUser, thread.RoleSystem, thread.RoleAssistant:
			for _, content := range m.Contents {
				if content.Type == thread.ContentTypeText {
					chatMessages[i].Content += content.Data.(string) + "\n"
				}
			}
		case thread.RoleTool:
			continue
		}
	}

	return chatMessages
}
