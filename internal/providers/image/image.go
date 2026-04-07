package image

import "context"

type GenerateRequest struct {
	Prompt string
	Width  int
	Height int
}

type GenerateResponse struct {
	ImageBytes []byte
	MimeType   string
}

type Provider interface {
	Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error)
}
