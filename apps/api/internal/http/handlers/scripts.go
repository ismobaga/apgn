package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/ismobaga/apgn/internal/domain/script"
)

// ScriptsHandler handles script draft endpoints.
type ScriptsHandler struct {
	repo script.Repository
}

func NewScriptsHandler(repo script.Repository) *ScriptsHandler {
	return &ScriptsHandler{repo: repo}
}

func (h *ScriptsHandler) List(w http.ResponseWriter, r *http.Request) {
	episodeID, err := parseUUID(chi.URLParam(r, "episodeID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid episode ID")
		return
	}
	drafts, err := h.repo.ListDrafts(episodeID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if drafts == nil {
		drafts = []*script.ScriptDraft{}
	}
	respondJSON(w, http.StatusOK, drafts)
}

func (h *ScriptsHandler) Get(w http.ResponseWriter, r *http.Request) {
	episodeID, err := parseUUID(chi.URLParam(r, "episodeID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid episode ID")
		return
	}
	draft, err := h.repo.GetLatestDraft(episodeID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if draft == nil {
		respondError(w, http.StatusNotFound, "script not found")
		return
	}
	respondJSON(w, http.StatusOK, draft)
}

func (h *ScriptsHandler) Generate(w http.ResponseWriter, r *http.Request) {
	episodeID, err := parseUUID(chi.URLParam(r, "episodeID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid episode ID")
		return
	}
	draft := &script.ScriptDraft{
		EpisodeID: episodeID,
		Format:    "solo",
		Status:    script.ScriptStatusDraft,
	}
	if err := h.repo.CreateDraft(draft); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, draft)
}

func (h *ScriptsHandler) Approve(w http.ResponseWriter, r *http.Request) {
	draftID, err := parseUUID(chi.URLParam(r, "draftID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid draft ID")
		return
	}
	draft, err := h.repo.GetDraft(draftID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if draft == nil {
		respondError(w, http.StatusNotFound, "draft not found")
		return
	}
	draft.Status = script.ScriptStatusApproved
	if err := h.repo.UpdateDraft(draft); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, draft)
}
