package loader

import (
	"context"
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

type PubMedLoader struct {
	loader Loader

	pubMedIDs []string
}

func NewPubmedLoader(pubMedIDs []string) *PubMedLoader {
	return &PubMedLoader{
		pubMedIDs: pubMedIDs,
	}
}

func (p *PubMedLoader) WithTextSplitter(textSplitter TextSplitter) *PubMedLoader {
	p.loader.textSplitter = textSplitter
	return p
}

func (p *PubMedLoader) Load(ctx context.Context) ([]document.Document, error) {
	documens := make([]document.Document, len(p.pubMedIDs))

	for i, pubMedID := range p.pubMedIDs {
		doc, err := p.load(ctx, pubMedID)
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

func (p *PubMedLoader) load(ctx context.Context, pubMedID string) (*document.Document, error) {
	url := fmt.Sprintf(pubMedBioCURLFormat, pubMedID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	jsonContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var pubMedDoc pubMedDocument
	err = json.Unmarshal(jsonContent, &pubMedDoc)
	if err != nil {
		return nil, err
	}

	content := ""
	for _, document := range pubMedDoc.Documents {
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
