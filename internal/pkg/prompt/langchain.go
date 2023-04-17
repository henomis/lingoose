package prompt

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

type langchainPromptTemplate struct {
	InputVariables []string    `json:"input_variables" yaml:"input_variables"`
	OutputParser   interface{} `json:"output_parser" yaml:"output_parser"`
	Template       string      `json:"template" yaml:"template"`
	TemplateFormat string      `json:"template_format" yaml:"template_format"`
}

func (p *langchainPromptTemplate) toPromptTemplate() *PromptTemplate {
	//TODO: add outputs variables
	return New(p.InputVariables, []string{}, replaceVariables(p.Template))
}

func (p *langchainPromptTemplate) ImportFromLangchain(url string) error {

	if err := validateURL(url); err != nil {
		return err
	}

	//TODO: move to a shared package
	resp, err := http.Get(buildURL(url))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ErrUnableToLoadLangchainPrompt
	}

	if strings.HasSuffix(url, ".json") {
		return json.NewDecoder(resp.Body).Decode(p)
	} else if strings.HasSuffix(url, ".yaml") {
		return yaml.NewDecoder(resp.Body).Decode(p)
	}

	return ErrUnableToLoadLangchainPrompt
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

func replaceVariables(input string) string {
	re := regexp.MustCompile(`\{(\w+)\}`)
	return re.ReplaceAllStringFunc(input, func(match string) string {
		variable := strings.Trim(match, "{}")
		return "{{." + variable + "}}"
	})
}
