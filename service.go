package tinyserp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	ErrUnsupportedEngine = errors.New("unsupported engine")
	ErrEngineRequired    = errors.New("engine is required")
	ErrQueryRequired     = errors.New("query parameter q is required")
	ErrUpstreamBlocked   = errors.New("upstream blocked the request")
	ErrUpstreamStatus    = errors.New("unexpected upstream status")
)

var defaultHTTPClient = &http.Client{Timeout: 10 * time.Second}

// Service executes upstream searches and parses the returned HTML.
type Service struct {
	engine    Engine
	client    *http.Client
	userAgent string
	language  string
}

// NewService creates a search service bound to a single engine.
func NewService(engine Engine, client *http.Client) *Service {
	if client == nil {
		client = defaultHTTPClient
	}

	userAgent := strings.TrimSpace(os.Getenv("TINY_SERP_USER_AGENT"))
	if userAgent == "" {
		userAgent = "tiny-serp/0.1 (+https://github.com/okayama-daiki/tiny-serp)"
	}

	return &Service{
		engine:    engine,
		client:    client,
		userAgent: userAgent,
		language:  "en-US,en;q=0.9",
	}
}

// Search executes a query against the configured engine.
func (s *Service) Search(ctx context.Context, query string) (SearchResponse, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return SearchResponse{}, ErrQueryRequired
	}

	if s.engine == nil {
		return SearchResponse{}, ErrEngineRequired
	}

	req, err := s.engine.BuildRequest(ctx, query)
	if err != nil {
		return SearchResponse{}, fmt.Errorf("build request for %s: %w", s.engine.Name(), err)
	}
	if req.Header == nil {
		req.Header = make(http.Header)
	}
	if s.userAgent != "" && req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", s.userAgent)
	}
	if s.language != "" && req.Header.Get("Accept-Language") == "" {
		req.Header.Set("Accept-Language", s.language)
	}
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "text/html,application/xhtml+xml")
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return SearchResponse{}, fmt.Errorf("request %s: %w", s.engine.Name(), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return SearchResponse{}, fmt.Errorf("%w: %s returned status %d", ErrUpstreamStatus, s.engine.Name(), resp.StatusCode)
	}

	items, err := s.engine.Parse(resp.Body)
	if err != nil {
		return SearchResponse{}, err
	}

	for i := range items {
		items[i].Rank = i + 1
	}

	return SearchResponse{
		SearchInformation: SearchInformation{
			Query:           query,
			Engine:          s.engine.Name(),
			ResultsReturned: len(items),
		},
		Items: items,
	}, nil
}
