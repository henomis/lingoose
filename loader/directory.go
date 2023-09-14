package loader

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/document"

	"os"
	"path/filepath"
	"regexp"
)

type DirectoryLoader struct {
	loader Loader

	dirname        string
	regExPathMatch string
}

func NewDirectoryLoader(dirname string, regExPathMatch string) *DirectoryLoader {
	return &DirectoryLoader{
		dirname:        dirname,
		regExPathMatch: regExPathMatch,
	}
}

func (d *DirectoryLoader) WithTextSplitter(textSplitter TextSplitter) *DirectoryLoader {
	d.loader.textSplitter = textSplitter
	return d
}

func (d *DirectoryLoader) Load(ctx context.Context) ([]document.Document, error) {
	err := d.validate()
	if err != nil {
		return nil, err
	}

	regExp, err := regexp.Compile(d.regExPathMatch)
	if err != nil {
		return nil, err
	}

	docs := []document.Document{}

	err = filepath.Walk(d.dirname, func(path string, info os.FileInfo, err error) error {
		if err == nil && regExp.MatchString(info.Name()) {
			d, err := NewTextLoader(path, nil).Load(ctx)
			if err != nil {
				return err
			}

			docs = append(docs, d...)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if d.loader.textSplitter != nil {
		docs = d.loader.textSplitter.SplitDocuments(docs)
	}

	return docs, nil
}

func (d *DirectoryLoader) validate() error {
	fileStat, err := os.Stat(d.dirname)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrorInternal, err)
	}

	if !fileStat.IsDir() {
		return fmt.Errorf("%s: %w", ErrorInternal, os.ErrNotExist)
	}

	return nil
}
