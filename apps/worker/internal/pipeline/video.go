package pipeline

import (
	"context"

	"github.com/ismobaga/apgn/internal/domain/asset"
	"github.com/ismobaga/apgn/internal/domain/job"
)

// VideoStage packages the episode into a video format (e.g., audiogram).
type VideoStage struct {
	assets asset.Repository
}

func NewVideoStage(assets asset.Repository) *VideoStage {
	return &VideoStage{assets: assets}
}

func (s *VideoStage) Run(ctx context.Context, payload job.Payload) error {
	// Video packaging is optional in the MVP; mark as no-op if no video tools available.
	// In a full implementation this would use FFmpeg to create an audiogram.
	return nil
}
