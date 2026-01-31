package api

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
)

// SetupRoutes configures all HTTP routes
func SetupRoutes(h *Handler, staticFS embed.FS) *http.ServeMux {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", h.HealthCheck)

	// API routes
	mux.HandleFunc("GET /api/stats", h.GetDashboardStats)
	mux.HandleFunc("GET /api/search", h.Search)
	mux.HandleFunc("GET /api/jobs/{id}", h.GetJob)
	mux.HandleFunc("GET /api/jobs/pending/{queue}/paginated", h.GetPendingJobsPaginated)
	mux.HandleFunc("GET /api/jobs/pending/{queue}/count", h.GetPendingCount)
	mux.HandleFunc("GET /api/jobs/pending/{queue}", h.GetPendingJobs)
	mux.HandleFunc("GET /api/jobs", h.GetJobsByStatus)
	mux.HandleFunc("DELETE /api/jobs/{id}", h.DeleteJob)
	mux.HandleFunc("GET /api/chains/{id}", h.GetChain)
	mux.HandleFunc("GET /api/chains", h.ListChains)
	mux.HandleFunc("GET /api/groups/{id}", h.GetGroup)
	mux.HandleFunc("GET /api/groups", h.ListGroups)

	// Serve static files and index.html
	staticSub, err := fs.Sub(staticFS, "web")
	if err != nil {
		log.Fatal(err)
	}

	fileServer := http.FileServer(http.FS(staticSub))
	mux.Handle("/", fileServer)

	return mux
}

// CORSMiddleware adds CORS headers
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
