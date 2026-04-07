package events

import "github.com/google/uuid"

const (
	// Episode pipeline events
	EventEpisodeQueued         = "episode.queued"
	EventEpisodeStageCompleted = "episode.stage.completed"
	EventEpisodeStageFailed    = "episode.stage.failed"
	EventEpisodePublished      = "episode.published"
)

type EpisodeEvent struct {
	Type      string    `json:"type"`
	EpisodeID uuid.UUID `json:"episode_id"`
	ShowID    uuid.UUID `json:"show_id"`
	Stage     string    `json:"stage,omitempty"`
	Error     string    `json:"error,omitempty"`
}
