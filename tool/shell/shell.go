package shell

import (
	"bytes"
	"fmt"
	"os/exec"
)

type Tool struct {
	shell         string
	askForConfirm bool
}

func New() *Tool {
	return &Tool{
		shell:         "bash",
		askForConfirm: true,
	}
}

func (t *Tool) WithShell(shell string) *Tool {
	t.shell = shell
	return t
}

func (t *Tool) WithAskForConfirm(askForConfirm bool) *Tool {
	t.askForConfirm = askForConfirm
	return t
}

type Input struct {
	BashScript string `json:"bash_code" jsonschema:"description=shell script"`
}

type Output struct {
	Error  string `json:"error,omitempty"`
	Result string `json:"result,omitempty"`
}

type FnPrototype = func(Input) Output

func (t *Tool) Name() string {
	return "bash"
}

func (t *Tool) Description() string {
	return "A tool that runs a shell script using the " + t.shell + " interpreter. Use it to interact with the OS."
}

func (t *Tool) Fn() any {
	return t.fn
}

//nolint:gosec
func (t *Tool) fn(i Input) Output {
	// Ask for confirmation if the flag is set.
	if t.askForConfirm {
		fmt.Println("Are you sure you want to run the following script?")
		fmt.Println("-------------------------------------------------")
		fmt.Println(i.BashScript)
		fmt.Println("-------------------------------------------------")
		fmt.Print("Type 'yes' to confirm > ")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "yes" {
			return Output{
				Error: "script execution aborted",
			}
		}
	}

	// Create a command to run the Bash interpreter with the script.
	cmd := exec.Command(t.shell, "-c", i.BashScript)

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
