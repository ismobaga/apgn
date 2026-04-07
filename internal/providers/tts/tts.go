package tts

import "context"

type SynthesizeRequest struct {
	VoiceID  string
	Text     string
	Format   string
	Rate     float32
	Metadata map[string]any
}

type SynthesizeResponse struct {
	AudioBytes []byte
	DurationMS int
}

type Provider interface {
	Synthesize(ctx context.Context, req SynthesizeRequest) (SynthesizeResponse, error)
}
