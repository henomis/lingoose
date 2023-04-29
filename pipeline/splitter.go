package pipeline

import (
	"context"
	"fmt"
	"sync"

	"github.com/henomis/lingoose/types"
)

var (
	ErrSplitFunction = fmt.Errorf("split function error")
)

type splitter struct {
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
) *splitter {
	return &splitter{
		name:       name,
		llm:        llm,
		decoder:    outputDecoder,
		memory:     memory,
		splitterFn: splitterFn,
	}
}

func (s *splitter) Run(ctx context.Context, input types.M) (types.M, error) {

	splittedInputs, err := s.splitterFn(input)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrSplitFunction, err)
	}

	var wg sync.WaitGroup
	var sm sync.Mutex
	pipeOutpus := []types.M{}

	for i, splittedInput := range splittedInputs {
		wg.Add(1)
		go func(i int, splittedInput types.M) {
			defer wg.Done()
			tube := NewTube(
				fmt.Sprintf("%s-%d", s.name, i),
				s.llm,
				s.decoder,
				s.memory,
			)

			output, err := tube.Run(ctx, splittedInput)
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
