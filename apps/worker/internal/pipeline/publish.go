package pipeline

import (
	"context"
	"fmt"

	"github.com/ismobaga/apgn/internal/domain/episode"
	"github.com/ismobaga/apgn/internal/domain/job"
)

// PublishStage handles preparing and delivering episode to publishing platforms.
type PublishStage struct {
	episodes episode.Repository
}

func NewPublishStage(episodes episode.Repository) *PublishStage {
	return &PublishStage{episodes: episodes}
}

func (s *PublishStage) RunPrepare(ctx context.Context, payload job.Payload) error {
	ep, err := s.episodes.GetEpisode(payload.EpisodeID)
	if err != nil || ep == nil {
		return fmt.Errorf("get episode: %w", err)
	}
	// Prepare RSS feed entry, platform-specific metadata, etc.
	return nil
}

func (s *PublishStage) RunDeliver(ctx context.Context, payload job.Payload) error {
	ep, err := s.episodes.GetEpisode(payload.EpisodeID)
	if err != nil || ep == nil {
		return fmt.Errorf("get episode: %w", err)
	}
	// Deliver to podcast hosting platform via API
	return s.episodes.UpdateEpisodeStatus(ep.ID, episode.StatusPublished, "")
}
