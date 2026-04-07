package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ismobaga/apgn/internal/domain/episode"
	"github.com/ismobaga/apgn/internal/orchestrator"
)

// EpisodesHandler handles episode endpoints.
type EpisodesHandler struct {
	repo         episode.Repository
	orchestrator *orchestrator.Orchestrator
}

func NewEpisodesHandler(repo episode.Repository, orch *orchestrator.Orchestrator) *EpisodesHandler {
	return &EpisodesHandler{repo: repo, orchestrator: orch}
}

func (h *EpisodesHandler) List(w http.ResponseWriter, r *http.Request) {
	var showIDPtr *uuid.UUID
	if sid := r.URL.Query().Get("show_id"); sid != "" {
		id, err := uuid.Parse(sid)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid show_id")
			return
		}
		showIDPtr = &id
	}
	var statusPtr *episode.Status
	if s := r.URL.Query().Get("status"); s != "" {
		st := episode.Status(s)
		statusPtr = &st
	}

	eps, err := h.repo.ListEpisodes(showIDPtr, statusPtr)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if eps == nil {
		eps = []*episode.Episode{}
	}
	respondJSON(w, http.StatusOK, eps)
}

func (h *EpisodesHandler) Create(w http.ResponseWriter, r *http.Request) {
	showID, err := parseUUID(chi.URLParam(r, "showID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid show ID")
		return
	}
	var e episode.Episode
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	e.ShowID = showID
	if err := h.repo.CreateEpisode(&e); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, e)
}

func (h *EpisodesHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "episodeID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid episode ID")
		return
	}
	e, err := h.repo.GetEpisode(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if e == nil {
		respondError(w, http.StatusNotFound, "episode not found")
		return
	}
	respondJSON(w, http.StatusOK, e)
}

func (h *EpisodesHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "episodeID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid episode ID")
		return
	}
	existing, err := h.repo.GetEpisode(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if existing == nil {
		respondError(w, http.StatusNotFound, "episode not found")
		return
	}
	if err := json.NewDecoder(r.Body).Decode(existing); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	existing.ID = id
	if err := h.repo.UpdateEpisode(existing); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, existing)
}

func (h *EpisodesHandler) Queue(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "episodeID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid episode ID")
		return
	}
	e, err := h.repo.GetEpisode(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if e == nil {
		respondError(w, http.StatusNotFound, "episode not found")
		return
	}
	if err := h.orchestrator.QueueEpisode(r.Context(), e); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusAccepted, map[string]string{"status": "queued"})
}

func (h *EpisodesHandler) Retry(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "episodeID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid episode ID")
		return
	}
	e, err := h.repo.GetEpisode(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if e == nil {
		respondError(w, http.StatusNotFound, "episode not found")
		return
	}
	if e.Status != episode.StatusFailed {
		respondError(w, http.StatusBadRequest, "episode is not in failed state")
		return
	}
	if err := h.orchestrator.QueueEpisode(r.Context(), e); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusAccepted, map[string]string{"status": "retrying"})
}
