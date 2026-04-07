package pipeline

import (
	"context"
	"fmt"

	"github.com/ismobaga/apgn/internal/domain/episode"
	"github.com/ismobaga/apgn/internal/domain/job"
	"github.com/ismobaga/apgn/internal/providers/llm"
)

// MetadataStage generates episode title, description, and subtitle.
type MetadataStage struct {
	episodes episode.Repository
	llm      llm.Provider
}

func NewMetadataStage(episodes episode.Repository, llm llm.Provider) *MetadataStage {
	return &MetadataStage{episodes: episodes, llm: llm}
}

func (s *MetadataStage) Run(ctx context.Context, payload job.Payload) error {
	ep, err := s.episodes.GetEpisode(payload.EpisodeID)
	if err != nil || ep == nil {
		return fmt.Errorf("get episode: %w", err)
	}

	if ep.Title != "" {
		// Metadata already set
		return nil
	}

	prompt := fmt.Sprintf(`Generate podcast episode metadata for topic: "%s", angle: "%s".
Return JSON with: title, subtitle, description (2-3 sentences for podcast directories).`, ep.Topic, ep.Angle)

	resp, err := s.llm.Generate(ctx, llm.GenerateRequest{
		UserPrompt:  prompt,
		Temperature: 0.7,
		MaxTokens:   500,
	})
	if err != nil {
		return fmt.Errorf("llm generate: %w", err)
	}

	ep.Description = resp.Text
	if ep.Title == "" {
		ep.Title = ep.Topic
	}
	return s.episodes.UpdateEpisode(ep)
}
