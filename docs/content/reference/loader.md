---
title: "Load Content into a Document"
description:
linkTitle: "Loader"
menu: { main: { parent: 'reference', weight: -95 } }
---

The `Loader` interface provides a way to load content into a document. It is used to transform a text content into a document, a structured representation of the content. LinGoose offers loaders for different types of content:

- Plain text
- CSV
- PDF (via pdftotext)
- Docx, odf, rtf, and other office formats (via LibreOffice)
- OCR (via Tesseract)
- Audio/STT (via OpenAI whisper, whispercpp or Hugging Face)
- Youtube (via youtube-dl)
- Pubmed
- Image to text (via Hugging Face)

## Using Loader

To use a loader, you need to create an instance of your preferred loader. Here we show how to create an instance of a loader using the plain text loader:

```go
pdfLoader := loader.NewPDFToText().WithPDFToTextPath("/opt/homebrew/bin/pdftotext")
kbDocuments := loader.LoadFromSource(context.Background(),"./kb/mydocument.pdf")
```

### Splitting documents

A loader produces a document for each content it loads. However documents may contain a huge amount of text, and it's convenient to split them into smaller parts.

```go
audioLoader := loader.NewWhisper().
		WithTextSplitter(textsplitter.NewRecursiveCharacterTextSplitter(2000, 200)).
		LoadFromSource(context.Background(), "audio.mp3")
```

A text splitter is a component that splits a document into documents of a smaller size. The `RecursiveCharacterTextSplitter` accepts as parameters the size of the text chunks and the size of chunk overlap.