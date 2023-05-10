package tool

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type duckDuckGo struct {
	name        string
	description string
}

const (
	baseUrl = "https://api.duckduckgo.com/?q=%s&format=json&pretty=1&no_redirect=1"
)

func NewDuckDuckGo() *duckDuckGo {
	return &duckDuckGo{
		name:        "duckduckgo",
		description: "A wrapper around DuckDuckGo Search. Useful for when you need to answer questions about current events. Input should be a search query.",
	}
}

func (m *duckDuckGo) Name() string {
	return m.name
}

func (m *duckDuckGo) Description() string {
	return m.description
}

// A Qresult holds the returned query data
type Qresult struct {
	Title string
	Info  string
	Ref   string
}

// Requests the query and puts the results into an array
func query(q string, it int) ([]Qresult, string) {

	qf := fmt.Sprintf("https://duckduckgo.com/html/?q=%s", url.QueryEscape(q))

	resp, _ := http.Get(qf)

	doc, err := goquery.NewDocumentFromReader(resp.Body)

	results := []Qresult{}

	if err != nil {
		log.Fatal(err)
	}

	sel := doc.Find(".web-result")

	for i := range sel.Nodes {
		// Break loop once required amount of results are add
		if it == len(results) {
			break
		}

		single := sel.Eq(i)
		titleNode := single.Find(".result__a")
		info := single.Find(".result__snippet").Text()
		title := titleNode.Text()
		ref, _ := url.QueryUnescape(strings.TrimPrefix(titleNode.Nodes[0].Attr[2].Val, "/l/?kh=-1&uddg="))

		results = append(results[:], Qresult{title, info, ref})

	}

	// Return array of results and formated query used to get the results
	return results, qf
}

func (m *duckDuckGo) Execute(ctx context.Context, input string) (string, error) {

	input = strings.ReplaceAll(input, "\"", "")
	res, _ := query(input, 1)

	if len(res) == 0 {
		return "", fmt.Errorf("no results found")
	}

	return res[0].Info, nil

}
