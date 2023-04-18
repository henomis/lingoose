package llm

import (
	"fmt"

	"github.com/henomis/lingoose/prompt"
	"github.com/henomis/lingoose/prompt/chat"
)

type Llm interface {
	Completion(prompt *prompt.Prompt) (interface{}, error)
	Chat(chat chat.Chat) (interface{}, error)
}

type LlmMock struct {
}

func (l *LlmMock) Completion(prompt *prompt.Prompt) (interface{}, error) {
	formattedPrompt, err := prompt.Format()
	if err != nil {
		return nil, err
	}
	_ = formattedPrompt

	fmt.Printf("User: %s\n", formattedPrompt)

	output := "Fine, thanks."

	fmt.Printf("AI: %s\n", output)

	var llmResponse interface{}
	_ = llmResponse // llm response

	// decode output
	err = prompt.OutputDecoder(output).Decode(prompt.Output)
	if err != nil {
		return nil, err
	}

	return llmResponse, err
}

func (l *LlmMock) Chat(chat chat.Chat) (interface{}, error) {
	return nil, nil
}
