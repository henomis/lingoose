package loader

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/henomis/lingoose/document"
	"github.com/henomis/lingoose/types"
)

type CSVLoader struct {
	separator  rune
	filename   string
	lazyQuotes bool
}

func NewCSVLoader(filename string) *CSVLoader {
	return &CSVLoader{
		filename:  filename,
		separator: ',',
	}
}

func (c *CSVLoader) WithLazyQuotes() *CSVLoader {
	c.lazyQuotes = true
	return c
}

func (c *CSVLoader) WithSeparator(separator rune) *CSVLoader {
	c.separator = separator
	return c
}

func (t *CSVLoader) WithTextSplitter(textSplitter TextSplitter) *CSVLoader {
	// can't split csv
	return t
}

func (c *CSVLoader) Load(ctx context.Context) ([]document.Document, error) {

	err := c.validate()
	if err != nil {
		return nil, err
	}

	documents, err := c.readCSV()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrorInternal, err)
	}

	return documents, nil
}

func (c *CSVLoader) validate() error {

	fileStat, err := os.Stat(c.filename)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrorInternal, err)
	}

	if fileStat.IsDir() {
		return fmt.Errorf("%s: %w", ErrorInternal, os.ErrNotExist)
	}

	return nil
}

func (c *CSVLoader) readCSV() ([]document.Document, error) {
	csvFile, err := os.Open(c.filename)
	if err != nil {
		log.Fatal(err)
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)
	reader.Comma = c.separator
	reader.LazyQuotes = c.lazyQuotes

	var documents []document.Document
	var titles []string

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if titles == nil {

			titles = make([]string, len(record))
			for i, r := range record {
				titles[i] = strings.ReplaceAll(r, "\"", "")
				titles[i] = strings.TrimSpace(titles[i])
			}

			continue
		}

		var content string
		for i, title := range titles {
			value := strings.ReplaceAll(record[i], "\"", "")
			value = strings.TrimSpace(value)
			content += fmt.Sprintf("%s: %s", title, value)
			content += "\n"
		}

		documents = append(documents, document.Document{
			Content: content,
			Metadata: types.Meta{
				SourceMetadataKey: c.filename,
			},
		})
	}

	return documents, nil
}
