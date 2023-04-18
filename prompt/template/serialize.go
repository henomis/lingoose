package template

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Save saves the prompt template to a file.
func (p *Prompt) Save(path string) error {

	var data []byte
	var err error

	if strings.HasSuffix(path, ".yaml") {
		data, err = yaml.Marshal(p)
	} else if strings.HasSuffix(path, ".json") {
		data, err = json.MarshalIndent(p, "", " ")
	} else {
		return fmt.Errorf("invalid file extension (only .yaml and .json are supported)")
	}
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// Load loads the prompt template from a file.
func (p *Prompt) Load(path string) error {

	file, err := os.ReadFile(path)

	if err != nil {
		return err
	}

	if strings.HasSuffix(path, ".yaml") {
		if err := yaml.Unmarshal(file, p); err != nil {
			return err
		}
	} else if strings.HasSuffix(path, ".json") {
		if err := json.Unmarshal(file, p); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("invalid file extension (only .yaml and .json are supported)")
	}

	return nil
}
