package huggingface

import "fmt"

type options struct {
	UseGPU       *bool `json:"use_gpu,omitempty"`
	UseCache     *bool `json:"use_cache,omitempty"`
	WaitForModel *bool `json:"wait_for_model,omitempty"`
}

func debugCompletion(prompt string, content string) {
	fmt.Printf("---USER---\n%s\n", prompt)
	fmt.Printf("---AI---\n%s\n", content)
}
