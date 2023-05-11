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
	llm Llm,
	splitterFn SplitterFn,
) *splitter {
	return &splitter{
		llm:        llm,
		splitterFn: splitterFn,
	}
}

func (s *splitter) WithDecoder(decoder Decoder) *splitter {
	s.decoder = decoder
	return s
}

func (s *splitter) WithMemory(name string, memory Memory) *splitter {
	s.name = name
	s.memory = memory
	return s
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

			tube := NewTube(s.llm)
			if s.memory != nil {
				tube = tube.WithMemory(fmt.Sprintf("%s-%d", s.name, i), s.memory)
			}
			if s.decoder != nil {
				tube = tube.WithDecoder(s.decoder)
			}

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
