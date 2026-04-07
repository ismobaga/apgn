package pipeline

import (
	"context"
	"fmt"

	"github.com/ismobaga/apgn/internal/domain/episode"
	"github.com/ismobaga/apgn/internal/domain/job"
)

// ResearchStage handles research collection and filtering.
type ResearchStage struct {
	episodes episode.Repository
}

func NewResearchStage(episodes episode.Repository) *ResearchStage {
	return &ResearchStage{episodes: episodes}
}

func (s *ResearchStage) Run(ctx context.Context, payload job.Payload) error {
	ep, err := s.episodes.GetEpisode(payload.EpisodeID)
	if err != nil || ep == nil {
		return fmt.Errorf("get episode: %w", err)
	}
	// Research collection: in a real implementation this would crawl URLs,
	// search news APIs, etc. For the MVP we log and proceed.
	return nil
}
