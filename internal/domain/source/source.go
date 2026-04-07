package source

import (
	"time"

	"github.com/google/uuid"
)

type SourceType string

const (
	SourceTypeURL    SourceType = "url"
	SourceTypeText   SourceType = "text"
	SourceTypeUpload SourceType = "upload"
	SourceTypeFeed   SourceType = "feed"
)

type EpisodeSource struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	EpisodeID         uuid.UUID  `json:"episode_id" db:"episode_id"`
	SourceType        SourceType `json:"source_type" db:"source_type"`
	SourceURL         string     `json:"source_url" db:"source_url"`
	SourceTitle       string     `json:"source_title" db:"source_title"`
	SourceAuthor      string     `json:"source_author" db:"source_author"`
	SourcePublishedAt *time.Time `json:"source_published_at,omitempty" db:"source_published_at"`
	ExtractedText     string     `json:"extracted_text" db:"extracted_text"`
	Summary           string     `json:"summary" db:"summary"`
	RelevanceScore    float64    `json:"relevance_score" db:"relevance_score"`
	TrustScore        float64    `json:"trust_score" db:"trust_score"`
	Selected          bool       `json:"selected" db:"selected"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
}

type Repository interface {
	CreateSource(source *EpisodeSource) error
	GetSource(id uuid.UUID) (*EpisodeSource, error)
	ListSources(episodeID uuid.UUID) ([]*EpisodeSource, error)
	UpdateSource(source *EpisodeSource) error
}
