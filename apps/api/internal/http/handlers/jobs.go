package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/ismobaga/apgn/internal/domain/job"
)

// JobsHandler handles job run endpoints.
type JobsHandler struct {
	repo job.Repository
}

func NewJobsHandler(repo job.Repository) *JobsHandler {
	return &JobsHandler{repo: repo}
}

func (h *JobsHandler) List(w http.ResponseWriter, r *http.Request) {
	episodeID, err := parseUUID(chi.URLParam(r, "episodeID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid episode ID")
		return
	}
	jobs, err := h.repo.ListJobRuns(episodeID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if jobs == nil {
		jobs = []*job.JobRun{}
	}
	respondJSON(w, http.StatusOK, jobs)
}

func (h *JobsHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "jobID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid job ID")
		return
	}
	j, err := h.repo.GetJobRun(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if j == nil {
		respondError(w, http.StatusNotFound, "job not found")
		return
	}
	respondJSON(w, http.StatusOK, j)
}
