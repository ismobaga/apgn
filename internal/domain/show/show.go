package show

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusActive   Status = "active"
	StatusArchived Status = "archived"
	StatusPaused   Status = "paused"
)

type Show struct {
	ID                     uuid.UUID  `json:"id" db:"id"`
	Slug                   string     `json:"slug" db:"slug"`
	Name                   string     `json:"name" db:"name"`
	Description            string     `json:"description" db:"description"`
	Language               string     `json:"language" db:"language"`
	Niche                  string     `json:"niche" db:"niche"`
	Tone                   string     `json:"tone" db:"tone"`
	Status                 Status     `json:"status" db:"status"`
	Cadence                string     `json:"cadence" db:"cadence"`
	DefaultDurationMinutes int        `json:"default_duration_minutes" db:"default_duration_minutes"`
	DefaultFormat          string     `json:"default_format" db:"default_format"`
	IntroAssetID           *uuid.UUID `json:"intro_asset_id,omitempty" db:"intro_asset_id"`
	OutroAssetID           *uuid.UUID `json:"outro_asset_id,omitempty" db:"outro_asset_id"`
	CoverAssetID           *uuid.UUID `json:"cover_asset_id,omitempty" db:"cover_asset_id"`
	CreatedAt              time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at" db:"updated_at"`
}

type HostProfile struct {
	ID                 uuid.UUID `json:"id" db:"id"`
	ShowID             uuid.UUID `json:"show_id" db:"show_id"`
	DisplayName        string    `json:"display_name" db:"display_name"`
	PersonaSummary     string    `json:"persona_summary" db:"persona_summary"`
	SpeakingStyle      string    `json:"speaking_style" db:"speaking_style"`
	Provider           string    `json:"provider" db:"provider"`
	VoiceID            string    `json:"voice_id" db:"voice_id"`
	SpeakingRate       float64   `json:"speaking_rate" db:"speaking_rate"`
	PronunciationRules []byte    `json:"pronunciation_rules_json" db:"pronunciation_rules_json"`
	PromptRules        []byte    `json:"prompt_rules_json" db:"prompt_rules_json"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

type Repository interface {
	CreateShow(show *Show) error
	GetShow(id uuid.UUID) (*Show, error)
	GetShowBySlug(slug string) (*Show, error)
	ListShows() ([]*Show, error)
	UpdateShow(show *Show) error

	CreateHostProfile(host *HostProfile) error
	GetHostProfile(id uuid.UUID) (*HostProfile, error)
	ListHostProfiles(showID uuid.UUID) ([]*HostProfile, error)
}
