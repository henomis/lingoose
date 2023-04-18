// Package example provides a way to define examples for a prompt.
package example

type Examples struct {
	Examples  []Example
	Separator string
	Prefix    string
	Suffix    string
}

type Example map[string]interface{}
