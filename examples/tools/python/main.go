package main

import (
	"fmt"

	"github.com/henomis/lingoose/tools/python"
)

func main() {
	t := python.New().WithPythonPath("python3")

	pythonScript := `print("Hello from Python!")`
	f := t.Fn().(python.FnPrototype)

	fmt.Println(f(python.Input{PythonCode: pythonScript}))
}
