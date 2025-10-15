package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) FileServerHitsIncrement(n http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		n.ServeHTTP(w, r)
	})
}

type Handler struct{}

func (h *Handler) HandleFileServer(cfg *apiConfig) http.Handler {
	dir := http.Dir(".")
	res := http.StripPrefix("/app", http.FileServer(dir))
	return cfg.FileServerHitsIncrement(res)
}

func (h *Handler) HandleHealthz() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		message := "OK"
		w.Write([]byte(message))
	})
}

func (h *Handler) HandleMetrics(cfg *apiConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		num := cfg.fileserverHits.Load()

		html := fmt.Sprintf(`<html>
      <body>
        <h1>Welcome, Chirpy Admin</h1>
        <p>Chirpy has been visited %d times!</p>
      </body>
    </html>`, num)
		w.Write([]byte(html))
	})
}

func (h *Handler) HandleReset(cfg *apiConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		cfg.fileserverHits.Store(0)
	})
}

func main() {
	handler := &Handler{}
	cfg := &apiConfig{}

	mux := http.NewServeMux()

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	mux.Handle("/app/", handler.HandleFileServer(cfg))
	mux.Handle("GET /api/healthz", handler.HandleHealthz())
	mux.Handle("GET /api/metrics", handler.HandleMetrics(cfg))
	mux.Handle("POST /api/reset", handler.HandleReset(cfg))

	server.ListenAndServe()
}
