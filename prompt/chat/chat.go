// Package chat provides a chat prompt template.
// Sometimes you need to define a chat prompt, this package provides a way to do that.
package chat

import "github.com/henomis/lingoose/prompt/template"

type Chat struct {
	messagesPromptTemplate []MessageTemplate
}

type MessageType string

const (
	MessageTypeSystem MessageType = "system"
	MessageTypeUser   MessageType = "user"
	MessageTypeAI     MessageType = "ai"
)

type MessageTemplate struct {
	Type     MessageType
	Template *template.Prompt
}

type Message struct {
	Type    MessageType
	Content string
}

func New(messages []MessageTemplate) *Chat {
	chatPromptTemplate := &Chat{}
	for _, message := range messages {
		chatPromptTemplate.AddMessagePromptTemplate(message)
	}

	return chatPromptTemplate
}

func (p *Chat) AddMessagePromptTemplate(message MessageTemplate) {
	p.messagesPromptTemplate = append(p.messagesPromptTemplate, message)
}

// ToMessages converts the chat prompt template to a list of messages.
func (p *Chat) ToMessages(inputs template.Inputs) ([]Message, error) {
	var messages []Message

	for _, messagePromptTemplate := range p.messagesPromptTemplate {
		var message Message
		message.Type = messagePromptTemplate.Type

		if messagePromptTemplate.Template != nil {
			var err error
			selectedInputs := filterInputs(inputs, messagePromptTemplate.Template.InputsSet())
			message.Content, err = messagePromptTemplate.Template.Format(selectedInputs)
			if err != nil {
				return nil, err
			}
		}

		messages = append(messages, message)
	}

	return messages, nil
}

func filterInputs(inputs template.Inputs, inputsSet map[string]struct{}) template.Inputs {
	selectedInputs := make(template.Inputs)

	for input, value := range inputs {
		if _, ok := inputsSet[input]; ok {
			selectedInputs[input] = value
		}
	}

	return selectedInputs
}
