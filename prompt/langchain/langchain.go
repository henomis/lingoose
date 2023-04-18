package langchain

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	langchainURLSchema = "lc://"
	baseURL            = "https://raw.githubusercontent.com/hwchase17/langchain-hub/master/"
)

var (
	ErrInvalidLangchainPromptURL   = fmt.Errorf("invalid langchain prompt url")
	ErrUnableToLoadLangchainPrompt = fmt.Errorf("unable to load langchain prompt")
)

type Template struct {
	InputVariables []string    `json:"input_variables" yaml:"input_variables"`
	OutputParser   interface{} `json:"output_parser" yaml:"output_parser"`
	Template       string      `json:"template" yaml:"template"`
	Format         string      `json:"template_format" yaml:"template_format"`
}

func New(url string) (*Template, error) {

	if err := validateURL(url); err != nil {
		return nil, err
	}

	//TODO: move to a shared package
	resp, err := http.Get(buildURL(url))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrUnableToLoadLangchainPrompt
	}

	var langchainPromptTemplate Template

	if strings.HasSuffix(url, ".json") {
		json.NewDecoder(resp.Body).Decode(&langchainPromptTemplate)
	} else if strings.HasSuffix(url, ".yaml") {
		yaml.NewDecoder(resp.Body).Decode(&langchainPromptTemplate)
	} else {
		return nil, ErrUnableToLoadLangchainPrompt
	}

	return &langchainPromptTemplate, nil
}

func (p *Template) ConvertedTemplate() string {
	re := regexp.MustCompile(`\{(\w+)\}`)
	return re.ReplaceAllStringFunc(p.Template, func(match string) string {
		variable := strings.Trim(match, "{}")
		return "{{." + variable + "}}"
	})
}

func validateURL(url string) error {
	if !strings.HasPrefix(url, langchainURLSchema) {
		return ErrInvalidLangchainPromptURL
	}

	return nil
}

func buildURL(url string) string {
	return fmt.Sprintf("%s%s", baseURL, strings.Replace(url, langchainURLSchema, "", 1))
}
