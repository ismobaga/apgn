package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/ismobaga/apgn/internal/domain/source"
)

// SourcesHandler handles episode source endpoints.
type SourcesHandler struct {
	repo source.Repository
}

func NewSourcesHandler(repo source.Repository) *SourcesHandler {
	return &SourcesHandler{repo: repo}
}

func (h *SourcesHandler) List(w http.ResponseWriter, r *http.Request) {
	episodeID, err := parseUUID(chi.URLParam(r, "episodeID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid episode ID")
		return
	}
	sources, err := h.repo.ListSources(episodeID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if sources == nil {
		sources = []*source.EpisodeSource{}
	}
	respondJSON(w, http.StatusOK, sources)
}

func (h *SourcesHandler) ImportURL(w http.ResponseWriter, r *http.Request) {
	episodeID, err := parseUUID(chi.URLParam(r, "episodeID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid episode ID")
		return
	}
	var body struct {
		URL   string `json:"url"`
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.URL == "" {
		respondError(w, http.StatusBadRequest, "url is required")
		return
	}
	s := &source.EpisodeSource{
		EpisodeID:    episodeID,
		SourceType:   source.SourceTypeURL,
		SourceURL:    body.URL,
		SourceTitle:  body.Title,
		Selected:     true,
	}
	if err := h.repo.CreateSource(s); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, s)
}

func (h *SourcesHandler) ImportText(w http.ResponseWriter, r *http.Request) {
	episodeID, err := parseUUID(chi.URLParam(r, "episodeID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid episode ID")
		return
	}
	var body struct {
		Text   string `json:"text"`
		Title  string `json:"title"`
		Author string `json:"author"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Text == "" {
		respondError(w, http.StatusBadRequest, "text is required")
		return
	}
	s := &source.EpisodeSource{
		EpisodeID:     episodeID,
		SourceType:    source.SourceTypeText,
		SourceTitle:   body.Title,
		SourceAuthor:  body.Author,
		ExtractedText: body.Text,
		Selected:      true,
	}
	if err := h.repo.CreateSource(s); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, s)
}

func (h *SourcesHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "sourceID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid source ID")
		return
	}
	s, err := h.repo.GetSource(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if s == nil {
		respondError(w, http.StatusNotFound, "source not found")
		return
	}
	respondJSON(w, http.StatusOK, s)
}
