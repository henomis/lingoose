package python

import (
	"bytes"
	"fmt"
	"os/exec"
)

type Tool struct {
	pythonPath string
}

func New() *Tool {
	return &Tool{
		pythonPath: "python3",
	}
}

func (t *Tool) WithPythonPath(pythonPath string) *Tool {
	t.pythonPath = pythonPath
	return t
}

type Input struct {
	// nolint:lll
	PythonCode string `json:"python_code" jsonschema:"description=python code that uses print() to print the final result to stdout."`
}

type Output struct {
	Error  string `json:"error,omitempty"`
	Result string `json:"result,omitempty"`
}

type FnPrototype = func(Input) Output

func (t *Tool) Name() string {
	return "python"
}

//nolint:lll
func (t *Tool) Description() string {
	// nolint:lll
	return "Use this tool to solve calculations, manipulate data, or perform any other Python-related tasks. The code should use print() to print the final result to stdout."
}

func (t *Tool) Fn() any {
	return t.fn
}

//nolint:gosec
func (t *Tool) fn(i Input) Output {
	// Create a command to run the Python interpreter with the script.
	cmd := exec.Command(t.pythonPath, "-c", i.PythonCode)

	// Create a buffer to capture the output.
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	// Run the command.
	err := cmd.Run()
	if err != nil {
		return Output{
			Error: fmt.Sprintf("failed to run script: %v, stderr: %v", err, stderr.String()),
		}
	}

	if out.String() == "" {
		return Output{
			Error: "no output from script, script must print the final result to stdout",
		}
	}

	// Return the output as a string.
	return Output{Result: out.String()}
}
