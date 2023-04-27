package pipeline

import (
	"context"
)

type Funnel struct {
	name     string
	llm      Llm
	decoder  Decoder
	memory   Memory
	funnelFn FunnelFn
}

type FunnelFn func(input []map[string]interface{}) (interface{}, error)

func NewFunnel(
	name string,
	llm Llm,
	outputDecoder Decoder,
	memory Memory,
	funnelFn FunnelFn,
) *Funnel {
	return &Funnel{
		name:     name,
		llm:      llm,
		decoder:  outputDecoder,
		memory:   memory,
		funnelFn: funnelFn,
	}
}

func (s *Funnel) Run(ctx context.Context, inputs interface{}) (interface{}, error) {

	input, err := s.funnelFn(inputs.([]map[string]interface{}))
	if err != nil {
		return nil, err
	}

	step := NewStep(
		s.name,
		s.llm,
		s.decoder,
		s.memory,
	)

	return step.Run(ctx, input)

}
