package asset

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AssetType string

const (
	AssetTypeIntro      AssetType = "intro"
	AssetTypeOutro      AssetType = "outro"
	AssetTypeNarration  AssetType = "narration"
	AssetTypeFinalMP3   AssetType = "final_mp3"
	AssetTypeFinalVideo AssetType = "final_video"
	AssetTypeCoverImage AssetType = "cover_image"
	AssetTypeMusicBed   AssetType = "music_bed"
)

type AudioAsset struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	EpisodeID       uuid.UUID       `json:"episode_id" db:"episode_id"`
	AssetType       AssetType       `json:"asset_type" db:"asset_type"`
	StorageKey      string          `json:"storage_key" db:"storage_key"`
	MimeType        string          `json:"mime_type" db:"mime_type"`
	DurationSeconds int             `json:"duration_seconds" db:"duration_seconds"`
	Metadata        json.RawMessage `json:"metadata_json" db:"metadata_json"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
}

type Repository interface {
	CreateAsset(asset *AudioAsset) error
	GetAsset(id uuid.UUID) (*AudioAsset, error)
	ListAssets(episodeID uuid.UUID) ([]*AudioAsset, error)
}
