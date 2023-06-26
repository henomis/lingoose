// Package chat provides a chat prompt template.
// Sometimes you need to define a chat prompt, this package provides a way to do that.
package chat

import (
	"errors"
	"fmt"

	"github.com/henomis/lingoose/types"
)

var (
	ErrChatMessages = errors.New("unable to convert chat messages")
)

type Chat struct {
	promptMessages PromptMessages
}

type Prompt interface {
	String() string
	Format(input types.M) error
}

type MessageType string

const (
	MessageTypeSystem    MessageType = "system"
	MessageTypeUser      MessageType = "user"
	MessageTypeAssistant MessageType = "assistant"
	MessageTypeFunction  MessageType = "function"
)

type PromptMessage struct {
	Type   MessageType
	Prompt Prompt
}

type PromptMessages []PromptMessage

type Message struct {
	Type    MessageType
	Content string
}

type Messages []Message

func New(promptMessages ...PromptMessage) *Chat {
	chatPromptTemplate := &Chat{}
	for _, message := range promptMessages {
		chatPromptTemplate.addMessagePromptTemplate(message)
	}

	return chatPromptTemplate
}

func (p *Chat) addMessagePromptTemplate(message PromptMessage) {
	p.promptMessages = append(p.promptMessages, message)
}

// ToMessages converts the chat prompt template to a list of messages.
func (p *Chat) ToMessages() (Messages, error) {
	var messages Messages
	var err error

	for _, messagePromptTemplate := range p.promptMessages {
		var message Message
		message.Type = messagePromptTemplate.Type

		if messagePromptTemplate.Prompt != nil {
			if len(messagePromptTemplate.Prompt.String()) == 0 {
				err = messagePromptTemplate.Prompt.Format(types.M{})
				if err != nil {
					return nil, fmt.Errorf("%s: %w", ErrChatMessages, err)
				}
			}
			message.Content = messagePromptTemplate.Prompt.String()
		}

		messages = append(messages, message)
	}

	return messages, nil
}

func (c *Chat) PromptMessages() PromptMessages {
	return c.promptMessages
}
