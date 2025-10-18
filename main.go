package main

import (
	"encoding/json"
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

func (h *Handler) HandleValidateChirp(cfg *apiConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		decoder := json.NewDecoder(r.Body)
		var req struct {
			Body string `json:"body"`
		}
		err := decoder.Decode(&req)
		if err != nil {
			var body struct {
				Body string `json:"body"`
			}
			body.Body = "Something went wrong"
			data, _ := json.Marshal(body)
			w.WriteHeader(400)
			w.Write(data)
			return
		}
		if len(req.Body) > 140 {
			var body struct {
				Body string `json:"body"`
			}
			body.Body = "Chirp is too long"
			data, _ := json.Marshal(body)
			w.WriteHeader(400)
			w.Write(data)
			return
		}
		var cleanedBody struct {
			Cleaned_body string `json:"cleaned_body"`
		}
		consored := Cenzo(req.Body)
		cleanedBody.Cleaned_body = consored
		data, _ := json.Marshal(cleanedBody)
		w.WriteHeader(200)
		w.Write(data)
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
	mux.Handle("GET /admin/metrics", handler.HandleMetrics(cfg))
	mux.Handle("POST /admin/reset", handler.HandleReset(cfg))
	mux.Handle("POST /api/validate_chirp", handler.HandleValidateChirp(cfg))

	server.ListenAndServe()
}
