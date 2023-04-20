package decoder

import (
	"encoding/json"
	"regexp"
)

type Decoder func(string, interface{}) (interface{}, error)

func NewDefaultDecoder() Decoder {
	return func(input string, output interface{}) (interface{}, error) {
		output = map[string]interface{}{
			"output": input,
		}
		return output, nil
	}
}

func NewJSONDecoder() Decoder {
	return func(input string, output interface{}) (interface{}, error) {
		err := json.Unmarshal([]byte(input), output)
		return output, err
	}
}

func NewRegExDecoder(regex string) Decoder {
	return func(input string, output interface{}) (interface{}, error) {
		//use regex to parse input
		re, err := regexp.Compile(regex) // Prepare our regex
		if err != nil {
			return nil, err
		}
		matches := re.FindStringSubmatch(input)

		output = []string{}
		for i, match := range matches {
			if i == 0 {
				continue
			}
			output = append(output.([]string), match)
		}

		return output, nil
	}
}
