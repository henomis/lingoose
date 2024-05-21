package main

import (
	"fmt"

	"github.com/henomis/lingoose/tools/duckduckgo"
)

func main() {

	t := duckduckgo.New().WithMaxResults(5)
	f := t.Fn().(duckduckgo.FnPrototype)

	fmt.Println(f(duckduckgo.Input{Query: "Simone Vellei"}))
}
