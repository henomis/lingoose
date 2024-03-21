package observer

import (
	"context"
	"sync"
	"time"
)

// Event

type Event struct {
	Name      string
	CreatedAt time.Time
	Data      any
}

func NewEvent(name string, data any) *Event {
	return &Event{
		Name: name,
		Data: data,
	}
}

// Handler

type HandlerFn func(event *Event)

type handler struct {
	eventCh chan *Event
	doneCh  chan struct{}
	fn      HandlerFn
}

func newHandler(fn HandlerFn) *handler {
	return &handler{
		eventCh: make(chan *Event),
		doneCh:  make(chan struct{}),
		fn:      fn,
	}
}

func (h *handler) listen() {
	for {
		event, ok := <-h.eventCh
		if !ok {
			h.doneCh <- struct{}{}
			return
		}
		h.fn(event)
	}
}

// Dispatcher

type dispatcher struct {
	handlerChan chan *Event
	wg          sync.WaitGroup
}

func newDispatcher(ch chan *Event) *dispatcher {
	return &dispatcher{
		handlerChan: ch,
	}
}

func (d *dispatcher) dispatch(event *Event) {
	event.CreatedAt = time.Now().UTC()
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		d.handlerChan <- event
	}()
}

// Observer

type Observer struct {
	handler    *handler
	dispatcher *dispatcher
}

func New(fn HandlerFn) *Observer {
	handler := newHandler(fn)
	dispatcher := newDispatcher(handler.eventCh)
	o := &Observer{
		handler:    handler,
		dispatcher: dispatcher,
	}

	go o.handler.listen()

	return o
}

func (o *Observer) Dispatch(event *Event) {
	o.dispatcher.dispatch(event)
}

func (o *Observer) Wait(ctx context.Context) {
	done := make(chan struct{})
	go func() {
		// wait for all events to be sent
		o.dispatcher.wg.Wait()
		// close the handler event channel
		close(o.handler.eventCh)
		// wait for the handler to finish processing all events
		<-o.handler.doneCh
		// close the done channel
		close(done)
	}()

	select {
	case <-done:
		return
	case <-ctx.Done():
		return
	}
}
