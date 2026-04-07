package pipeline

import (
	"context"
	"fmt"

	"github.com/ismobaga/apgn/internal/domain/brief"
	"github.com/ismobaga/apgn/internal/domain/episode"
	"github.com/ismobaga/apgn/internal/domain/job"
	"github.com/ismobaga/apgn/internal/domain/script"
	"github.com/ismobaga/apgn/internal/providers/llm"
)

// ScriptStage handles script generation stages.
type ScriptStage struct {
	episodes episode.Repository
	briefs   brief.Repository
	scripts  script.Repository
	llm      llm.Provider
}

func NewScriptStage(
	episodes episode.Repository,
	briefs brief.Repository,
	scripts script.Repository,
	llm llm.Provider,
) *ScriptStage {
	return &ScriptStage{episodes: episodes, briefs: briefs, scripts: scripts, llm: llm}
}

func (s *ScriptStage) Run(ctx context.Context, payload job.Payload) error {
	ep, err := s.episodes.GetEpisode(payload.EpisodeID)
	if err != nil || ep == nil {
		return fmt.Errorf("get episode: %w", err)
	}

	b, err := s.briefs.GetBriefByEpisode(payload.EpisodeID)
	if err != nil {
		return fmt.Errorf("get brief: %w", err)
	}

	briefContext := "No brief available"
	if b != nil {
		briefContext = fmt.Sprintf("Angle: %s\nAudience: %s\nTone: %s\nOpening Hook: %s",
			b.Angle, b.Audience, b.Tone, b.OpeningHook)
	}

	systemPrompt := `You are a professional podcast scriptwriter. 
Write a complete podcast script based on the episode brief. 
Include an engaging introduction, main content sections, and a conclusion with CTA.
Format the script naturally for spoken audio.`

	userPrompt := fmt.Sprintf(`Topic: %s
Angle: %s

Brief:
%s

Write a complete podcast script for this episode.`, ep.Topic, ep.Angle, briefContext)

	resp, err := s.llm.Generate(ctx, llm.GenerateRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  0.8,
		MaxTokens:    4000,
	})
	if err != nil {
		return fmt.Errorf("llm generate: %w", err)
	}

	draft := &script.ScriptDraft{
		EpisodeID:  payload.EpisodeID,
		Format:     "solo",
		FullText:   resp.Text,
		SpokenText: resp.Text,
		Status:     script.ScriptStatusDraft,
	}
	return s.scripts.CreateDraft(draft)
}

// ScriptRewriteStage rewrites a script for optimal audio delivery.
type ScriptRewriteStage struct {
	scripts script.Repository
	llm     llm.Provider
}

func NewScriptRewriteStage(scripts script.Repository, llm llm.Provider) *ScriptRewriteStage {
	return &ScriptRewriteStage{scripts: scripts, llm: llm}
}

func (s *ScriptRewriteStage) Run(ctx context.Context, payload job.Payload) error {
	draft, err := s.scripts.GetLatestDraft(payload.EpisodeID)
	if err != nil || draft == nil {
		return fmt.Errorf("get latest draft: %w", err)
	}

	systemPrompt := `You are an audio script editor. 
Rewrite the script to optimize it for text-to-speech delivery:
- Remove markdown, headers, and stage directions
- Make sentences natural and conversational
- Add pronunciation guides for difficult words
- Ensure smooth transitions between sections`

	resp, err := s.llm.Generate(ctx, llm.GenerateRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   fmt.Sprintf("Rewrite this script for audio:\n\n%s", draft.FullText),
		Temperature:  0.6,
		MaxTokens:    4000,
	})
	if err != nil {
		return fmt.Errorf("llm rewrite: %w", err)
	}

	draft.SpokenText = resp.Text
	return s.scripts.UpdateDraft(draft)
}
