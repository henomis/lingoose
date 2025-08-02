---
linkTitle: LLM Providers
title: Language Model Providers
description: Guide to using different Large Language Model providers with LinGoose
menu: main
weight: -2
---

LinGoose supports multiple Large Language Model (LLM) providers, allowing you to choose the best model for your specific use case. This guide covers all supported providers with practical examples.

## OpenAI

OpenAI provides state-of-the-art language models including GPT-4, GPT-3.5, and specialized models.

### Basic Usage

```go
package main

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/thread"
)

func main() {
	// Create OpenAI LLM instance
	llm := openai.New()

	// Create a conversation thread
	myThread := thread.New().AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent("Explain what a neural network is"),
		),
	)

	// Generate response
	err := llm.Generate(context.Background(), myThread)
	if err != nil {
		panic(err)
	}

	fmt.Println(myThread.LastMessage().Contents[0].AsText())
}
```

### Configuration Options

```go
llm := openai.New().
	WithModel("gpt-4").
	WithTemperature(0.7).
	WithMaxTokens(2048).
	WithTimeout(30 * time.Second)
```

### Function Calling

```go
// Define a custom function
function := openai.Function{
	Name:        "get_weather",
	Description: "Get current weather for a location",
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"location": map[string]interface{}{
				"type":        "string",
				"description": "City name",
			},
		},
		"required": []string{"location"},
	},
}

llm := openai.New().WithFunctions([]openai.Function{function})
```

### Streaming Responses

```go
// Stream responses for real-time output
err := llm.GenerateStream(context.Background(), myThread, func(response string) {
	fmt.Print(response)
})
```

## Anthropic (Claude)

Anthropic's Claude models excel at reasoning, analysis, and safe AI interactions.

### Basic Usage

```go
package main

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/llm/anthropic"
	"github.com/henomis/lingoose/thread"
)

func main() {
	llm := anthropic.New().
		WithModel("claude-3-sonnet-20240229").
		WithTemperature(0.5)

	myThread := thread.New().AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent("Analyze the pros and cons of renewable energy"),
		),
	)

	err := llm.Generate(context.Background(), myThread)
	if err != nil {
		panic(err)
	}

	fmt.Println(myThread.LastMessage().Contents[0].AsText())
}
```

### Multimodal Support

```go
// Add image content to your message
myThread := thread.New().AddMessage(
	thread.NewUserMessage().AddContent(
		thread.NewTextContent("What do you see in this image?"),
		thread.NewImageContentFromURL("https://example.com/image.jpg"),
	),
)
```

## Cohere

Cohere provides efficient and powerful language models optimized for various use cases.

### Basic Usage

```go
package main

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/llm/cohere"
	"github.com/henomis/lingoose/thread"
)

func main() {
	llm := cohere.New().
		WithModel("command").
		WithTemperature(0.8)

	myThread := thread.New().AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent("Write a creative story about AI"),
		),
	)

	err := llm.Generate(context.Background(), myThread)
	if err != nil {
		panic(err)
	}

	fmt.Println(myThread.LastMessage().Contents[0].AsText())
}
```

## Groq

Groq provides ultra-fast inference for various language models.

### Basic Usage

```go
package main

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/llm/groq"
	"github.com/henomis/lingoose/thread"
)

func main() {
	llm := groq.New().
		WithModel("mixtral-8x7b-32768").
		WithTemperature(0.6)

	myThread := thread.New().AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent("Explain quantum computing simply"),
		),
	)

	err := llm.Generate(context.Background(), myThread)
	if err != nil {
		panic(err)
	}

	fmt.Println(myThread.LastMessage().Contents[0].AsText())
}
```

## Ollama (Local Models)

Ollama allows you to run open-source models locally for privacy and cost efficiency.

### Basic Usage

```go
package main

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/llm/ollama"
	"github.com/henomis/lingoose/thread"
)

func main() {
	// Configure Ollama endpoint (default: http://localhost:11434)
	llm := ollama.New().
		WithEndpoint("http://localhost:11434").
		WithModel("llama2").
		WithTemperature(0.7)

	myThread := thread.New().AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent("What are the benefits of local AI models?"),
		),
	)

	err := llm.Generate(context.Background(), myThread)
	if err != nil {
		panic(err)
	}

	fmt.Println(myThread.LastMessage().Contents[0].AsText())
}
```

### Multimodal with Ollama

```go
// Use vision-capable models like llava
llm := ollama.New().WithModel("llava")

myThread := thread.New().AddMessage(
	thread.NewUserMessage().AddContent(
		thread.NewTextContent("Describe this image"),
		thread.NewImageContentFromFile("./image.jpg"),
	),
)
```

## HuggingFace

Access thousands of open-source models through HuggingFace's Inference API.

### Text Generation

```go
package main

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/llm/huggingface"
)

func main() {
	llm := huggingface.NewTextGeneration().
		WithModel("microsoft/DialoGPT-large").
		WithToken("your-hf-token")

	response, err := llm.Generate(context.Background(), "Hello, how are you?")
	if err != nil {
		panic(err)
	}

	fmt.Println(response)
}
```

### Conversational AI

```go
llm := huggingface.NewConversational().
	WithModel("microsoft/DialoGPT-medium").
	WithToken("your-hf-token")

conversation := &huggingface.Conversation{
	PastUserInputs:    []string{"Hello"},
	GeneratedResponses: []string{"Hi there!"},
	Text:              "How's the weather?",
}

response, err := llm.Chat(context.Background(), conversation)
if err != nil {
	panic(err)
}

fmt.Println(response.GeneratedText)
```

## LocalAI

LocalAI provides OpenAI-compatible API for local model deployment.

### Basic Usage

```go
package main

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/llm/localai"
	"github.com/henomis/lingoose/thread"
)

func main() {
	llm := localai.New().
		WithEndpoint("http://localhost:8080").
		WithModel("gpt-3.5-turbo").  // Your local model name
		WithTemperature(0.7)

	myThread := thread.New().AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent("Explain the benefits of local AI deployment"),
		),
	)

	err := llm.Generate(context.Background(), myThread)
	if err != nil {
		panic(err)
	}

	fmt.Println(myThread.LastMessage().Contents[0].AsText())
}
```

## LlamaCPP

Direct integration with llama.cpp for efficient local inference.

### Basic Usage

```go
package main

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/llm/llamacpp"
	"github.com/henomis/lingoose/thread"
)

func main() {
	llm := llamacpp.New().
		WithEndpoint("http://localhost:8080").
		WithTemperature(0.8)

	myThread := thread.New().AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent("What are the advantages of running models locally?"),
		),
	)

	err := llm.Generate(context.Background(), myThread)
	if err != nil {
		panic(err)
	}

	fmt.Println(myThread.LastMessage().Contents[0].AsText())
}
```

## Caching

Improve performance and reduce costs by caching LLM responses.

### Redis Cache

```go
package main

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/llm/cache"
	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/thread"
)

func main() {
	// Create base LLM
	baseLLM := openai.New()

	// Add caching layer
	cachedLLM := cache.New(baseLLM).
		WithRedis("localhost:6379", "", 0).
		WithTTL(3600) // Cache for 1 hour

	myThread := thread.New().AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent("What is machine learning?"),
		),
	)

	// First call will hit the LLM, subsequent calls will use cache
	err := cachedLLM.Generate(context.Background(), myThread)
	if err != nil {
		panic(err)
	}

	fmt.Println(myThread.LastMessage().Contents[0].AsText())
}
```

## Error Handling and Best Practices

### Robust Error Handling

```go
func generateWithRetry(ctx context.Context, llm LLM, thread *thread.Thread) error {
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		err := llm.Generate(ctx, thread)
		if err == nil {
			return nil
		}
		
		// Check for specific error types
		if isRateLimitError(err) {
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}
		
		return fmt.Errorf("generation failed after %d retries: %w", i+1, err)
	}
	return fmt.Errorf("max retries exceeded")
}
```

### Environment Configuration

```go
// Set up environment variables for API keys
// OPENAI_API_KEY=your-openai-key
// ANTHROPIC_API_KEY=your-anthropic-key
// COHERE_API_KEY=your-cohere-key
// HF_TOKEN=your-huggingface-token

// LinGoose automatically picks up these environment variables
llm := openai.New() // Will use OPENAI_API_KEY automatically
```

### Model Selection Guidelines

- **GPT-4**: Best for complex reasoning, analysis, and creative tasks
- **GPT-3.5 Turbo**: Fast and cost-effective for most applications
- **Claude 3**: Excellent for safe, helpful, and harmless interactions
- **Llama 2/3**: Great open-source alternative for local deployment
- **Mixtral**: Efficient mixture-of-experts model via Groq
- **Cohere Command**: Optimized for business and enterprise use cases

Choose your provider based on your requirements for cost, latency, privacy, and model capabilities.