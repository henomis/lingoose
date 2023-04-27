package pipeline

import (
	"context"
	"fmt"
	"sync"

	"github.com/henomis/lingoose/types"
)

type Splitter struct {
	name       string
	llm        Llm
	decoder    Decoder
	memory     Memory
	splitterFn SplitterFn
}

type SplitterFn func(input types.M) ([]types.M, error)

func NewSplitter(
	name string,
	llm Llm,
	outputDecoder Decoder,
	memory Memory,
	splitterFn SplitterFn,
) *Splitter {
	return &Splitter{
		name:       name,
		llm:        llm,
		decoder:    outputDecoder,
		memory:     memory,
		splitterFn: splitterFn,
	}
}

func (s *Splitter) Run(ctx context.Context, input types.M) (types.M, error) {

	splittedInputs, err := s.splitterFn(input)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	var sm sync.Mutex
	pipeOutpus := []types.M{}

	for i, splittedInput := range splittedInputs {
		wg.Add(1)
		go func(i int, splittedInput types.M) {
			defer wg.Done()
			step := NewTube(
				fmt.Sprintf("%s-%d", s.name, i),
				s.llm,
				s.decoder,
				s.memory,
			)

			output, err := step.Run(ctx, splittedInput)
			if err != nil {
				return
			}

			sm.Lock()
			pipeOutpus = append(pipeOutpus, output)
			sm.Unlock()
		}(i, splittedInput)
	}

	wg.Wait()

	return types.M{types.DefaultOutputKey: pipeOutpus}, nil

}
