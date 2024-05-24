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
	SearchMetadata                SearchMetadata                  `json:"search_metadata"`
	SearchParameters              SearchParameters                `json:"search_parameters"`
	SearchInformation             SearchInformation               `json:"search_information"`
	InlineImagesSuggestedSearches []InlineImagesSuggestedSearches `json:"inline_images_suggested_searches"`
	InlineImages                  []InlineImages                  `json:"inline_images"`
	AnswerBox                     AnswerBox                       `json:"answer_box"`
	OrganicResults                []OrganicResults                `json:"organic_results"`
	Pagination                    Pagination                      `json:"pagination"`
	SerpapiPagination             SerpapiPagination               `json:"serpapi_pagination"`
}
type SearchMetadata struct {
	ID             string  `json:"id"`
	Status         string  `json:"status"`
	JSONEndpoint   string  `json:"json_endpoint"`
	CreatedAt      string  `json:"created_at"`
	ProcessedAt    string  `json:"processed_at"`
	GoogleURL      string  `json:"google_url"`
	RawHTMLFile    string  `json:"raw_html_file"`
	TotalTimeTaken float64 `json:"total_time_taken"`
}
type SearchParameters struct {
	Engine       string `json:"engine"`
	Q            string `json:"q"`
	GoogleDomain string `json:"google_domain"`
	Hl           string `json:"hl"`
	Gl           string `json:"gl"`
	Device       string `json:"device"`
}
type SearchInformation struct {
	QueryDisplayed      string  `json:"query_displayed"`
	TotalResults        int     `json:"total_results"`
	TimeTakenDisplayed  float64 `json:"time_taken_displayed"`
	OrganicResultsState string  `json:"organic_results_state"`
}
type InlineImagesSuggestedSearches struct {
	Name        string `json:"name"`
	Link        string `json:"link"`
	Uds         string `json:"uds"`
	Q           string `json:"q"`
	SerpapiLink string `json:"serpapi_link"`
	Thumbnail   string `json:"thumbnail"`
}
type InlineImages struct {
	Link       string `json:"link"`
	Source     string `json:"source"`
	Thumbnail  string `json:"thumbnail"`
	Original   string `json:"original"`
	Title      string `json:"title"`
	SourceName string `json:"source_name"`
}
type AnswerBox struct {
	Type      string `json:"type"`
	Title     string `json:"title"`
	Thumbnail string `json:"thumbnail"`
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
type OtherPages struct {
	Num2 string `json:"2"`
	Num3 string `json:"3"`
	Num4 string `json:"4"`
	Num5 string `json:"5"`
}
type Pagination struct {
	Current    int        `json:"current"`
	Next       string     `json:"next"`
	OtherPages OtherPages `json:"other_pages"`
}
type SerpapiPagination struct {
	Current    int        `json:"current"`
	NextLink   string     `json:"next_link"`
	Next       string     `json:"next"`
	OtherPages OtherPages `json:"other_pages"`
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
