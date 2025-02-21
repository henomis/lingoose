package main

import (
	"fmt"

	"github.com/rsest/lingoose/tool/serpapi"
)

func main() {

	t := serpapi.New()
	f := t.Fn().(serpapi.FnPrototype)

	fmt.Println(f(serpapi.Input{Query: "Simone Vellei"}))
}
