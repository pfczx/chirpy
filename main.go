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
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		type request struct {
			Body string `json:"body"`
		}
		type response struct {
			Body bool `json:"valid"`
		}
		type errorJson struct {
			Body string `json:"error"`
		}
		decoder := json.NewDecoder(r.Body)
		req := request{}
		err := decoder.Decode(&req)
		if err != nil {
			body := errorJson{
				Body: "Something went wrong",
			}
			data, _ := json.Marshal(body)
			w.WriteHeader(400)			
			w.Write(data)
			return
		}
		if len(req.Body) > 140 {
			body := errorJson{
				Body: "Chirp is too long",
			}
			data, _ := json.Marshal(body)
			w.WriteHeader(400)
			w.Write(data)
			return
		}
		resp := response{
			Body: true,
		}
		data, _ := json.Marshal(resp)
		w.Write(data)
		w.WriteHeader(200)
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
	mux.Handle("POST /api/validate_chirp",handler.HandleValidateChirp(cfg))

	server.ListenAndServe()
}
