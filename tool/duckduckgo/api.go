package duckduckgo

import (
	"bytes"
	"io"
	"regexp"
	"strings"

	"github.com/henomis/restclientgo"
	"golang.org/x/net/html"
)

const (
	class = "class"
)

type request struct {
	Query string
}

type response struct {
	MaxResults     uint
	HTTPStatusCode int
	RawBody        []byte
	Results        []result
}

type result struct {
	Title string
	Info  string
	URL   string
}

func (r *request) Path() (string, error) {
	return "/html/?q=" + r.Query, nil
}

func (r *request) Encode() (io.Reader, error) {
	return nil, nil
}

func (r *request) ContentType() string {
	return ""
}

func (r *response) Decode(body io.Reader) error {
	results, err := r.parseBody(body)
	if err != nil {
		return err
	}

	r.Results = results
	return nil
}

func (r *response) SetBody(body io.Reader) error {
	r.RawBody, _ = io.ReadAll(body)
	return nil
}

func (r *response) AcceptContentType() string {
	return "text/html"
}

func (r *response) SetStatusCode(code int) error {
	r.HTTPStatusCode = code
	return nil
}

func (r *response) SetHeaders(_ restclientgo.Headers) error { return nil }

func (r *response) parseBody(body io.Reader) ([]result, error) {
	doc, err := html.Parse(body)
	if err != nil {
		return nil, err
	}
	ch := make(chan result)
	go r.findWebResults(ch, doc)

	results := []result{}
	for n := range ch {
		results = append(results, n)
	}

	return results, nil
}

func (r *response) findWebResults(ch chan result, doc *html.Node) {
	var results uint
	var f func(*html.Node)
	f = func(n *html.Node) {
		if results >= r.MaxResults {
			return
		}
		if n.Type == html.ElementNode && n.Data == "div" {
			for _, div := range n.Attr {
				if div.Key == class && strings.Contains(div.Val, "web-result") {
					info, href := r.findInfo(n)
					ch <- result{
						Title: r.findTitle(n),
						Info:  info,
						URL:   href,
					}
					results++
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	close(ch)
}

func (r *response) findTitle(n *html.Node) string {
	var title string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == class && strings.Contains(a.Val, "result__a") {
					title = n.FirstChild.Data
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return title
}

//nolint:gocognit
func (r *response) findInfo(n *html.Node) (string, string) {
	var info string
	var link string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == class && strings.Contains(a.Val, "result__snippet") {
					var b bytes.Buffer
					_ = html.Render(&b, n)

					re := regexp.MustCompile("<.*?>")
					info = html.UnescapeString(re.ReplaceAllString(b.String(), ""))

					for _, h := range n.Attr {
						if h.Key == "href" {
							link = "https:" + h.Val
							break
						}
					}
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return info, link
}
