package pipeline

import "github.com/henomis/lingoose/types"

type Decoder interface {
	Decode(input string) (types.M, error)
}

type defaultDecoder struct {
	output types.M
}

func (d *defaultDecoder) Decode(input string) (types.M, error) {
	d.output = types.M{
		types.DefaultOutputKey: input,
	}

	return d.output, nil
}
