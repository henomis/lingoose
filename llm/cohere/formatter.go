package cohere

import (
	"github.com/henomis/cohere-go/model"
	"github.com/henomis/cohere-go/request"
	"github.com/rsest/lingoose/thread"
)

var threadRoleToCohereRole = map[thread.Role]model.ChatMessageRole{
	thread.RoleSystem:    model.ChatMessageRoleChatbot,
	thread.RoleUser:      model.ChatMessageRoleUser,
	thread.RoleAssistant: model.ChatMessageRoleChatbot,
}

func (c *Cohere) buildChatCompletionRequest(t *thread.Thread) *request.Chat {
	message, history := threadToChatMessages(t)

	return &request.Chat{
		Model:       c.model,
		ChatHistory: history,
		Message:     message,
	}
}

func threadToChatMessages(t *thread.Thread) (string, []model.ChatMessage) {
	var history []model.ChatMessage
	var message string

	for _, m := range t.Messages {
		chatMessage := model.ChatMessage{
			Role: threadRoleToCohereRole[m.Role],
		}

		switch m.Role {
		case thread.RoleUser, thread.RoleSystem, thread.RoleAssistant:
			for _, content := range m.Contents {
				if content.Type == thread.ContentTypeText {
					chatMessage.Message += content.Data.(string) + "\n"
				}
			}
		case thread.RoleTool:
			continue
		}

		history = append(history, chatMessage)
	}

	lastMessage := t.LastMessage()
	if lastMessage.Role == thread.RoleUser {
		for _, content := range lastMessage.Contents {
			if content.Type == thread.ContentTypeText {
				message += content.Data.(string) + "\n"
			}
		}

		history = history[:len(history)-1]
	}

	return message, history
}
