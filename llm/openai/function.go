package openai

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/invopop/jsonschema"
	"github.com/sashabaranov/go-openai"
)

type Function struct {
	Name        string
	Description string
	Parameters  map[string]interface{}
	Fn          interface{}
}

type FunctionParameterOption func(map[string]interface{}) error

func (o *OpenAI) BindFunction(
	fn interface{},
	name string,
	description string,
	functionParamenterOptions ...FunctionParameterOption,
) error {
	parameter, err := extractFunctionParameter(fn)
	if err != nil {
		return err
	}

	for _, option := range functionParamenterOptions {
		err = option(parameter)
		if err != nil {
			return err
		}
	}

	function := Function{
		Name:        name,
		Description: description,
		Parameters:  parameter,
		Fn:          fn,
	}

	o.functions[name] = function

	return nil
}

func (o *OpenAI) getFunctions() []openai.Tool {
	tools := []openai.Tool{}

	for _, function := range o.functions {
		tools = append(tools, openai.Tool{
			Type: "function",
			Function: openai.FunctionDefinition{
				Name:        function.Name,
				Description: function.Description,
				Parameters:  function.Parameters,
			},
		})
	}

	return tools
}

func extractFunctionParameter(f interface{}) (map[string]interface{}, error) {
	// Get the type of the input function
	fnType := reflect.TypeOf(f)

	if fnType.Kind() != reflect.Func {
		return nil, errors.New("input must be a function")
	}

	// Check that the function only has one argument
	if fnType.NumIn() != 1 {
		return nil, errors.New("function must have exactly one argument")
	}

	// Check that the argument is of type struct
	argType := fnType.In(0)
	if argType.Kind() != reflect.Struct {
		return nil, errors.New("argument must be of type struct")
	}

	// Create a new instance of the argument type
	argValue := reflect.New(argType).Elem().Interface()

	parameter, err := structAsJSONSchema(argValue)
	if err != nil {
		return nil, err
	}

	return parameter, nil
}

func structAsJSONSchema(v interface{}) (map[string]interface{}, error) {
	r := new(jsonschema.Reflector)
	r.DoNotReference = true
	schema := r.Reflect(v)

	b, err := json.Marshal(schema)
	if err != nil {
		return nil, err
	}

	var jsonSchema map[string]interface{}
	err = json.Unmarshal(b, &jsonSchema)
	if err != nil {
		return nil, err
	}

	delete(jsonSchema, "$schema")

	return jsonSchema, nil
}

func callFnWithArgumentAsJSON(fn interface{}, argumentAsJSON string) (string, error) {
	// Get the type of the input function
	fnType := reflect.TypeOf(fn)

	// Check that the function has one argument
	if fnType.NumIn() != 1 {
		return "", fmt.Errorf("function must have one argument")
	}

	// Check that the argument is a struct
	argType := fnType.In(0)
	if argType.Kind() != reflect.Struct {
		return "", fmt.Errorf("argument must be a struct")
	}

	// Create a slice to hold the function argument
	args := make([]reflect.Value, 1)

	// Unmarshal the JSON string into an interface{} value
	var argValue interface{}
	err := json.Unmarshal([]byte(argumentAsJSON), &argValue)
	if err != nil {
		return "", fmt.Errorf("error unmarshaling argument: %w", err)
	}

	// Convert the argument value to the correct type
	argValueReflect := reflect.New(argType).Elem()
	jsonData, err := json.Marshal(argValue)
	if err != nil {
		return "", fmt.Errorf("error marshaling argument: %w", err)
	}
	err = json.Unmarshal(jsonData, argValueReflect.Addr().Interface())
	if err != nil {
		return "", fmt.Errorf("error unmarshaling argument: %w", err)
	}

	// Add the argument value to the slice
	args[0] = argValueReflect

	// Call the function with the argument
	fnValue := reflect.ValueOf(fn)
	result := fnValue.Call(args)

	// Marshal the function result to JSON
	if len(result) > 0 {
		jsonResultData, errMarshal := json.Marshal(result[0].Interface())
		if errMarshal != nil {
			return "", fmt.Errorf("error marshaling result: %w", errMarshal)
		}
		return string(jsonResultData), nil
	}

	return "", nil
}

func (o *OpenAI) functionCall(response openai.ChatCompletionResponse) (string, error) {
	fn, ok := o.functions[response.Choices[0].Message.ToolCalls[0].Function.Name]
	if !ok {
		return "", fmt.Errorf("%w: unknown function %s", ErrOpenAIChat, response.Choices[0].Message.ToolCalls[0].Function.Name)
	}

	resultAsJSON, err := callFnWithArgumentAsJSON(fn.Fn, response.Choices[0].Message.ToolCalls[0].Function.Arguments)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrOpenAIChat, err)
	}

	o.calledFunctionName = &fn.Name

	return resultAsJSON, nil
}
