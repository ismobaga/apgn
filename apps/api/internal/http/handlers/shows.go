package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ismobaga/apgn/internal/domain/show"
)

// ShowsHandler handles show and host profile endpoints.
type ShowsHandler struct {
	repo show.Repository
}

func NewShowsHandler(repo show.Repository) *ShowsHandler {
	return &ShowsHandler{repo: repo}
}

func (h *ShowsHandler) List(w http.ResponseWriter, r *http.Request) {
	shows, err := h.repo.ListShows()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if shows == nil {
		shows = []*show.Show{}
	}
	respondJSON(w, http.StatusOK, shows)
}

func (h *ShowsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var s show.Show
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	s.Slug = slugify(s.Name)
	if err := h.repo.CreateShow(&s); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, s)
}

func (h *ShowsHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "showID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid show ID")
		return
	}
	s, err := h.repo.GetShow(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if s == nil {
		respondError(w, http.StatusNotFound, "show not found")
		return
	}
	respondJSON(w, http.StatusOK, s)
}

func (h *ShowsHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "showID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid show ID")
		return
	}
	existing, err := h.repo.GetShow(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if existing == nil {
		respondError(w, http.StatusNotFound, "show not found")
		return
	}
	if err := json.NewDecoder(r.Body).Decode(existing); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	existing.ID = id
	if err := h.repo.UpdateShow(existing); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, existing)
}

// Host profiles

func (h *ShowsHandler) ListHostProfiles(w http.ResponseWriter, r *http.Request) {
	showID, err := parseUUID(chi.URLParam(r, "showID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid show ID")
		return
	}
	profiles, err := h.repo.ListHostProfiles(showID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if profiles == nil {
		profiles = []*show.HostProfile{}
	}
	respondJSON(w, http.StatusOK, profiles)
}

func (h *ShowsHandler) CreateHostProfile(w http.ResponseWriter, r *http.Request) {
	showID, err := parseUUID(chi.URLParam(r, "showID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid show ID")
		return
	}
	var hp show.HostProfile
	if err := json.NewDecoder(r.Body).Decode(&hp); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	hp.ShowID = showID
	if err := h.repo.CreateHostProfile(&hp); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, hp)
}

func (h *ShowsHandler) GetHostProfile(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "hostID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid host ID")
		return
	}
	hp, err := h.repo.GetHostProfile(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if hp == nil {
		respondError(w, http.StatusNotFound, "host profile not found")
		return
	}
	respondJSON(w, http.StatusOK, hp)
}

// Helpers

func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func respondError(w http.ResponseWriter, status int, msg string) {
	respondJSON(w, status, map[string]string{"error": msg})
}
