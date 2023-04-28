// Package decoder provides a set of decoders to decode the output of a command
package decoder

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"github.com/henomis/lingoose/types"
)

var (
	ErrDecoding = errors.New("decoding output error")
)

type jsonDecoder struct {
	output types.M
}

func NewJSONDecoder() *jsonDecoder {
	return &jsonDecoder{}
}

func (d *jsonDecoder) Decode(input string) (types.M, error) {
	err := json.Unmarshal([]byte(input), &d.output)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrDecoding, err)
	}

	return types.M{
		types.DefaultOutputKey: d.output,
	}, nil
}

type regExDecoder struct {
	output types.M
	regex  string
}

func NewRegExDecoder(regex string) *regExDecoder {
	return &regExDecoder{
		regex: regex,
	}
}

func (d *regExDecoder) Decode(input string) (types.M, error) {
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

	d.output = types.M{
		types.DefaultOutputKey: outputMatches,
	}

	return d.output, nil
}
