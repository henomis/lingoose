package loader

import (
	"github.com/henomis/lingoose/document"

	"os"
	"path/filepath"
	"regexp"
)

type directoryLoader struct {
	dirname        string
	regExPathMatch string
}

func NewDirectoryLoader(dirname string, regExPathMatch string) *directoryLoader {

	return &directoryLoader{
		dirname:        dirname,
		regExPathMatch: regExPathMatch,
	}

}

func (d *directoryLoader) Load() ([]document.Document, error) {

	regExp, err := regexp.Compile(d.regExPathMatch)
	if err != nil {
		return nil, err
	}

	docs := []document.Document{}

	err = filepath.Walk(d.dirname, func(path string, info os.FileInfo, err error) error {
		if err == nil && regExp.MatchString(info.Name()) {

			d, err := NewTextLoader(path, nil).Load()
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

	return docs, nil
}
