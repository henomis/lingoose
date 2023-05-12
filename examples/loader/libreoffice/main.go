package main

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/loader"
)

func main() {

	l := loader.NewLibreOfficeLoader("/tmp/doc.docx")

	docs, err := l.Load(context.Background())
	if err != nil {
		panic(err)
	}

	fmt.Println(docs[0].Content)

}
