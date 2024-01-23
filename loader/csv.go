package loader

import (
	"context"
	"encoding/csv"
	"errors"
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

//nolint:revive
func (c *CSVLoader) WithTextSplitter(textSplitter TextSplitter) *CSVLoader {
	// can't split csv
	return c
}

func (c *CSVLoader) Load(ctx context.Context) ([]document.Document, error) {
	_ = ctx
	err := c.validate()
	if err != nil {
		return nil, err
	}

	documents, err := c.readCSV()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInternal, err)
	}

	return documents, nil
}

func (c *CSVLoader) LoadFromSource(ctx context.Context, source string) ([]document.Document, error) {
	c.filename = source
	return c.Load(ctx)
}

func (c *CSVLoader) validate() error {
	fileStat, err := os.Stat(c.filename)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInternal, err)
	}

	if fileStat.IsDir() {
		return fmt.Errorf("%w: %w", ErrInternal, os.ErrNotExist)
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
		record, errRead := reader.Read()
		if errors.Is(errRead, io.EOF) {
			break
		}
		if errRead != nil {
			return nil, errRead
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
