package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ismobaga/apgn/internal/providers/llm"
)

const defaultModel = "gpt-4o-mini"
const apiURL = "https://api.openai.com/v1/chat/completions"

type Provider struct {
	apiKey string
	model  string
	client *http.Client
}

func New(apiKey, model string) *Provider {
	if model == "" {
		model = defaultModel
	}
	return &Provider{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{},
	}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float32       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (p *Provider) Generate(ctx context.Context, req llm.GenerateRequest) (llm.GenerateResponse, error) {
	messages := []chatMessage{}
	if req.SystemPrompt != "" {
		messages = append(messages, chatMessage{Role: "system", Content: req.SystemPrompt})
	}
	messages = append(messages, chatMessage{Role: "user", Content: req.UserPrompt})

	body, err := json.Marshal(chatRequest{
		Model:       p.model,
		Messages:    messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
	})
	if err != nil {
		return llm.GenerateResponse{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return llm.GenerateResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return llm.GenerateResponse{}, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return llm.GenerateResponse{}, err
	}

	var chatResp chatResponse
	if err := json.Unmarshal(data, &chatResp); err != nil {
		return llm.GenerateResponse{}, err
	}

	if chatResp.Error != nil {
		return llm.GenerateResponse{}, fmt.Errorf("openai error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return llm.GenerateResponse{}, fmt.Errorf("no choices returned")
	}

	return llm.GenerateResponse{Text: chatResp.Choices[0].Message.Content}, nil
}
