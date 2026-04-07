package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// PublishingHandler handles publishing endpoints.
type PublishingHandler struct{}

func NewPublishingHandler() *PublishingHandler {
	return &PublishingHandler{}
}

func (h *PublishingHandler) Publish(w http.ResponseWriter, r *http.Request) {
	episodeID, err := parseUUID(chi.URLParam(r, "episodeID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid episode ID")
		return
	}
	respondJSON(w, http.StatusAccepted, map[string]string{
		"status":     "publishing",
		"episode_id": episodeID.String(),
		"message":    "publishing job queued",
	})
}

func (h *PublishingHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	episodeID, err := parseUUID(chi.URLParam(r, "episodeID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid episode ID")
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{
		"episode_id": episodeID.String(),
		"status":     "unknown",
	})
}
