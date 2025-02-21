package main

import (
	"context"
	"fmt"

	"github.com/rsest/lingoose/loader"
)

func main() {

	l := loader.NewWhisperCppLoader("/tmp/hello.mp3").WithWhisperCppPath("/tmp/whisper.cpp/main").WithModel("/tmp/whisper.cpp/models/ggml-base.bin")

	docs, err := l.Load(context.Background())
	if err != nil {
		panic(err)
	}

	fmt.Println(docs[0].Content)

}
