// Package llmmock provides a mock implementation of the LLM interface.
package llmmock

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/henomis/lingoose/legacy/chat"
)

type LlmMock struct {
}

func (l *LlmMock) Completion(ctx context.Context, prompt string) (string, error) {
	_ = ctx
	fmt.Printf("User: %s\n", prompt)

	//nolint:gosec
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	number := r.Intn(3) + 3

	randomStrings := getRandomStrings(number)
	output := strings.Join(randomStrings, " ")

	fmt.Printf("AI: %s\n", output)

	var llmResponse interface{}
	_ = llmResponse

	return output, nil
}

func (l *LlmMock) Chat(ctx context.Context, prompt *chat.Chat) (string, error) {
	_ = ctx
	messages, err := prompt.ToMessages()
	if err != nil {
		return "", err
	}

	for _, message := range messages {
		if message.Type == chat.MessageTypeUser {
			fmt.Printf("User: %s\n", message.Content)
		} else if message.Type == chat.MessageTypeAssistant {
			fmt.Printf("AI: %s\n", message.Content)
		} else if message.Type == chat.MessageTypeSystem {
			fmt.Printf("System: %s\n", message.Content)
		}
	}

	//nolint:gosec
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	number := r.Intn(3) + 3

	randomStrings := getRandomStrings(number)
	output := strings.Join(randomStrings, " ")

	fmt.Printf("AI: %s\n", output)

	var llmResponse interface{}
	_ = llmResponse

	return output, nil
}

type JSONLllMock struct{}

func (l *JSONLllMock) Completion(ctx context.Context, prompt string) (string, error) {
	_ = ctx
	fmt.Printf("User: %s\n", prompt)

	//nolint:gosec
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	//nolint:gosec
	output := `{"first": "` + strings.Join(getRandomStrings(r.Intn(5)+1), " ") + `", "second": "` +
		strings.Join(getRandomStrings(r.Intn(5)+1), " ") + `"}`

	fmt.Printf("AI: %s\n", output)

	var llmResponse interface{}
	_ = llmResponse

	return output, nil
}

// getRandomStrings returns a random selection of strings from the data slice.
// this function has been generate by AI! ;)
func getRandomStrings(number int) []string {
	data := []string{"air", "fly", "ball", "kite", "tree", "grass", "house", "ocean", "river", "lake", "road",
		"bridge", "mountain", "valley", "desert", "flower", "wind", "book", "table", "chair", "television", "computer",
		"window", "door", "cup", "plate", "spoon", "fork", "knife", "bottle", "glass"}

	//nolint:gosec
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	result := []string{}

	for i := 0; i < number; i++ {
		//nolint:gosec
		result = append(result, data[r.Intn(len(data))])
	}

	return result
}

func (l *JSONLllMock) Chat(ctx context.Context, prompt *chat.Chat) (string, error) {
	_ = ctx
	messages, err := prompt.ToMessages()
	if err != nil {
		return "", err
	}

	for _, message := range messages {
		if message.Type == chat.MessageTypeUser {
			fmt.Printf("User: %s\n", message.Content)
		} else if message.Type == chat.MessageTypeAssistant {
			fmt.Printf("AI: %s\n", message.Content)
		} else if message.Type == chat.MessageTypeSystem {
			fmt.Printf("System: %s\n", message.Content)
		}
	}

	//nolint:gosec
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	output := `{"first": "` + strings.Join(getRandomStrings(r.Intn(5)+1), " ") + `", "second": "` +
		strings.Join(getRandomStrings(r.Intn(5)+1), " ") + `"}`

	fmt.Printf("AI: %s\n", output)

	var llmResponse interface{}
	_ = llmResponse

	return output, nil
}
