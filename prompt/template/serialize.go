package template

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type promtSerialize struct {
	Inputs   []string `json:"inputs" yaml:"inputs"`
	Outputs  []string `json:"outputs" yaml:"outputs"`
	Template string   `json:"template" yaml:"template"`
	Partials Inputs   `json:"partials" yaml:"partials"`
}

func (p *Template) serialize() *promtSerialize {
	return &promtSerialize{
		Inputs:   p.inputs,
		Outputs:  p.outputs,
		Template: p.template,
		Partials: p.partials,
	}
}

func (p *Template) deserialize(promptSerialize *promtSerialize) {
	p.inputs = promptSerialize.Inputs
	p.outputs = promptSerialize.Outputs
	p.template = promptSerialize.Template
	p.partials = promptSerialize.Partials
}

// Save saves the prompt template to a file.
func (p *Template) Save(path string) error {

	var data []byte
	var err error

	if strings.HasSuffix(path, ".yaml") {
		data, err = yaml.Marshal(p.serialize())
	} else if strings.HasSuffix(path, ".json") {
		data, err = json.MarshalIndent(p.serialize(), "", " ")
	} else {
		return fmt.Errorf("invalid file extension (only .yaml and .json are supported)")
	}
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// Load loads the prompt template from a file.
func (p *Template) Load(path string) error {

	file, err := os.ReadFile(path)

	if err != nil {
		return err
	}

	var promptSerialize promtSerialize

	if strings.HasSuffix(path, ".yaml") {
		if err := yaml.Unmarshal(file, &promptSerialize); err != nil {
			return err
		}
	} else if strings.HasSuffix(path, ".json") {
		if err := json.Unmarshal(file, &promptSerialize); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("invalid file extension (only .yaml and .json are supported)")
	}

	p.deserialize(&promptSerialize)

	return nil
}
