package main

import (
	"context"
	"time"

	"github.com/henomis/lingoose/observer"
)

func main() {
	o := observer.New(
		func(event *observer.Event) {
			time.Sleep(2 * time.Second)
			println(event.Name)
		},
	)

	o.Dispatch(observer.NewEvent("test", nil))
	o.Dispatch(observer.NewEvent("test2", nil))

	o.Wait(context.Background())
}
