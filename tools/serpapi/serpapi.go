package serpapi

import (
	"context"
	"fmt"
	"os"

	"github.com/henomis/restclientgo"
)

type Tool struct {
	restClient   *restclientgo.RestClient
	googleDomain string
	countryCode  string
	languageCode string
	apiKey       string
}

type Input struct {
	Query string `json:"query" jsonschema:"description=the query to search for"`
}

type Output struct {
	Error   string   `json:"error,omitempty"`
	Results []result `json:"results,omitempty"`
}

type FnPrototype = func(Input) Output

func New() *Tool {
	t := &Tool{
		apiKey:       os.Getenv("SERPAPI_API_KEY"),
		restClient:   restclientgo.New("https://serpapi.com"),
		googleDomain: "google.com",
		countryCode:  "us",
		languageCode: "en",
	}

	return t
}

func (t *Tool) WithGoogleDomain(googleDomain string) *Tool {
	t.googleDomain = googleDomain
	return t
}

func (t *Tool) WithCountryCode(countryCode string) *Tool {
	t.countryCode = countryCode
	return t
}

func (t *Tool) WithLanguageCode(languageCode string) *Tool {
	t.languageCode = languageCode
	return t
}

func (t *Tool) WithApiKey(apiKey string) *Tool {
	t.apiKey = apiKey
	return t
}

func (t *Tool) Name() string {
	return "google"
}

func (t *Tool) Description() string {
	return "A tool that uses the Google internet search engine for a query."
}

func (t *Tool) Fn() any {
	return t.fn
}

func (t *Tool) fn(i Input) Output {
	req := &request{
		Query:        i.Query,
		GoogleDomain: t.googleDomain,
		CountryCode:  t.countryCode,
		LanguageCode: t.languageCode,
		ApiKey:       t.apiKey,
	}
	res := &response{}

	err := t.restClient.Get(context.Background(), req, res)
	if err != nil {
		return Output{Error: fmt.Sprintf("failed to search serpapi: %v", err)}
	}

	return Output{Results: res.Results}
}
