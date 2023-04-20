package llm

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type LlmMock struct {
}

// getRandomStrings returns a random selection of strings from the data slice
func getRandomStrings(number int) []string {

	// pre-configured data slice with 30 random english words
	data := []string{"air", "fly", "ball", "kite", "tree", "grass", "house", "ocean", "river", "lake", "road", "bridge", "mountain", "valley", "desert", "flower", "wind", "book", "table", "chair", "television", "computer", "window", "door", "cup", "plate", "spoon", "fork", "knife", "bottle", "glass"}

	rand.Seed(time.Now().UnixNano())

	// create empty slice
	result := []string{}

	// append randomly selected strings from data to result
	for i := 0; i < number; i++ {
		result = append(result, data[rand.Intn(len(data))])
	}

	return result
}

func (l *LlmMock) Completion(prompt string) (string, error) {
	fmt.Printf("User: %s\n", prompt)

	// generate random number between 1 and 5
	rand.Seed(time.Now().UnixNano())
	number := rand.Intn(3) + 3

	// get random strings
	randomStrings := getRandomStrings(number)
	output := strings.Join(randomStrings, " ")

	fmt.Printf("AI: %s\n", output)

	var llmResponse interface{}
	_ = llmResponse // llm response

	return output, nil
}

// func (l *LlmMock) Chat(chat chat.Chat) (interface{}, error) {
// 	return nil, nil
// }

type JsonLllMock struct{}

func (l *JsonLllMock) Completion(prompt string) (string, error) {
	fmt.Printf("User: %s\n", prompt)

	// generate random number between 1 and 5
	rand.Seed(time.Now().UnixNano())

	// get random strings

	output := `{"first": "` + strings.Join(getRandomStrings(rand.Intn(5)+1), " ") + `", "second": "` +
		strings.Join(getRandomStrings(rand.Intn(5)+1), " ") + `"}`

	fmt.Printf("AI: %s\n", output)

	var llmResponse interface{}
	_ = llmResponse // llm response

	return output, nil
}
