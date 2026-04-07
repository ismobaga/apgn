package job

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Stage string

const (
	StageTopicSelect           Stage = "topic_select"
	StageResearchCollect       Stage = "research_collect"
	StageResearchFilter        Stage = "research_filter"
	StageBriefGenerate         Stage = "brief_generate"
	StageScriptOutlineGenerate Stage = "script_outline_generate"
	StageScriptGenerate        Stage = "script_generate"
	StageScriptRewriteForAudio Stage = "script_rewrite_for_audio"
	StageVoiceRenderSegments   Stage = "voice_render_segments"
	StageAudioAssemble         Stage = "audio_assemble"
	StageTranscriptFinalize    Stage = "transcript_finalize"
	StageMetadataGenerate      Stage = "metadata_generate"
	StageVideoPackage          Stage = "video_package"
	StagePublishPrepare        Stage = "publish_prepare"
	StagePublishDeliver        Stage = "publish_deliver"
)

type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusRetrying  JobStatus = "retrying"
)

type JobRun struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	EpisodeID    uuid.UUID       `json:"episode_id" db:"episode_id"`
	Stage        Stage           `json:"stage" db:"stage"`
	Status       JobStatus       `json:"status" db:"status"`
	Attempt      int             `json:"attempt" db:"attempt"`
	Input        json.RawMessage `json:"input_json" db:"input_json"`
	Output       json.RawMessage `json:"output_json" db:"output_json"`
	StartedAt    *time.Time      `json:"started_at,omitempty" db:"started_at"`
	FinishedAt   *time.Time      `json:"finished_at,omitempty" db:"finished_at"`
	ErrorMessage string          `json:"error_message" db:"error_message"`
}

type Payload struct {
	EpisodeID     uuid.UUID `json:"episode_id"`
	ShowID        uuid.UUID `json:"show_id"`
	Stage         Stage     `json:"stage"`
	Attempt       int       `json:"attempt"`
	CorrelationID uuid.UUID `json:"correlation_id"`
}

type Repository interface {
	CreateJobRun(job *JobRun) error
	GetJobRun(id uuid.UUID) (*JobRun, error)
	ListJobRuns(episodeID uuid.UUID) ([]*JobRun, error)
	UpdateJobRun(job *JobRun) error
}
