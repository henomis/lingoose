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
	PythonCode string `json:"python_code" jsonschema:"description=python code that prints the final result to stdout."`
}

type Output struct {
	Error  string `json:"error,omitempty"`
	Result string `json:"result,omitempty"`
}

type FnPrototype = func(Input) Output

func (t *Tool) Name() string {
	return "python"
}

func (t *Tool) Description() string {
	return "A tool that runs Python code using the Python interpreter. The code should print the final result to stdout."
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

	// Return the output as a string.
	return Output{Result: out.String()}
}
