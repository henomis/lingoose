package serpapi

import (
	"encoding/json"
	"io"

	"github.com/henomis/restclientgo"
)

type request struct {
	Query        string
	GoogleDomain string
	CountryCode  string
	LanguageCode string
	ApiKey       string
}

type response struct {
	HTTPStatusCode int
	Map            map[string]interface{}
	RawBody        []byte
	apiResponse    apiResponse
	Results        []result
}

type apiResponse struct {
	OrganicResults []OrganicResults `json:"organic_results"`
}

type Top struct {
	Extensions []string `json:"extensions"`
}

type RichSnippet struct {
	Top Top `json:"top"`
}

type OrganicResults struct {
	Position                int         `json:"position"`
	Title                   string      `json:"title"`
	Link                    string      `json:"link"`
	RedirectLink            string      `json:"redirect_link"`
	DisplayedLink           string      `json:"displayed_link"`
	Thumbnail               string      `json:"thumbnail,omitempty"`
	Favicon                 string      `json:"favicon"`
	Snippet                 string      `json:"snippet"`
	Source                  string      `json:"source"`
	RichSnippet             RichSnippet `json:"rich_snippet,omitempty"`
	SnippetHighlightedWords []string    `json:"snippet_highlighted_words,omitempty"`
}

type result struct {
	Title string
	Info  string
	URL   string
}

func (r *request) Path() (string, error) {
	urlValues := restclientgo.NewURLValues()
	urlValues.Add("q", &r.Query)
	urlValues.Add("api_key", &r.ApiKey)

	if r.GoogleDomain != "" {
		urlValues.Add("google_domain", &r.GoogleDomain)
	}

	if r.CountryCode != "" {
		urlValues.Add("gl", &r.CountryCode)
	}

	if r.LanguageCode != "" {
		urlValues.Add("hl", &r.LanguageCode)
	}

	params := urlValues.Encode()

	return "/search?" + params, nil
}

func (r *request) Encode() (io.Reader, error) {
	return nil, nil
}

func (r *request) ContentType() string {
	return ""
}

func (r *response) Decode(body io.Reader) error {
	err := json.NewDecoder(body).Decode(&r.apiResponse)
	if err != nil {
		return err
	}

	for _, res := range r.apiResponse.OrganicResults {
		r.Results = append(r.Results, result{
			Title: res.Title,
			Info:  res.Snippet,
			URL:   res.Link,
		})
	}

	return nil
}

func (r *response) SetBody(body io.Reader) error {
	r.RawBody, _ = io.ReadAll(body)
	return nil
}

func (r *response) AcceptContentType() string {
	return "application/json"
}

func (r *response) SetStatusCode(code int) error {
	r.HTTPStatusCode = code
	return nil
}

func (r *response) SetHeaders(_ restclientgo.Headers) error { return nil }
