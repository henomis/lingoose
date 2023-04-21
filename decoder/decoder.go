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

type DefaultDecoder struct {
	output interface{}
}

func (d *DefaultDecoder) Decode(input string) (interface{}, error) {
	d.output = map[string]interface{}{
		"output": input,
	}

	return d.output, nil
}

func NewDefaultDecoder() *DefaultDecoder {
	return &DefaultDecoder{}
}

type JSONDecoder struct {
	output interface{}
}

func NewJSONDecoder(output interface{}) *JSONDecoder {
	return &JSONDecoder{
		output: output,
	}
}

func (d *JSONDecoder) Decode(input string) (interface{}, error) {
	err := json.Unmarshal([]byte(input), d.output)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrDecoding, err)
	}

	return d.output, nil
}

type RegExDecoder struct {
	output interface{}
	regex  string
}

func NewRegExDecoder(regex string) *RegExDecoder {
	return &RegExDecoder{
		regex: regex,
	}
}

func (d *RegExDecoder) Decode(input string) (interface{}, error) {
	re, err := regexp.Compile(d.regex)
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

	d.output = outputMatches

	return d.output, nil
}
