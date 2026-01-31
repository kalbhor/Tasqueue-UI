package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/kalbhor/tasqueue-ui/internal/service"
)

// Handler holds the service instance and provides HTTP handlers
type Handler struct {
	service *service.Service
}

// NewHandler creates a new API handler
func NewHandler(svc *service.Service) *Handler {
	return &Handler{
		service: svc,
	}
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError sends an error response
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, ErrorResponse{Error: message})
}

// GetDashboardStats handles GET /api/stats
func (h *Handler) GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetDashboardStats(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, stats)
}

// GetJob handles GET /api/jobs/:id
func (h *Handler) GetJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "job ID is required")
		return
	}

	job, err := h.service.GetJob(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, job)
}

// GetPendingJobs handles GET /api/jobs/pending/:queue
// Deprecated: Use GetPendingJobsPaginated for better performance
func (h *Handler) GetPendingJobs(w http.ResponseWriter, r *http.Request) {
	queue := r.PathValue("queue")

	jobs, err := h.service.GetPendingJobs(r.Context(), queue)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, jobs)
}

// GetPendingJobsPaginated handles GET /api/jobs/pending/:queue/paginated?offset=0&limit=20
func (h *Handler) GetPendingJobsPaginated(w http.ResponseWriter, r *http.Request) {
	queue := r.PathValue("queue")

	// Parse query parameters
	offset := 0
	limit := 20

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if n, err := strconv.Atoi(offsetStr); err == nil && n >= 0 {
			offset = n
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if n, err := strconv.Atoi(limitStr); err == nil && n > 0 {
			limit = n
		}
	}

	result, err := h.service.GetPendingJobsWithPagination(r.Context(), queue, offset, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// GetPendingCount handles GET /api/jobs/pending/:queue/count
func (h *Handler) GetPendingCount(w http.ResponseWriter, r *http.Request) {
	queue := r.PathValue("queue")

	count, err := h.service.GetPendingCount(r.Context(), queue)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"queue": queue,
		"count": count,
	})
}

// GetJobsByStatus handles GET /api/jobs?status=<status>
func (h *Handler) GetJobsByStatus(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status == "" {
		respondError(w, http.StatusBadRequest, "status parameter is required")
		return
	}

	jobIDs, err := h.service.GetJobsByStatus(r.Context(), status)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":  status,
		"job_ids": jobIDs,
		"count":   len(jobIDs),
	})
}

// DeleteJob handles DELETE /api/jobs/:id
func (h *Handler) DeleteJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "job ID is required")
		return
	}

	err := h.service.DeleteJob(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "job deleted successfully"})
}

// GetChain handles GET /api/chains/:id
func (h *Handler) GetChain(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "chain ID is required")
		return
	}

	chain, err := h.service.GetChain(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, chain)
}

// ListChains handles GET /api/chains
func (h *Handler) ListChains(w http.ResponseWriter, r *http.Request) {
	chains, err := h.service.ListChains(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"chains": chains,
		"count":  len(chains),
	})
}

// GetGroup handles GET /api/groups/:id
func (h *Handler) GetGroup(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "group ID is required")
		return
	}

	group, err := h.service.GetGroup(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, group)
}

// ListGroups handles GET /api/groups
func (h *Handler) ListGroups(w http.ResponseWriter, r *http.Request) {
	groups, err := h.service.ListGroups(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"groups": groups,
		"count":  len(groups),
	})
}

// Search handles GET /api/search?q=<query>
func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		respondError(w, http.StatusBadRequest, "query parameter 'q' is required")
		return
	}

	result, err := h.service.Search(r.Context(), query)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// HealthCheck handles GET /health
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}
