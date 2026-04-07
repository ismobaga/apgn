package elevenlabs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ismobaga/apgn/internal/providers/tts"
)

const apiURL = "https://api.elevenlabs.io/v1/text-to-speech"

type Provider struct {
	apiKey string
	client *http.Client
}

func New(apiKey string) *Provider {
	return &Provider{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

type ttsRequest struct {
	Text          string        `json:"text"`
	ModelID       string        `json:"model_id"`
	VoiceSettings voiceSettings `json:"voice_settings"`
}

type voiceSettings struct {
	Stability       float32 `json:"stability"`
	SimilarityBoost float32 `json:"similarity_boost"`
	SpeakingRate    float32 `json:"speaking_rate,omitempty"`
}

func (p *Provider) Synthesize(ctx context.Context, req tts.SynthesizeRequest) (tts.SynthesizeResponse, error) {
	format := req.Format
	if format == "" {
		format = "mp3_44100_128"
	}

	payload := ttsRequest{
		Text:    req.Text,
		ModelID: "eleven_multilingual_v2",
		VoiceSettings: voiceSettings{
			Stability:       0.5,
			SimilarityBoost: 0.75,
			SpeakingRate:    req.Rate,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return tts.SynthesizeResponse{}, err
	}

	url := fmt.Sprintf("%s/%s?output_format=%s", apiURL, req.VoiceID, format)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return tts.SynthesizeResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("xi-api-key", p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return tts.SynthesizeResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return tts.SynthesizeResponse{}, fmt.Errorf("elevenlabs error %d: %s", resp.StatusCode, string(data))
	}

	audio, err := io.ReadAll(resp.Body)
	if err != nil {
		return tts.SynthesizeResponse{}, err
	}

	return tts.SynthesizeResponse{
		AudioBytes: audio,
		DurationMS: 0, // unknown until playback
	}, nil
}
