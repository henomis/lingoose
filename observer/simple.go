package observer

import (
	"encoding/json"
	"fmt"
)

func NewSimpleObserver() *Observer {
	return New(simple())
}

func simple() HandlerFn {
	return func(e *Event) {
		eventAsJSON, _ := json.MarshalIndent(e, " ", "  ")
		fmt.Println(string(eventAsJSON))
	}
}
