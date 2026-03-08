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
	if engines == nil {
		engines = tinyserp.DefaultEngines()
	}
	engines = normalizeEngines(engines)

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

		engine, ok := engines[normalizeEngineName(engineName)]
		if !ok {
			err := fmt.Errorf("%w: %s", tinyserp.ErrUnsupportedEngine, engineName)
			writeJSON(w, statusForError(err), map[string]string{"error": err.Error()})
			return
		}

		response, err := tinyserp.NewService(engine, client).Search(r.Context(), query)
		if err != nil {
			writeJSON(w, statusForError(err), map[string]string{"error": err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, response)
	})

	return mux
}

func normalizeEngines(engines map[string]tinyserp.Engine) map[string]tinyserp.Engine {
	normalized := make(map[string]tinyserp.Engine, len(engines))
	for name, engine := range engines {
		if engine == nil {
			continue
		}

		key := normalizeEngineName(name)
		if key == "" {
			continue
		}

		normalized[key] = engine
	}

	return normalized
}

func normalizeEngineName(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
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
		return http.StatusInternalServerError
	}
}
