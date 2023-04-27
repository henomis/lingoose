package pipeline

import (
	"context"
	"fmt"
	"sync"
)

type Splitter struct {
	name       string
	llm        Llm
	decoder    Decoder
	memory     Memory
	splitterFn SplitterFn
}

type SplitterFn func(input interface{}) ([]interface{}, error)

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

func (s *Splitter) Run(ctx context.Context, input interface{}) (interface{}, error) {

	splittedInputs, err := s.splitterFn(input)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	var sm sync.Mutex
	stepsOutput := []map[string]interface{}{}

	for i, splittedInput := range splittedInputs {
		wg.Add(1)
		go func(i int, splittedInput interface{}) {
			defer wg.Done()
			step := NewStep(
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
			stepsOutput = append(stepsOutput, output.(map[string]interface{}))
			sm.Unlock()
		}(i, splittedInput)
	}

	wg.Wait()

	return stepsOutput, nil

}
