package prompt

type ChatPromptTemplate struct {
	messagesPromptTemplate []MessagePromptTemplate
}

type MessageType string

const (
	MessageTypeSystem MessageType = "system"
	MessageTypeUser   MessageType = "user"
	MessageTypeAI     MessageType = "ai"
)

type MessagePromptTemplate struct {
	Type   MessageType
	Prompt *PromptTemplate
}

type Message struct {
	Type    MessageType
	Content string
}

func (p *ChatPromptTemplate) AddMessagePromptTemplate(message MessagePromptTemplate) {
	p.messagesPromptTemplate = append(p.messagesPromptTemplate, message)
}

func (p *ChatPromptTemplate) ToMessages(inputs Inputs) ([]Message, error) {
	var messages []Message

	for _, messagePromptTemplate := range p.messagesPromptTemplate {
		var message Message
		message.Type = messagePromptTemplate.Type

		if messagePromptTemplate.Prompt != nil {
			var err error
			selectedInputs := filterInputs(inputs, messagePromptTemplate.Prompt.inputsSet)
			message.Content, err = messagePromptTemplate.Prompt.Format(selectedInputs)
			if err != nil {
				return nil, err
			}
		}

		messages = append(messages, message)
	}

	return messages, nil
}

func filterInputs(inputs Inputs, inputsSet map[string]struct{}) Inputs {
	selectedInputs := make(Inputs)

	for input, value := range inputs {
		if _, ok := inputsSet[input]; ok {
			selectedInputs[input] = value
		}
	}

	return selectedInputs
}
