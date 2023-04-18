package decoder

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var (
	ErrInvalidType = fmt.Errorf("invalid type")
)

type Decoder interface {
	Decode(interface{}) error
}

type DecoderFn func(string) Decoder

func NewJsonDecoderFn() DecoderFn {
	return func(s string) Decoder {
		return json.NewDecoder(strings.NewReader(s))
	}
}

type stringDecoder string

func (d *stringDecoder) Decode(v interface{}) error {

	_, ok := v.(*string)
	if !ok {
		return ErrInvalidType
	}

	*v.(*string) = string(*d)
	return nil
}

func NewStringDecoderFn() DecoderFn {
	return func(s string) Decoder {
		return (*stringDecoder)(&s)
	}
}

type regexDecoder struct {
	regex string
	value string
}

func (d *regexDecoder) Decode(v interface{}) error {

	_, ok := v.(*[]string)
	if !ok {
		return ErrInvalidType
	}

	re := regexp.MustCompile(d.regex) // Prepare our regex
	matches := re.FindStringSubmatch(d.value)

	for i, match := range matches {
		if i == 0 {
			continue
		}
		*v.(*[]string) = append(*v.(*[]string), match)
		i++
	}

	return nil
}

func NewRegexDecoderFn(regex string) DecoderFn {
	return func(s string) Decoder {
		return &regexDecoder{
			regex: regex,
			value: s,
		}
	}
}
