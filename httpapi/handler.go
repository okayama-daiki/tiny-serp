package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	tinyserp "github.com/okayama-daiki/tiny-serp"
)

// NewHandler builds the HTTP API for tiny-serp.
func NewHandler(client *http.Client, engines map[string]tinyserp.Engine) http.Handler {
	registry := tinyserp.NewEngineRegistry(engines)

	mux := http.NewServeMux()
	mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method must be GET"})
			return
		}

		engineName := strings.TrimSpace(r.URL.Query().Get("engine"))
		query := strings.TrimSpace(r.URL.Query().Get("q"))
		if query == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": tinyserp.ErrQueryRequired.Error()})
			return
		}
		if engineName == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "query parameter engine is required"})
			return
		}

		engine, ok := registry.Resolve(engineName)
		if !ok {
			err := fmt.Errorf("%w: %s", tinyserp.ErrUnsupportedEngine, engineName)
			status := statusForError(err)
			writeJSON(w, status, errorPayload(status, err))
			return
		}

		response, err := tinyserp.NewService(engine, client).Search(r.Context(), query)
		if err != nil {
			status := statusForError(err)
			writeJSON(w, status, errorPayload(status, err))
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

func errorPayload(status int, err error) map[string]string {
	message := err.Error()
	if status == http.StatusInternalServerError {
		message = "internal server error"
	}

	return map[string]string{"error": message}
}

func statusForError(err error) int {
	switch {
	case errors.Is(err, tinyserp.ErrQueryRequired), errors.Is(err, tinyserp.ErrUnsupportedEngine):
		return http.StatusBadRequest
	case errors.Is(err, tinyserp.ErrUpstreamBlocked), errors.Is(err, tinyserp.ErrUpstreamStatus):
		return http.StatusBadGateway
	default:
		return http.StatusInternalServerError
	}
}
