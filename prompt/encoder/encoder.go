package encoder

import "fmt"

var (
	ErrInvalidType = fmt.Errorf("invalid type")
)

type Encoder interface {
	Encode(interface{}) error
}

type EncoderFn func(interface{}) Encoder

type stringEncoder string

func (s *stringEncoder) Encode(v *string) error {
	*s = stringEncoder(*v)
	return nil
}

func NewStringEncoderFn() EncoderFn {
	return func(v interface{}) Encoder {
		return (*stringEncoder)(nil)
	}
}
