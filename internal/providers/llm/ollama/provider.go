package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ismobaga/apgn/internal/providers/llm"
)

const (
	defaultModel   = "gemma3:latest"
	defaultBaseURL = "http://localhost:11434"
)

type Provider struct {
	baseURL string
	model   string
	client  *http.Client
}

func New(baseURL, model string) *Provider {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultBaseURL
	}
	if strings.TrimSpace(model) == "" {
		model = defaultModel
	}

	return &Provider{
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		client:  &http.Client{Timeout: 2 * time.Minute},
	}
}

type generateRequest struct {
	Model   string         `json:"model"`
	Prompt  string         `json:"prompt"`
	System  string         `json:"system,omitempty"`
	Stream  bool           `json:"stream"`
	Options map[string]any `json:"options,omitempty"`
}

type generateResponse struct {
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
}

func (p *Provider) Generate(ctx context.Context, req llm.GenerateRequest) (llm.GenerateResponse, error) {
	payload := generateRequest{
		Model:  p.model,
		Prompt: req.UserPrompt,
		System: req.SystemPrompt,
		Stream: false,
	}

	options := map[string]any{}
	if req.Temperature > 0 {
		options["temperature"] = req.Temperature
	}
	if req.MaxTokens > 0 {
		options["num_predict"] = req.MaxTokens
	}
	if len(options) > 0 {
		payload.Options = options
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return llm.GenerateResponse{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return llm.GenerateResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return llm.GenerateResponse{}, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return llm.GenerateResponse{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return llm.GenerateResponse{}, fmt.Errorf("ollama error %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}

	var genResp generateResponse
	if err := json.Unmarshal(data, &genResp); err != nil {
		return llm.GenerateResponse{}, err
	}
	if genResp.Error != "" {
		return llm.GenerateResponse{}, fmt.Errorf("ollama error: %s", genResp.Error)
	}
	if strings.TrimSpace(genResp.Response) == "" {
		return llm.GenerateResponse{}, fmt.Errorf("empty response from ollama")
	}

	return llm.GenerateResponse{Text: genResp.Response}, nil
}
