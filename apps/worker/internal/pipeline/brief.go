package pipeline

import (
	"context"
	"fmt"

	"github.com/ismobaga/apgn/internal/domain/brief"
	"github.com/ismobaga/apgn/internal/domain/job"
	"github.com/ismobaga/apgn/internal/domain/source"
	"github.com/ismobaga/apgn/internal/providers/llm"
)

// BriefStage handles the brief_generate stage.
// It reads episode sources and uses the LLM to produce an EpisodeBrief.
type BriefStage struct {
	briefs  brief.Repository
	sources source.Repository
	llm     llm.Provider
}

func NewBriefStage(briefs brief.Repository, sources source.Repository, llm llm.Provider) *BriefStage {
	return &BriefStage{briefs: briefs, sources: sources, llm: llm}
}

func (s *BriefStage) Run(ctx context.Context, payload job.Payload) error {
	srcs, err := s.sources.ListSources(payload.EpisodeID)
	if err != nil {
		return fmt.Errorf("list sources: %w", err)
	}

	// Build context for LLM from sources
	sourceContext := ""
	for _, src := range srcs {
		if src.Selected && src.ExtractedText != "" {
			sourceContext += fmt.Sprintf("Title: %s\n%s\n\n", src.SourceTitle, src.ExtractedText)
		}
	}

	systemPrompt := `You are a podcast producer. Based on the provided research sources, 
create a structured episode brief. Respond with a JSON object containing:
- audience: target audience description
- tone: episode tone
- angle: unique angle/perspective
- key_points: array of key points to cover
- claims: array of specific claims to make
- cta: call to action
- opening_hook: engaging opening hook
`
	userPrompt := fmt.Sprintf("Create a podcast episode brief based on these sources:\n\n%s", sourceContext)

	resp, err := s.llm.Generate(ctx, llm.GenerateRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  0.7,
		MaxTokens:    1500,
	})
	if err != nil {
		return fmt.Errorf("llm generate: %w", err)
	}

	// Check if brief exists
	existing, err := s.briefs.GetBriefByEpisode(payload.EpisodeID)
	if err != nil {
		return fmt.Errorf("get brief: %w", err)
	}

	if existing != nil {
		existing.Audience = "general audience"
		existing.Tone = "informative"
		existing.Angle = resp.Text
		return s.briefs.UpdateBrief(existing)
	}

	b := &brief.EpisodeBrief{
		EpisodeID: payload.EpisodeID,
		Audience:  "general audience",
		Tone:      "informative",
		Angle:     resp.Text,
	}
	return s.briefs.CreateBrief(b)
}
