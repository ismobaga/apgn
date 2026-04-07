package script

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ScriptStatus string

const (
	ScriptStatusDraft    ScriptStatus = "draft"
	ScriptStatusApproved ScriptStatus = "approved"
	ScriptStatusRejected ScriptStatus = "rejected"
)

type ScriptDraft struct {
	ID         uuid.UUID       `json:"id" db:"id"`
	EpisodeID  uuid.UUID       `json:"episode_id" db:"episode_id"`
	Version    int             `json:"version" db:"version"`
	Format     string          `json:"format" db:"format"`
	Outline    json.RawMessage `json:"outline_json" db:"outline_json"`
	Sections   json.RawMessage `json:"sections_json" db:"sections_json"`
	FullText   string          `json:"full_text" db:"full_text"`
	SpokenText string          `json:"spoken_text" db:"spoken_text"`
	Status     ScriptStatus    `json:"status" db:"status"`
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
}

type Repository interface {
	CreateDraft(draft *ScriptDraft) error
	GetDraft(id uuid.UUID) (*ScriptDraft, error)
	GetLatestDraft(episodeID uuid.UUID) (*ScriptDraft, error)
	ListDrafts(episodeID uuid.UUID) ([]*ScriptDraft, error)
	UpdateDraft(draft *ScriptDraft) error
}
