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

type Splitter struct {
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
) *Splitter {
	return &Splitter{
		llm:        llm,
		splitterFn: splitterFn,
	}
}

func (s *Splitter) WithDecoder(decoder Decoder) *Splitter {
	s.decoder = decoder
	return s
}

func (s *Splitter) WithMemory(name string, memory Memory) *Splitter {
	s.name = name
	s.memory = memory
	return s
}

func (s *Splitter) Run(ctx context.Context, input types.M) (types.M, error) {

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
