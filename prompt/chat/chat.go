// Package chat provides a chat prompt template.
// Sometimes you need to define a chat prompt, this package provides a way to do that.
package chat

import "github.com/henomis/lingoose/prompt"

type Chat struct {
	messagesPrompt []PromptMessage
}

type MessageType string

const (
	MessageTypeSystem MessageType = "system"
	MessageTypeUser   MessageType = "user"
	MessageTypeAI     MessageType = "ai"
)

type PromptMessage struct {
	Type   MessageType
	Prompt *prompt.Prompt
}

type PromptMessages []PromptMessage

type Message struct {
	Type    MessageType
	Content string
}

type Messages []Message

func New(messages PromptMessages) *Chat {
	chatPromptTemplate := &Chat{}
	for _, message := range messages {
		chatPromptTemplate.AddMessagePromptTemplate(message)
	}

	return chatPromptTemplate
}

func (p *Chat) AddMessagePromptTemplate(message PromptMessage) {
	p.messagesPrompt = append(p.messagesPrompt, message)
}

// ToMessages converts the chat prompt template to a list of messages.
func (p *Chat) ToMessages() (Messages, error) {
	var messages Messages
	var err error

	for _, messagePromptTemplate := range p.messagesPrompt {
		var message Message
		message.Type = messagePromptTemplate.Type

		if messagePromptTemplate.Prompt != nil {
			message.Content, err = messagePromptTemplate.Prompt.Format()
			if err != nil {
				return nil, err
			}
		}

		messages = append(messages, message)
	}

	return messages, nil
}
