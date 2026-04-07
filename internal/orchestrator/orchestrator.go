package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/ismobaga/apgn/internal/domain/episode"
	"github.com/ismobaga/apgn/internal/domain/job"
	"github.com/ismobaga/apgn/internal/queue"
)

const defaultQueueName = "pipeline"

// Repos aggregates the repository dependencies.
type Repos struct {
	Episodes episode.Repository
	Jobs     job.Repository
}

// Orchestrator manages the episode pipeline state machine.
// It validates prerequisites, writes job_runs, advances episode status, and enqueues the next stage.
type Orchestrator struct {
	repos Repos
	queue queue.Queue
}

func New(repos Repos, q queue.Queue) *Orchestrator {
	return &Orchestrator{repos: repos, queue: q}
}

// stageOrder defines the pipeline sequence.
var stageOrder = []job.Stage{
	job.StageResearchCollect,
	job.StageResearchFilter,
	job.StageBriefGenerate,
	job.StageScriptOutlineGenerate,
	job.StageScriptGenerate,
	job.StageScriptRewriteForAudio,
	job.StageVoiceRenderSegments,
	job.StageAudioAssemble,
	job.StageTranscriptFinalize,
	job.StageMetadataGenerate,
	job.StageVideoPackage,
	job.StagePublishPrepare,
	job.StagePublishDeliver,
}

// stageToEpisodeStatus maps pipeline stages to episode status values.
var stageToEpisodeStatus = map[job.Stage]episode.Status{
	job.StageResearchCollect:       episode.StatusResearching,
	job.StageResearchFilter:        episode.StatusResearching,
	job.StageBriefGenerate:         episode.StatusResearching,
	job.StageScriptOutlineGenerate: episode.StatusScripting,
	job.StageScriptGenerate:        episode.StatusScripting,
	job.StageScriptRewriteForAudio: episode.StatusScripting,
	job.StageVoiceRenderSegments:   episode.StatusVoiceRendering,
	job.StageAudioAssemble:         episode.StatusRendering,
	job.StageTranscriptFinalize:    episode.StatusRendering,
	job.StageMetadataGenerate:      episode.StatusRendering,
	job.StageVideoPackage:          episode.StatusRendering,
	job.StagePublishPrepare:        episode.StatusScheduled,
	job.StagePublishDeliver:        episode.StatusPublished,
}

// QueueEpisode starts the pipeline for an episode by enqueuing the first stage.
func (o *Orchestrator) QueueEpisode(ctx context.Context, ep *episode.Episode) error {
	if err := o.repos.Episodes.UpdateEpisodeStatus(ep.ID, episode.StatusQueued, ""); err != nil {
		return fmt.Errorf("update episode status: %w", err)
	}
	return o.enqueueStage(ctx, ep.ID, ep.ShowID, job.StageResearchCollect, 1)
}

// AdvanceStage is called when a stage completes successfully. It updates job state,
// advances episode status, and enqueues the next stage.
func (o *Orchestrator) AdvanceStage(ctx context.Context, payload job.Payload, output json.RawMessage) error {
	now := time.Now()

	// Update existing job run if we can find it
	runs, err := o.repos.Jobs.ListJobRuns(payload.EpisodeID)
	if err != nil {
		return err
	}
	for _, r := range runs {
		if r.Stage == payload.Stage && r.Status == job.JobStatusRunning {
			r.Status = job.JobStatusCompleted
			r.FinishedAt = &now
			r.Output = output
			if err := o.repos.Jobs.UpdateJobRun(r); err != nil {
				return err
			}
			break
		}
	}

	// Determine next stage
	nextStage, ok := nextPipelineStage(payload.Stage)
	if !ok {
		// Pipeline complete
		return o.repos.Episodes.UpdateEpisodeStatus(payload.EpisodeID, episode.StatusPublished, "")
	}

	// Advance episode status
	if epStatus, ok := stageToEpisodeStatus[nextStage]; ok {
		if err := o.repos.Episodes.UpdateEpisodeStatus(payload.EpisodeID, epStatus, ""); err != nil {
			return err
		}
	}

	return o.enqueueStage(ctx, payload.EpisodeID, payload.ShowID, nextStage, 1)
}

// FailStage marks a job as failed and updates episode status.
func (o *Orchestrator) FailStage(ctx context.Context, payload job.Payload, errMsg string) error {
	now := time.Now()

	runs, err := o.repos.Jobs.ListJobRuns(payload.EpisodeID)
	if err != nil {
		return err
	}
	for _, r := range runs {
		if r.Stage == payload.Stage && r.Status == job.JobStatusRunning {
			r.Status = job.JobStatusFailed
			r.FinishedAt = &now
			r.ErrorMessage = errMsg
			if err := o.repos.Jobs.UpdateJobRun(r); err != nil {
				return err
			}
			break
		}
	}

	return o.repos.Episodes.UpdateEpisodeStatus(payload.EpisodeID, episode.StatusFailed, errMsg)
}

// MarkRunning creates or updates a job_run record when a worker picks up a stage.
func (o *Orchestrator) MarkRunning(ctx context.Context, payload job.Payload) (*job.JobRun, error) {
	now := time.Now()
	j := &job.JobRun{
		EpisodeID: payload.EpisodeID,
		Stage:     payload.Stage,
		Status:    job.JobStatusRunning,
		Attempt:   payload.Attempt,
		StartedAt: &now,
	}
	if err := o.repos.Jobs.CreateJobRun(j); err != nil {
		return nil, err
	}
	return j, nil
}

func (o *Orchestrator) enqueueStage(ctx context.Context, episodeID, showID uuid.UUID, stage job.Stage, attempt int) error {
	payload := job.Payload{
		EpisodeID:     episodeID,
		ShowID:        showID,
		Stage:         stage,
		Attempt:       attempt,
		CorrelationID: uuid.New(),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return o.queue.Enqueue(ctx, defaultQueueName, data)
}

func nextPipelineStage(current job.Stage) (job.Stage, bool) {
	for i, s := range stageOrder {
		if s == current {
			if i+1 < len(stageOrder) {
				return stageOrder[i+1], true
			}
			return "", false
		}
	}
	return "", false
}
