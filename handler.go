package tinyserp

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

// NewHandler builds the HTTP API for tiny-serp.
func NewHandler(service *Service) http.Handler {
	if service == nil {
		service = NewService(nil)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method must be GET"})
			return
		}

		engine := strings.TrimSpace(r.URL.Query().Get("engine"))
		query := strings.TrimSpace(r.URL.Query().Get("q"))
		if query == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": ErrQueryRequired.Error()})
			return
		}
		if engine == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "query parameter engine is required"})
			return
		}

		response, err := service.Search(r.Context(), engine, query)
		if err != nil {
			status := http.StatusBadGateway
			switch {
			case errors.Is(err, ErrQueryRequired), errors.Is(err, ErrUnsupportedEngine):
				status = http.StatusBadRequest
			case errors.Is(err, ErrUpstreamBlocked), errors.Is(err, ErrUpstreamStatus):
				status = http.StatusBadGateway
			}
			writeJSON(w, status, map[string]string{"error": err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, response)
	})

	return mux
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
