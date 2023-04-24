package loader

import (
	"github.com/henomis/lingoose/document"

	"os"
	"path/filepath"
	"regexp"
)

type DirLoader struct {
	dirname  string
	metadata map[string]interface{}

	regExp *regexp.Regexp
}

func NewDirLoader(dirname string, regExMatch string) (*DirLoader, error) {

	regExp, err := regexp.Compile(regExMatch)
	if err != nil {
		return nil, err
	}

	return &DirLoader{
		dirname: dirname,
		regExp:  regExp,
	}, nil

}

func (t *DirLoader) Load() ([]document.Document, error) {
	docs := []document.Document{}

	err := filepath.Walk(t.dirname, func(path string, info os.FileInfo, err error) error {
		if err == nil && t.regExp.MatchString(info.Name()) {

			l, err := NewTextLoader(path, map[string]interface{}{"source": path})
			if err != nil {
				return err
			}

			d, err := l.Load()
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
