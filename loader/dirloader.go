package loader

import (
	"github.com/henomis/lingoose/document"

	"os"
	"path/filepath"
	"regexp"
)

type DirectoryLoader struct {
	dirname string
	regExp  *regexp.Regexp
}

func NewDirectoryLoader(dirname string, regExMatch string) (*DirectoryLoader, error) {

	regExp, err := regexp.Compile(regExMatch)
	if err != nil {
		return nil, err
	}

	return &DirectoryLoader{
		dirname: dirname,
		regExp:  regExp,
	}, nil

}

func (t *DirectoryLoader) Load() ([]document.Document, error) {
	docs := []document.Document{}

	err := filepath.Walk(t.dirname, func(path string, info os.FileInfo, err error) error {
		if err == nil && t.regExp.MatchString(info.Name()) {

			l, err := NewTextLoader(path, nil)
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
