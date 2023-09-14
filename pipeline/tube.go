package pipeline

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/chat"
	"github.com/henomis/lingoose/types"
	"github.com/mitchellh/mapstructure"
)

type Tube struct {
	llm       Llm
	decoder   Decoder
	namespace string
	memory    Memory
	history   History
}

func NewTube(
	llm Llm,
) *Tube {
	return &Tube{
		llm:     llm,
		decoder: &defaultDecoder{},
	}
}

func (t *Tube) Namespace() string {
	return t.namespace
}

func (t *Tube) WithMemory(namespace string, memory Memory) *Tube {
	t.namespace = namespace
	t.memory = memory
	return t
}

func (t *Tube) WithHistory(history History) *Tube {
	t.history = history
	return t
}

func (t *Tube) WithDecoder(decoder Decoder) *Tube {
	t.decoder = decoder
	return t
}

// Run execute the step and return the output.
// The prompt is formatted with the input and the output of the prompt is used as input for the LLM.
// If the step has a memory, the output is stored in the memory.
func (t *Tube) Run(ctx context.Context, input types.M) (types.M, error) {

	if input == nil {
		input = types.M{}
	}

	input, err := structToMap(input)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrDecoding, err)
	}

	if t.memory != nil {
		input = mergeMaps(input, t.memory.All())
	}

	response, err := t.executeLLM(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrLLMExecution, err)
	}

	decodedOutput, err := t.decoder.Decode(response)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrDecoding, err)
	}

	if t.memory != nil {
		err = t.memory.Set(t.namespace, decodedOutput)
		if err != nil {
			return nil, err
		}
	}

	return decodedOutput, nil

}

func (t *Tube) executeLLM(ctx context.Context, input types.M) (string, error) {
	if t.llm.LlmMode == LlmModeCompletion {
		return t.executeLLMCompletion(ctx, input)
	} else if t.llm.LlmMode == LlmModeChat {
		return t.executeLLMChat(ctx, input)
	}

	return "", ErrInvalidLmmMode
}

func (t *Tube) executeLLMCompletion(ctx context.Context, input types.M) (string, error) {
	err := t.llm.Prompt.Format(input)
	if err != nil {
		return "", err
	}

	if t.history != nil {
		err = t.history.Add(t.llm.Prompt.String(), nil)
		if err != nil {
			return "", err
		}
	}

	response, err := t.llm.LlmEngine.Completion(ctx, t.llm.Prompt.String())
	if err != nil {
		return "", err
	}

	if t.history != nil {
		err = t.history.Add(response, nil)
		if err != nil {
			return "", err
		}
	}

	return response, nil
}

func (t *Tube) executeLLMChat(ctx context.Context, input types.M) (string, error) {

	for _, promptMessage := range t.llm.Chat.PromptMessages() {
		err := promptMessage.Prompt.Format(input)
		if err != nil {
			return "", err
		}

		if t.history != nil {
			err = t.history.Add(
				promptMessage.Prompt.String(),
				types.Meta{
					"role": promptMessage.Type,
				},
			)
			if err != nil {
				return "", err
			}
		}
	}

	response, err := t.llm.LlmEngine.Chat(ctx, t.llm.Chat)
	if err != nil {
		return "", err
	}

	if t.history != nil {
		err = t.history.Add(
			response,
			types.Meta{
				"role": chat.MessageTypeAssistant,
			},
		)
		if err != nil {
			return "", err
		}
	}

	return response, nil
}

func mergeMaps(m1 types.M, m2 types.M) types.M {
	merged := make(types.M)
	for k, v := range m1 {
		merged[k] = v
	}
	for key, value := range m2 {
		merged[key] = value
	}
	return merged
}

func structToMap(obj interface{}) (types.M, error) {
	genericMap := types.M{}
	err := mapstructure.Decode(obj, &genericMap)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrDecoding, err)
	}

	return genericMap, nil
}
