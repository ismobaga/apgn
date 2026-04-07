package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/ismobaga/apgn/apps/worker/internal/pipeline"
	"github.com/ismobaga/apgn/internal/domain/asset"
	"github.com/ismobaga/apgn/internal/domain/brief"
	"github.com/ismobaga/apgn/internal/domain/episode"
	"github.com/ismobaga/apgn/internal/domain/job"
	"github.com/ismobaga/apgn/internal/domain/script"
	"github.com/ismobaga/apgn/internal/domain/show"
	"github.com/ismobaga/apgn/internal/domain/source"
	"github.com/ismobaga/apgn/internal/orchestrator"
	"github.com/ismobaga/apgn/internal/providers/llm"
	"github.com/ismobaga/apgn/internal/providers/tts"
	"github.com/ismobaga/apgn/internal/queue"
	"github.com/ismobaga/apgn/internal/storage"
)

const queueName = "pipeline"

// Repos aggregates all repository dependencies for the dispatcher.
type Repos struct {
	Shows    show.Repository
	Episodes episode.Repository
	Sources  source.Repository
	Briefs   brief.Repository
	Scripts  script.Repository
	Assets   asset.Repository
	Jobs     job.Repository
}

// Dispatcher receives queue messages and routes them to the appropriate pipeline stage.
type Dispatcher struct {
	repos        Repos
	orch         *orchestrator.Orchestrator
	queue        queue.Queue
	llm          llm.Provider
	tts          tts.Provider
	storage      storage.Storage
}

func NewDispatcher(
	repos Repos,
	orch *orchestrator.Orchestrator,
	q queue.Queue,
	llmProvider llm.Provider,
	ttsProvider tts.Provider,
	stor storage.Storage,
) *Dispatcher {
	return &Dispatcher{
		repos:   repos,
		orch:    orch,
		queue:   q,
		llm:     llmProvider,
		tts:     ttsProvider,
		storage: stor,
	}
}

// Run starts the dispatcher loop, blocking until context is cancelled.
func (d *Dispatcher) Run(ctx context.Context) {
	log.Println("worker dispatcher started")
	for {
		select {
		case <-ctx.Done():
			log.Println("worker dispatcher stopped")
			return
		default:
		}

		msg, err := d.queue.Dequeue(ctx, queueName)
		if err != nil {
			log.Printf("dequeue error: %v", err)
			continue
		}
		if msg == nil {
			continue
		}

		if err := d.handle(ctx, msg); err != nil {
			log.Printf("handle message error: %v", err)
		}
	}
}

func (d *Dispatcher) handle(ctx context.Context, msg *queue.Message) error {
	var payload job.Payload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	log.Printf("processing stage=%s episode=%s attempt=%d", payload.Stage, payload.EpisodeID, payload.Attempt)

	// Mark job as running
	if _, err := d.orch.MarkRunning(ctx, payload); err != nil {
		log.Printf("mark running error: %v", err)
	}

	stageErr := d.runStage(ctx, payload)
	output := json.RawMessage(`{}`)

	if stageErr != nil {
		log.Printf("stage %s failed: %v", payload.Stage, stageErr)
		if err := d.orch.FailStage(ctx, payload, stageErr.Error()); err != nil {
			log.Printf("fail stage error: %v", err)
		}
	} else {
		if err := d.orch.AdvanceStage(ctx, payload, output); err != nil {
			log.Printf("advance stage error: %v", err)
		}
	}

	// Acknowledge the message regardless of outcome to avoid reprocessing
	return d.queue.Acknowledge(ctx, queueName, msg.ID)
}

func (d *Dispatcher) runStage(ctx context.Context, payload job.Payload) error {
	switch payload.Stage {
	case job.StageResearchCollect, job.StageResearchFilter:
		stage := pipeline.NewResearchStage(d.repos.Episodes)
		return stage.Run(ctx, payload)

	case job.StageBriefGenerate:
		if d.llm == nil {
			return nil // Skip if no LLM provider configured
		}
		stage := pipeline.NewBriefStage(d.repos.Briefs, d.repos.Sources, d.llm)
		return stage.Run(ctx, payload)

	case job.StageScriptOutlineGenerate, job.StageScriptGenerate:
		if d.llm == nil {
			return nil
		}
		stage := pipeline.NewScriptStage(d.repos.Episodes, d.repos.Briefs, d.repos.Scripts, d.llm)
		return stage.Run(ctx, payload)

	case job.StageScriptRewriteForAudio:
		if d.llm == nil {
			return nil
		}
		stage := pipeline.NewScriptRewriteStage(d.repos.Scripts, d.llm)
		return stage.Run(ctx, payload)

	case job.StageVoiceRenderSegments:
		if d.tts == nil || d.storage == nil {
			return nil
		}
		stage := pipeline.NewVoiceStage(d.repos.Shows, d.repos.Scripts, d.repos.Assets, d.tts, d.storage)
		return stage.Run(ctx, payload)

	case job.StageAudioAssemble:
		stage := pipeline.NewAudioStage(d.repos.Assets)
		return stage.Run(ctx, payload)

	case job.StageTranscriptFinalize:
		return nil // No-op for MVP

	case job.StageMetadataGenerate:
		if d.llm == nil {
			return nil
		}
		stage := pipeline.NewMetadataStage(d.repos.Episodes, d.llm)
		return stage.Run(ctx, payload)

	case job.StageVideoPackage:
		stage := pipeline.NewVideoStage(d.repos.Assets)
		return stage.Run(ctx, payload)

	case job.StagePublishPrepare:
		stage := pipeline.NewPublishStage(d.repos.Episodes)
		return stage.RunPrepare(ctx, payload)

	case job.StagePublishDeliver:
		stage := pipeline.NewPublishStage(d.repos.Episodes)
		return stage.RunDeliver(ctx, payload)

	default:
		return fmt.Errorf("unknown stage: %s", payload.Stage)
	}
}
