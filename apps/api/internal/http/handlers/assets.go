package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/ismobaga/apgn/internal/domain/asset"
)

// AssetsHandler handles audio/video asset endpoints.
type AssetsHandler struct {
	repo asset.Repository
}

func NewAssetsHandler(repo asset.Repository) *AssetsHandler {
	return &AssetsHandler{repo: repo}
}

func (h *AssetsHandler) List(w http.ResponseWriter, r *http.Request) {
	episodeID, err := parseUUID(chi.URLParam(r, "episodeID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid episode ID")
		return
	}
	assets, err := h.repo.ListAssets(episodeID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if assets == nil {
		assets = []*asset.AudioAsset{}
	}
	respondJSON(w, http.StatusOK, assets)
}

func (h *AssetsHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "assetID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid asset ID")
		return
	}
	a, err := h.repo.GetAsset(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if a == nil {
		respondError(w, http.StatusNotFound, "asset not found")
		return
	}
	respondJSON(w, http.StatusOK, a)
}

func (h *AssetsHandler) RenderAudio(w http.ResponseWriter, r *http.Request) {
	episodeID, err := parseUUID(chi.URLParam(r, "episodeID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid episode ID")
		return
	}
	respondJSON(w, http.StatusAccepted, map[string]string{
		"status":     "queued",
		"episode_id": episodeID.String(),
		"message":    "audio render job queued",
	})
}

func (h *AssetsHandler) RenderVideo(w http.ResponseWriter, r *http.Request) {
	episodeID, err := parseUUID(chi.URLParam(r, "episodeID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid episode ID")
		return
	}
	respondJSON(w, http.StatusAccepted, map[string]string{
		"status":     "queued",
		"episode_id": episodeID.String(),
		"message":    "video render job queued",
	})
}
