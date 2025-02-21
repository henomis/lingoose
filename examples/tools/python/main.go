package main

import (
	"fmt"

	"github.com/rsest/lingoose/tool/python"
)

func main() {
	t := python.New().WithPythonPath("python3")

	pythonScript := `print("Hello from Python!")`
	f := t.Fn().(python.FnPrototype)

	fmt.Println(f(python.Input{PythonCode: pythonScript}))
}
