package llmmock

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/henomis/lingoose/chat"
)

type LlmMock struct {
}

func (l *LlmMock) Completion(prompt string) (string, error) {
	fmt.Printf("User: %s\n", prompt)

	rand.Seed(time.Now().UnixNano())
	number := rand.Intn(3) + 3

	randomStrings := getRandomStrings(number)
	output := strings.Join(randomStrings, " ")

	fmt.Printf("AI: %s\n", output)

	var llmResponse interface{}
	_ = llmResponse

	return output, nil
}

func (l *LlmMock) Chat(prompt *chat.Chat) (string, error) {

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

	rand.Seed(time.Now().UnixNano())
	number := rand.Intn(3) + 3

	randomStrings := getRandomStrings(number)
	output := strings.Join(randomStrings, " ")

	fmt.Printf("AI: %s\n", output)

	var llmResponse interface{}
	_ = llmResponse

	return output, nil
}

type JsonLllMock struct{}

func (l *JsonLllMock) Completion(prompt string) (string, error) {
	fmt.Printf("User: %s\n", prompt)

	rand.Seed(time.Now().UnixNano())
	output := `{"first": "` + strings.Join(getRandomStrings(rand.Intn(5)+1), " ") + `", "second": "` +
		strings.Join(getRandomStrings(rand.Intn(5)+1), " ") + `"}`

	fmt.Printf("AI: %s\n", output)

	var llmResponse interface{}
	_ = llmResponse

	return output, nil
}

// getRandomStrings returns a random selection of strings from the data slice.
// this function has been generate by AI! ;)
func getRandomStrings(number int) []string {

	data := []string{"air", "fly", "ball", "kite", "tree", "grass", "house", "ocean", "river", "lake", "road", "bridge", "mountain", "valley", "desert", "flower", "wind", "book", "table", "chair", "television", "computer", "window", "door", "cup", "plate", "spoon", "fork", "knife", "bottle", "glass"}

	rand.Seed(time.Now().UnixNano())

	result := []string{}

	for i := 0; i < number; i++ {
		result = append(result, data[rand.Intn(len(data))])
	}

	return result
}

func (l *JsonLllMock) Chat(prompt *chat.Chat) (string, error) {

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

	rand.Seed(time.Now().UnixNano())

	output := `{"first": "` + strings.Join(getRandomStrings(rand.Intn(5)+1), " ") + `", "second": "` +
		strings.Join(getRandomStrings(rand.Intn(5)+1), " ") + `"}`

	fmt.Printf("AI: %s\n", output)

	var llmResponse interface{}
	_ = llmResponse

	return output, nil
}
