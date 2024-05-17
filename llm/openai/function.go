package openai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

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

func bindFunction(
	fn interface{},
	name string,
	description string,
	functionParameterOptions ...FunctionParameterOption,
) (*Function, error) {
	parameter, err := extractFunctionParameter(fn)
	if err != nil {
		return nil, err
	}

	for _, option := range functionParameterOptions {
		err = option(parameter)
		if err != nil {
			return nil, err
		}
	}

	return &Function{
		Name:        name,
		Description: description,
		Parameters:  parameter,
		Fn:          fn,
	}, nil
}

func (o *Legacy) BindFunction(
	fn interface{},
	name string,
	description string,
	functionParameterOptions ...FunctionParameterOption,
) error {
	function, err := bindFunction(fn, name, description, functionParameterOptions...)
	if err != nil {
		return err
	}

	o.functions[name] = *function

	return nil
}

func (o *OpenAI) BindFunction(
	fn interface{},
	name string,
	description string,
	functionParameterOptions ...FunctionParameterOption,
) error {
	function, err := bindFunction(fn, name, description, functionParameterOptions...)
	if err != nil {
		return err
	}

	o.functions[name] = *function

	return nil
}

func (o *Legacy) getFunctions() []openai.FunctionDefinition {
	var functions []openai.FunctionDefinition

	for _, function := range o.functions {
		functions = append(functions, openai.FunctionDefinition{
			Name:        function.Name,
			Description: function.Description,
			Parameters:  function.Parameters,
		})
	}

	return functions
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
		var resultBytes bytes.Buffer
		enc := json.NewEncoder(&resultBytes)
		enc.SetEscapeHTML(false)
		err = enc.Encode(result[0].Interface())
		if err != nil {
			return "", fmt.Errorf("error marshaling result: %w", err)
		}
		return strings.TrimSpace(resultBytes.String()), nil
	}

	return "", nil
}

func (o *Legacy) functionCall(response openai.ChatCompletionResponse) (string, error) {
	fn, ok := o.functions[response.Choices[0].Message.FunctionCall.Name]
	if !ok {
		return "", fmt.Errorf("%w: unknown function %s", ErrOpenAIChat, response.Choices[0].Message.FunctionCall.Name)
	}

	resultAsJSON, err := callFnWithArgumentAsJSON(fn.Fn, response.Choices[0].Message.FunctionCall.Arguments)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrOpenAIChat, err)
	}
	o.calledFunctionName = &fn.Name
	return resultAsJSON, nil
}
