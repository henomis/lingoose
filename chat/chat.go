// Package chat provides a chat prompt template.
// Sometimes you need to define a chat prompt, this package provides a way to do that.
package chat

type Chat struct {
	PromptMessages PromptMessages
}

type Prompt interface {
	Prompt() string
	Format(input interface{}) error
}

type MessageType string

const (
	MessageTypeSystem    MessageType = "system"
	MessageTypeUser      MessageType = "user"
	MessageTypeAssistant MessageType = "assistant"
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

func New(promptMessages PromptMessages) *Chat {
	chatPromptTemplate := &Chat{}
	for _, message := range promptMessages {
		chatPromptTemplate.AddMessagePromptTemplate(message)
	}

	return chatPromptTemplate
}

func (p *Chat) AddMessagePromptTemplate(message PromptMessage) {
	p.PromptMessages = append(p.PromptMessages, message)
}

// ToMessages converts the chat prompt template to a list of messages.
func (p *Chat) ToMessages() (Messages, error) {
	var messages Messages
	var err error

	for _, messagePromptTemplate := range p.PromptMessages {
		var message Message
		message.Type = messagePromptTemplate.Type

		if messagePromptTemplate.Prompt != nil {
			err = messagePromptTemplate.Prompt.Format(map[string]interface{}{})
			if err != nil {
				return nil, err
			}
			message.Content = messagePromptTemplate.Prompt.Prompt()
		}

		messages = append(messages, message)
	}

	return messages, nil
}
