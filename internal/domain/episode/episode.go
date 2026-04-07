package episode

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusDraft          Status = "draft"
	StatusQueued         Status = "queued"
	StatusResearching    Status = "researching"
	StatusBriefReady     Status = "brief_ready"
	StatusScripting      Status = "scripting"
	StatusVoiceRendering Status = "voice_rendering"
	StatusRendering      Status = "rendering"
	StatusReadyForReview Status = "ready_for_review"
	StatusScheduled      Status = "scheduled"
	StatusPublished      Status = "published"
	StatusFailed         Status = "failed"
)

type Episode struct {
	ID                    uuid.UUID  `json:"id" db:"id"`
	ShowID                uuid.UUID  `json:"show_id" db:"show_id"`
	Status                Status     `json:"status" db:"status"`
	Topic                 string     `json:"topic" db:"topic"`
	Angle                 string     `json:"angle" db:"angle"`
	Title                 string     `json:"title" db:"title"`
	Subtitle              string     `json:"subtitle" db:"subtitle"`
	Description           string     `json:"description" db:"description"`
	Transcript            string     `json:"transcript" db:"transcript"`
	TargetDurationSeconds int        `json:"target_duration_seconds" db:"target_duration_seconds"`
	PlannedPublishAt      *time.Time `json:"planned_publish_at,omitempty" db:"planned_publish_at"`
	PublishedAt           *time.Time `json:"published_at,omitempty" db:"published_at"`
	ErrorMessage          string     `json:"error_message" db:"error_message"`
	CreatedAt             time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at" db:"updated_at"`
}

type Repository interface {
	CreateEpisode(episode *Episode) error
	GetEpisode(id uuid.UUID) (*Episode, error)
	ListEpisodes(showID *uuid.UUID, status *Status) ([]*Episode, error)
	UpdateEpisode(episode *Episode) error
	UpdateEpisodeStatus(id uuid.UUID, status Status, errMsg string) error
}
