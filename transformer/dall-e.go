package transformer

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"os"

	"github.com/sashabaranov/go-openai"
)

type DallEImageOutput any

type DallEImageSize string

const (
	DallEImageSize256x256   DallEImageSize = openai.CreateImageSize256x256
	DallEImageSize512x512   DallEImageSize = openai.CreateImageSize512x512
	DallEImageSize1024x1024 DallEImageSize = openai.CreateImageSize1024x1024
	DallEImageSize1792x104  DallEImageSize = openai.CreateImageSize1792x1024
	DallEImageSize1024x1792 DallEImageSize = openai.CreateImageSize1024x1792
)

type DallEImageFormat string

const (
	DallEImageFormatURL   DallEImageFormat = "url"
	DallEImageFormatFile  DallEImageFormat = "file"
	DallEImageFormatImage DallEImageFormat = "image"
)

type DallEModel string

const (
	DallEModel2 DallEModel = openai.CreateImageModelDallE2
	DallEModel3 DallEModel = openai.CreateImageModelDallE3
)

type DallEImageQuality string

const (
	DallEImageQualityHD       DallEImageQuality = openai.CreateImageQualityHD
	DallEImageQualityStandard DallEImageQuality = openai.CreateImageQualityStandard
)

type DallEImageStyle string

const (
	DallEImageStyleVivid   DallEImageStyle = openai.CreateImageStyleVivid
	DallEImageStyleNatural DallEImageStyle = openai.CreateImageStyleNatural
)

type DallE struct {
	openAIClient *openai.Client
	model        DallEModel
	imageSize    DallEImageSize
	imageFormat  DallEImageFormat
	imageStyle   DallEImageStyle
	imageQuality DallEImageQuality
}

func NewDallE() *DallE {
	openAIKey := os.Getenv("OPENAI_API_KEY")
	return &DallE{
		openAIClient: openai.NewClient(openAIKey),
		model:        DallEModel2,
		imageSize:    DallEImageSize256x256,
		imageFormat:  DallEImageFormatURL,
		imageStyle:   DallEImageStyleNatural,
		imageQuality: DallEImageQualityStandard,
	}
}

func (d *DallE) WithClient(client *openai.Client) *DallE {
	d.openAIClient = client
	return d
}

func (d *DallE) WithImageSize(imageSize DallEImageSize) *DallE {
	d.imageSize = imageSize
	return d
}

func (d *DallE) WithImageStyle(imageStyle DallEImageStyle) *DallE {
	d.imageStyle = imageStyle
	return d
}

func (d *DallE) WithImageQuality(imageQuality DallEImageQuality) *DallE {
	d.imageQuality = imageQuality
	return d
}

func (d *DallE) WithModel(model DallEModel) *DallE {
	d.model = model
	return d
}

func (d *DallE) WithImageFormat(imageFormat DallEImageFormat) *DallE {
	d.imageFormat = imageFormat
	return d
}

func (d *DallE) Transform(ctx context.Context, input string) (any, error) {
	switch d.imageFormat {
	case DallEImageFormatURL:
		return d.TransformAsURL(ctx, input)
	case DallEImageFormatFile:
		return d.TransformAsFile(ctx, input, nil)
	case DallEImageFormatImage:
		return d.TransformToImage(ctx, input)
	default:
		return "", fmt.Errorf("unknown image format: %s", d.imageFormat)
	}
}

func (d *DallE) TransformAsURL(ctx context.Context, input string) (string, error) {
	reqURL := openai.ImageRequest{
		Prompt:         input,
		Model:          string(d.model),
		Size:           string(d.imageSize),
		Quality:        string(d.imageQuality),
		Style:          string(d.imageStyle),
		ResponseFormat: openai.CreateImageResponseFormatURL,
		N:              1,
	}

	respURL, err := d.openAIClient.CreateImage(ctx, reqURL)
	if err != nil {
		return "", err
	}

	return respURL.Data[0].URL, nil
}

func (d *DallE) TransformAsFile(ctx context.Context, input string, file *os.File) (string, error) {
	imgData, err := d.TransformToImage(ctx, input)
	if err != nil {
		return "", err
	}

	if file == nil {
		// create a temporary file
		file, err = os.CreateTemp("", "dall-e-*.png")
		if err != nil {
			return "", err
		}
	}

	defer file.Close()

	err = png.Encode(file, imgData)
	if err != nil {
		return "", err
	}

	return file.Name(), nil
}

func (d *DallE) TransformToImage(ctx context.Context, input string) (image.Image, error) {
	reqBase64 := openai.ImageRequest{
		Prompt:         input,
		Size:           string(d.imageSize),
		ResponseFormat: openai.CreateImageResponseFormatB64JSON,
		N:              1,
	}

	respBase64, err := d.openAIClient.CreateImage(ctx, reqBase64)
	if err != nil {
		return nil, err
	}

	imgBytes, err := base64.StdEncoding.DecodeString(respBase64.Data[0].B64JSON)
	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(imgBytes)
	imgData, err := png.Decode(r)
	if err != nil {
		return nil, err
	}

	return imgData, nil
}
