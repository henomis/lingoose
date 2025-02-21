package main

import (
	"context"
	"fmt"

	"github.com/rsest/lingoose/loader"
)

func main() {

	l := loader.NewHFImageToTextLoader("/tmp/image.jpg")

	docs, err := l.Load(context.Background())
	if err != nil {
		panic(err)
	}

	fmt.Println(docs[0].Content)

}
