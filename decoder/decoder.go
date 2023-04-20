package decoder

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
)

var (
	ErrDecoding = errors.New("decoding output error")
)

type Decoder func(string, interface{}) (interface{}, error)

// NewDefaultDecoder is the default decoder that returns a map with a single key "output"
func NewDefaultDecoder() Decoder {
	return func(input string, output interface{}) (interface{}, error) {
		return map[string]interface{}{
			"output": input,
		}, nil
	}
}

// NewJSONDecoder is a decoder that decodes a JSON string into a struct
func NewJSONDecoder() Decoder {
	return func(input string, output interface{}) (interface{}, error) {
		err := json.Unmarshal([]byte(input), output)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", ErrDecoding, err)
		}
		return output, nil
	}
}

// NewRegExDecoder is a decoder that decodes a string into a slice of strings using a regular expression
func NewRegExDecoder(regex string) Decoder {
	return func(input string, output interface{}) (interface{}, error) {

		re, err := regexp.Compile(regex)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", ErrDecoding, err)
		}
		matches := re.FindStringSubmatch(input)

		outputMatches := []string{}
		for i, match := range matches {
			if i == 0 {
				continue
			}
			outputMatches = append(outputMatches, match)
		}

		return outputMatches, nil
	}
}
