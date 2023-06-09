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
func (s *Tube) Run(ctx context.Context, input types.M) (types.M, error) {

	if input == nil {
		input = types.M{}
	}

	input, err := structToMap(input)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrDecoding, err)
	}

	if s.memory != nil {
		input = mergeMaps(input, s.memory.All())
	}

	response, err := s.executeLLM(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrLLMExecution, err)
	}

	decodedOutput, err := s.decoder.Decode(response)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrDecoding, err)
	}

	if s.memory != nil {
		err = s.memory.Set(s.namespace, decodedOutput)
		if err != nil {
			return nil, err
		}
	}

	return decodedOutput, nil

}

func (s *Tube) executeLLM(ctx context.Context, input types.M) (string, error) {
	if s.llm.LlmMode == LlmModeCompletion {
		return s.executeLLMCompletion(ctx, input)
	} else if s.llm.LlmMode == LlmModeChat {
		return s.executeLLMChat(ctx, input)
	}

	return "", ErrInvalidLmmMode
}

func (s *Tube) executeLLMCompletion(ctx context.Context, input types.M) (string, error) {
	err := s.llm.Prompt.Format(input)
	if err != nil {
		return "", err
	}

	if s.history != nil {
		err = s.history.Add(s.llm.Prompt.String(), nil)
		if err != nil {
			return "", err
		}
	}

	response, err := s.llm.LlmEngine.Completion(ctx, s.llm.Prompt.String())
	if err != nil {
		return "", err
	}

	if s.history != nil {
		err = s.history.Add(response, nil)
		if err != nil {
			return "", err
		}
	}

	return response, nil
}

func (s *Tube) executeLLMChat(ctx context.Context, input types.M) (string, error) {

	for _, promptMessage := range s.llm.Chat.PromptMessages() {
		err := promptMessage.Prompt.Format(input)
		if err != nil {
			return "", err
		}

		if s.history != nil {
			err = s.history.Add(
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

	response, err := s.llm.LlmEngine.Chat(ctx, s.llm.Chat)
	if err != nil {
		return "", err
	}

	if s.history != nil {
		err = s.history.Add(
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
