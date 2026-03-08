package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	tinyserp "github.com/okayama-daiki/tiny-serp"
)

// NewHandler builds the HTTP API for tiny-serp.
func NewHandler(service *tinyserp.Service) http.Handler {
	if service == nil {
		service = tinyserp.NewService(nil)
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
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": tinyserp.ErrQueryRequired.Error()})
			return
		}
		if engine == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "query parameter engine is required"})
			return
		}

		response, err := service.Search(r.Context(), engine, query)
		if err != nil {
			writeJSON(w, statusForError(err), map[string]string{"error": err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, response)
	})

	return mux
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	body, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(append(body, '\n'))
}

func statusForError(err error) int {
	switch {
	case errors.Is(err, tinyserp.ErrQueryRequired), errors.Is(err, tinyserp.ErrUnsupportedEngine):
		return http.StatusBadRequest
	case errors.Is(err, tinyserp.ErrUpstreamBlocked), errors.Is(err, tinyserp.ErrUpstreamStatus):
		return http.StatusBadGateway
	default:
		return http.StatusBadGateway
	}
}
