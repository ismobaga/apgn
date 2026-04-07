package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/ismobaga/apgn/internal/domain/brief"
)

// BriefsHandler handles episode brief endpoints.
type BriefsHandler struct {
	repo brief.Repository
}

func NewBriefsHandler(repo brief.Repository) *BriefsHandler {
	return &BriefsHandler{repo: repo}
}

func (h *BriefsHandler) Get(w http.ResponseWriter, r *http.Request) {
	episodeID, err := parseUUID(chi.URLParam(r, "episodeID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid episode ID")
		return
	}
	b, err := h.repo.GetBriefByEpisode(episodeID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if b == nil {
		respondError(w, http.StatusNotFound, "brief not found")
		return
	}
	respondJSON(w, http.StatusOK, b)
}

func (h *BriefsHandler) Generate(w http.ResponseWriter, r *http.Request) {
	episodeID, err := parseUUID(chi.URLParam(r, "episodeID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid episode ID")
		return
	}
	// Check if brief already exists
	existing, err := h.repo.GetBriefByEpisode(episodeID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if existing != nil {
		respondJSON(w, http.StatusOK, existing)
		return
	}
	// Create a placeholder brief to be filled by the worker pipeline
	b := &brief.EpisodeBrief{
		EpisodeID: episodeID,
		Audience:  "general audience",
		Tone:      "informative",
	}
	if err := h.repo.CreateBrief(b); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, b)
}
