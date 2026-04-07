package pipeline

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/ismobaga/apgn/internal/domain/asset"
	"github.com/ismobaga/apgn/internal/domain/job"
)

// AudioStage assembles final audio using FFmpeg.
type AudioStage struct {
	assets asset.Repository
}

func NewAudioStage(assets asset.Repository) *AudioStage {
	return &AudioStage{assets: assets}
}

func (s *AudioStage) Run(ctx context.Context, payload job.Payload) error {
	assets, err := s.assets.ListAssets(payload.EpisodeID)
	if err != nil {
		return fmt.Errorf("list assets: %w", err)
	}

	var narrationKey string
	for _, a := range assets {
		if a.AssetType == asset.AssetTypeNarration {
			narrationKey = a.StorageKey
			break
		}
	}

	if narrationKey == "" {
		return fmt.Errorf("no narration asset found for episode %s", payload.EpisodeID)
	}

	// Check if FFmpeg is available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		// FFmpeg not available; record placeholder asset
		a := &asset.AudioAsset{
			EpisodeID:  payload.EpisodeID,
			AssetType:  asset.AssetTypeFinalMP3,
			StorageKey: narrationKey, // use narration as final for now
			MimeType:   "audio/mpeg",
		}
		return s.assets.CreateAsset(a)
	}

	outputKey := fmt.Sprintf("episodes/%s/final.mp3", payload.EpisodeID)

	// Use FFmpeg to copy/normalize narration to final output
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-i", narrationKey,
		"-af", "loudnorm=I=-16:TP=-1.5:LRA=11",
		"-codec:a", "libmp3lame",
		"-b:a", "128k",
		outputKey,
		"-y",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg error: %v\n%s", err, string(out))
	}

	a := &asset.AudioAsset{
		EpisodeID:  payload.EpisodeID,
		AssetType:  asset.AssetTypeFinalMP3,
		StorageKey: outputKey,
		MimeType:   "audio/mpeg",
	}
	return s.assets.CreateAsset(a)
}
