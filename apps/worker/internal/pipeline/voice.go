package pipeline

import (
	"bytes"
	"context"
	"fmt"

	"github.com/ismobaga/apgn/internal/domain/asset"
	"github.com/ismobaga/apgn/internal/domain/job"
	"github.com/ismobaga/apgn/internal/domain/script"
	"github.com/ismobaga/apgn/internal/domain/show"
	"github.com/ismobaga/apgn/internal/providers/tts"
	"github.com/ismobaga/apgn/internal/storage"
)

// VoiceStage handles TTS voice rendering for script segments.
type VoiceStage struct {
	shows   show.Repository
	scripts script.Repository
	assets  asset.Repository
	tts     tts.Provider
	storage storage.Storage
}

func NewVoiceStage(
	shows show.Repository,
	scripts script.Repository,
	assets asset.Repository,
	tts tts.Provider,
	stor storage.Storage,
) *VoiceStage {
	return &VoiceStage{
		shows:   shows,
		scripts: scripts,
		assets:  assets,
		tts:     tts,
		storage: stor,
	}
}

func (s *VoiceStage) Run(ctx context.Context, payload job.Payload) error {
	draft, err := s.scripts.GetLatestDraft(payload.EpisodeID)
	if err != nil || draft == nil {
		return fmt.Errorf("get latest draft: %w", err)
	}

	spokenText := draft.SpokenText
	if spokenText == "" {
		spokenText = draft.FullText
	}
	if spokenText == "" {
		return fmt.Errorf("no spoken text available for TTS")
	}

	// Get host profile voice for this show
	hosts, err := s.shows.ListHostProfiles(payload.ShowID)
	if err != nil {
		return fmt.Errorf("list host profiles: %w", err)
	}

	voiceID := "default"
	var speakingRate float32 = 1.0
	if len(hosts) > 0 {
		voiceID = hosts[0].VoiceID
		speakingRate = float32(hosts[0].SpeakingRate)
	}

	resp, err := s.tts.Synthesize(ctx, tts.SynthesizeRequest{
		VoiceID: voiceID,
		Text:    spokenText,
		Format:  "mp3_44100_128",
		Rate:    speakingRate,
	})
	if err != nil {
		return fmt.Errorf("tts synthesize: %w", err)
	}

	key := fmt.Sprintf("episodes/%s/narration.mp3", payload.EpisodeID)
	if err := s.storage.Put(ctx, key, bytes.NewReader(resp.AudioBytes), int64(len(resp.AudioBytes)), "audio/mpeg"); err != nil {
		return fmt.Errorf("storage put: %w", err)
	}

	a := &asset.AudioAsset{
		EpisodeID:       payload.EpisodeID,
		AssetType:       asset.AssetTypeNarration,
		StorageKey:      key,
		MimeType:        "audio/mpeg",
		DurationSeconds: resp.DurationMS / 1000,
	}
	return s.assets.CreateAsset(a)
}
