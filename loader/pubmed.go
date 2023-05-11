package loader

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/henomis/lingoose/document"
	"github.com/henomis/lingoose/types"
)

const pubMedBioCURLFormat = "https://ncbi.nlm.nih.gov/research/bionlp/RESTful/pmcoa.cgi/BioC_json/%s/unicode"

type pubMedDocument struct {
	Documents []struct {
		Passages []struct {
			Text string `json:"text"`
		} `json:"passages"`
	} `json:"documents"`
}

type pubMedLoader struct {
	loader loader

	pubMedIDs []string
}

func NewPubmedLoader(pubMedIDs []string) *pubMedLoader {
	return &pubMedLoader{
		pubMedIDs: pubMedIDs,
	}
}

func (p *pubMedLoader) WithTextSplitter(textSplitter TextSplitter) *pubMedLoader {
	p.loader.textSplitter = textSplitter
	return p
}

func (p *pubMedLoader) Load() ([]document.Document, error) {

	documens := make([]document.Document, len(p.pubMedIDs))

	for i, pubMedID := range p.pubMedIDs {

		doc, err := p.load(pubMedID)
		if err != nil {
			return nil, err
		}

		documens[i] = *doc
	}

	if p.loader.textSplitter != nil {
		documens = p.loader.textSplitter.SplitDocuments(documens)
	}

	return documens, nil
}

func (p *pubMedLoader) load(pubMedID string) (*document.Document, error) {

	url := fmt.Sprintf(pubMedBioCURLFormat, pubMedID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	jsonContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var pubMedDocument pubMedDocument
	err = json.Unmarshal(jsonContent, &pubMedDocument)
	if err != nil {
		return nil, err
	}

	content := ""
	for _, document := range pubMedDocument.Documents {
		for _, passage := range document.Passages {
			content += passage.Text
		}
	}

	return &document.Document{
		Content: content,
		Metadata: types.Meta{
			"source": url,
		},
	}, nil
}
