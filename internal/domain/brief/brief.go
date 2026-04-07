package brief

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type EpisodeBrief struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	EpisodeID   uuid.UUID       `json:"episode_id" db:"episode_id"`
	Audience    string          `json:"audience" db:"audience"`
	Tone        string          `json:"tone" db:"tone"`
	Angle       string          `json:"angle" db:"angle"`
	KeyPoints   json.RawMessage `json:"key_points_json" db:"key_points_json"`
	Claims      json.RawMessage `json:"claims_json" db:"claims_json"`
	CTA         string          `json:"cta" db:"cta"`
	OpeningHook string          `json:"opening_hook" db:"opening_hook"`
	Constraints json.RawMessage `json:"constraints_json" db:"constraints_json"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
}

type Repository interface {
	CreateBrief(brief *EpisodeBrief) error
	GetBriefByEpisode(episodeID uuid.UUID) (*EpisodeBrief, error)
	UpdateBrief(brief *EpisodeBrief) error
}
