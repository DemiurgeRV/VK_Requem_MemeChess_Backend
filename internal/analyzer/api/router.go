package api

import "net/http"

func NewRouter(h *Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/debug/cache", h.DebugCache)
	mux.HandleFunc("/api/v1/frontier", h.Frontier)
	mux.HandleFunc("/api/v1/health", h.Health)
	mux.HandleFunc("/api/v1/metrics", h.Metrics)
	mux.HandleFunc("/api/v1/warmup", h.Warmup)
	mux.HandleFunc("/api/v1/analyze/move", h.AnalyzeMove)

	return withCORS(mux)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		} else {
			w.Header().Add("Vary", "Origin")
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
