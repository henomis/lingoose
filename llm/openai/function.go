package openai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/henomis/lingoose/chat"
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

func (o *openAI) BindFunction(
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

func (o *openAI) getFunctions() []openai.FunctionDefinition {

	functions := []openai.FunctionDefinition{}

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

func callFnWithArgumentAsJson(fn interface{}, argumentAsJson string) (string, error) {
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
	err := json.Unmarshal([]byte(argumentAsJson), &argValue)
	if err != nil {
		return "", fmt.Errorf("error unmarshaling argument: %s", err)
	}

	// Convert the argument value to the correct type
	argValueReflect := reflect.New(argType).Elem()
	jsonData, err := json.Marshal(argValue)
	if err != nil {
		return "", fmt.Errorf("error marshaling argument: %s", err)
	}
	err = json.Unmarshal(jsonData, argValueReflect.Addr().Interface())
	if err != nil {
		return "", fmt.Errorf("error unmarshaling argument: %s", err)
	}

	// Add the argument value to the slice
	args[0] = argValueReflect

	// Call the function with the argument
	fnValue := reflect.ValueOf(fn)
	result := fnValue.Call(args)

	// Marshal the function result to JSON
	if len(result) > 0 {
		jsonData, err := json.Marshal(result[0].Interface())
		if err != nil {
			return "", fmt.Errorf("error marshaling result: %s", err)
		}
		return string(jsonData), nil
	}

	return "", nil
}

func (o *openAI) iterateFunctionCall(
	ctx context.Context,
	prompt *chat.Chat,
	messages []openai.ChatCompletionMessage,
	response openai.ChatCompletionResponse,
) (openai.ChatCompletionResponse, error) {

	var err error
	responsePtr := &response
	iteration := uint(0)
	for {

		if iteration > o.functionsMaxIterations {
			return response, fmt.Errorf("%s: max iterations reached", ErrOpenAIChat)
		}

		if responsePtr.Choices[0].Message.FunctionCall == nil || responsePtr.Choices[0].FinishReason == "stop" {
			return *responsePtr, nil
		}

		responsePtr, err = o.handleFunctionCall(ctx, prompt, messages, responsePtr)
		if err != nil {
			return openai.ChatCompletionResponse{}, err
		}

		iteration++

	}
}

func (o *openAI) handleFunctionCall(
	ctx context.Context,
	prompt *chat.Chat,
	messages []openai.ChatCompletionMessage,
	response *openai.ChatCompletionResponse,
) (*openai.ChatCompletionResponse, error) {
	// append the request to call the function
	messages = append(messages, response.Choices[0].Message)

	fn, ok := o.functions[response.Choices[0].Message.FunctionCall.Name]
	if !ok {
		return nil, fmt.Errorf("%s: unknown function %s", ErrOpenAIChat, response.Choices[0].Message.FunctionCall.Name)
	}

	resultAsJSON, err := callFnWithArgumentAsJson(fn.Fn, response.Choices[0].Message.FunctionCall.Arguments)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrOpenAIChat, err)
	}

	if o.stopOnFirstFunctionCall {
		response.Choices[0].FinishReason = "stop"
		return response, nil
	}

	// append the result of the function call
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleFunction,
		Name:    response.Choices[0].Message.FunctionCall.Name,
		Content: resultAsJSON,
	})

	chatCompletionRequest := openai.ChatCompletionRequest{
		Model:       string(o.model),
		Messages:    messages,
		MaxTokens:   o.maxTokens,
		Temperature: o.temperature,
		N:           DefaultOpenAINumResults,
		TopP:        DefaultOpenAITopP,
		Stop:        o.stop,
	}

	if len(o.functions) > 0 {
		chatCompletionRequest.Functions = o.getFunctions()
	}

	newResponse, err := o.openAIClient.CreateChatCompletion(
		ctx,
		chatCompletionRequest,
	)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrOpenAIChat, err)
	}

	if o.usageCallback != nil {
		o.setUsageMetadata(response.Usage)
	}

	if len(newResponse.Choices) == 0 {
		return nil, fmt.Errorf("%s: no choices returned", ErrOpenAIChat)
	}

	return &newResponse, err
}
