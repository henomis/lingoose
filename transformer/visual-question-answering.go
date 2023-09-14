package transformer

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
)

const (
	hfAPIBaseURL                          = "https://api-inference.huggingface.co/models/"
	hfDefaultVisualQuestionAnsweringModel = "dandelin/vilt-b32-finetuned-vqa"
)

type VisualQuestionAnswering struct {
	mediaFile string
	token     string
	model     string
}

type VisualQuestionAnsweringRequest struct {
	Inputs VisualQuestionAnsweringRequestInputs `json:"inputs"`
}

type VisualQuestionAnsweringRequestInputs struct {
	Image    string `json:"image"`
	Question string `json:"question"`
}

type VisualQuestionAnsweringResponse struct {
	Score  float64 `json:"score"`
	Answer string  `json:"answer"`
}

func NewHFVisualQuestionAnswering(mediaFile string) *VisualQuestionAnswering {
	return &VisualQuestionAnswering{
		mediaFile: mediaFile,
		model:     hfDefaultVisualQuestionAnsweringModel,
		token:     os.Getenv("HUGGING_FACE_HUB_TOKEN"),
	}
}

func (v *VisualQuestionAnswering) WithModel(model string) *VisualQuestionAnswering {
	v.model = model
	return v
}

func (v *VisualQuestionAnswering) WithToken(token string) *VisualQuestionAnswering {
	v.token = token
	return v
}

func (v *VisualQuestionAnswering) WithImage(mediaFile string) *VisualQuestionAnswering {
	v.mediaFile = mediaFile
	return v
}

func (v *VisualQuestionAnswering) Transform(ctx context.Context, input string, all bool) (any, error) {
	respJSON, err := hfVisualQuestionAnsweringHTTPCall(ctx, v.token, v.model, v.mediaFile, input)
	if err != nil {
		return "", err
	}

	var resp []VisualQuestionAnsweringResponse
	err = json.Unmarshal(respJSON, &resp)
	if err != nil {
		return "", err
	}

	if all {
		return resp, nil
	}

	return resp[0].Answer, nil
}

func hfVisualQuestionAnsweringHTTPCall(ctx context.Context, token, model, mediaFile, question string) ([]byte, error) {
	var inputs VisualQuestionAnsweringRequest

	base64String, err := imageToBase64(mediaFile)
	if err != nil {
		return nil, err
	}

	inputs.Inputs = VisualQuestionAnsweringRequestInputs{
		Image:    base64String,
		Question: question,
	}

	buf, err := json.Marshal(inputs)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, hfAPIBaseURL+model, bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, errors.New("nil request created")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = hfCheckHTTPResponse(respBody)
	if err != nil {
		return nil, err
	}

	return respBody, nil
}

func hfCheckHTTPResponse(respJSON []byte) error {
	type apiError struct {
		Error string `json:"error,omitempty"`
	}

	type apiErrors struct {
		Errors []string `json:"error,omitempty"`
	}

	{
		buf := make([]byte, len(respJSON))
		copy(buf, respJSON)
		apiErr := apiError{}
		err := json.Unmarshal(buf, &apiErr)
		if err != nil {
			//nolint:nilerr
			return nil
		}
		if apiErr.Error != "" {
			return errors.New(string(respJSON))
		}
	}

	{
		buf := make([]byte, len(respJSON))
		copy(buf, respJSON)
		apiErrs := apiErrors{}
		err := json.Unmarshal(buf, &apiErrs)
		if err != nil {
			//nolint:nilerr
			return nil
		}
		if apiErrs.Errors != nil {
			return errors.New(string(respJSON))
		}
	}

	return nil
}

func imageToBase64(mediaFile string) (string, error) {
	img, err := os.Open(mediaFile)
	if err != nil {
		return "", err
	}
	defer img.Close()

	imgBytes, err := io.ReadAll(img)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(imgBytes), nil
}
