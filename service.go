package tinyserp

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	ErrUnsupportedEngine = errors.New("unsupported engine")
	ErrQueryRequired     = errors.New("query parameter q is required")
	ErrUpstreamBlocked   = errors.New("upstream blocked the request")
	ErrUpstreamStatus    = errors.New("unexpected upstream status")
)

type engineConfig struct {
	endpoint string
	parse    func(io.Reader) ([]SearchItem, error)
}

var engines = map[string]engineConfig{
	"duckduckgo": {
		endpoint: "https://html.duckduckgo.com/html/",
		parse:    parseDuckDuckGo,
	},
	"bing": {
		endpoint: "https://www.bing.com/search",
		parse:    parseBing,
	},
}

// Service executes upstream searches and parses the returned HTML.
type Service struct {
	client    *http.Client
	userAgent string
	language  string
}

// NewService creates a search service with a default timeout when client is nil.
func NewService(client *http.Client) *Service {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	userAgent := strings.TrimSpace(os.Getenv("TINY_SERP_USER_AGENT"))
	if userAgent == "" {
		userAgent = "tiny-serp/0.1 (+https://github.com/okayama-daiki/tiny-serp)"
	}

	return &Service{
		client:    client,
		userAgent: userAgent,
		language:  "en-US,en;q=0.9",
	}
}

// Search executes a query against the selected engine.
func (s *Service) Search(ctx context.Context, engineName, query string) (SearchResponse, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return SearchResponse{}, ErrQueryRequired
	}

	trimmedEngineName := strings.TrimSpace(engineName)
	engineName = normalizeEngineName(engineName)
	config, ok := engines[engineName]
	if !ok {
		return SearchResponse{}, fmt.Errorf("%w: %s", ErrUnsupportedEngine, trimmedEngineName)
	}

	endpoint, err := url.Parse(config.endpoint)
	if err != nil {
		return SearchResponse{}, fmt.Errorf("parse endpoint for %s: %w", engineName, err)
	}

	values := endpoint.Query()
	values.Set("q", query)
	endpoint.RawQuery = values.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return SearchResponse{}, fmt.Errorf("create request: %w", err)
	}
	if s.userAgent != "" {
		req.Header.Set("User-Agent", s.userAgent)
	}
	if s.language != "" {
		req.Header.Set("Accept-Language", s.language)
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := s.client.Do(req)
	if err != nil {
		return SearchResponse{}, fmt.Errorf("request %s: %w", engineName, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return SearchResponse{}, fmt.Errorf("%w: %s returned status %d", ErrUpstreamStatus, engineName, resp.StatusCode)
	}

	items, err := config.parse(resp.Body)
	if err != nil {
		return SearchResponse{}, err
	}

	for i := range items {
		items[i].Rank = i + 1
	}

	return SearchResponse{
		SearchInformation: SearchInformation{
			Query:           query,
			Engine:          engineName,
			ResultsReturned: len(items),
		},
		Items: items,
	}, nil
}

func parseDuckDuckGo(r io.Reader) ([]SearchItem, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("parse duckduckgo document: %w", err)
	}

	if doc.Find(".anomaly-modal__modal").Length() > 0 {
		return nil, ErrUpstreamBlocked
	}

	items := make([]SearchItem, 0, 10)
	doc.Find(".result").Each(func(_ int, selection *goquery.Selection) {
		anchor := selection.Find(".result__title .result__a").First()
		if anchor.Length() == 0 {
			anchor = selection.Find(".result__a").First()
		}

		title := normalizeSpace(anchor.Text())
		href, _ := anchor.Attr("href")
		link := normalizeDuckDuckGoLink(href)
		snippet := normalizeSpace(selection.Find(".result__snippet").First().Text())
		if title == "" || link == "" {
			return
		}

		items = append(items, SearchItem{
			Title:   title,
			Link:    link,
			Snippet: snippet,
		})
	})

	return items, nil
}

func parseBing(r io.Reader) ([]SearchItem, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("parse bing document: %w", err)
	}

	items := make([]SearchItem, 0, 10)
	doc.Find("#b_results .b_algo").Each(func(_ int, selection *goquery.Selection) {
		anchor := selection.Find("h2 a").First()
		title := normalizeSpace(anchor.Text())
		href, _ := anchor.Attr("href")
		link := normalizeBingLink(href)
		snippet := normalizeSpace(selection.Find(".b_caption p").First().Text())
		if title == "" || link == "" {
			return
		}

		items = append(items, SearchItem{
			Title:   title,
			Link:    link,
			Snippet: snippet,
		})
	})

	return items, nil
}

func normalizeDuckDuckGoLink(raw string) string {
	if raw == "" {
		return ""
	}
	raw = withHTTPSchemeIfProtocolRelative(raw)

	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}

	if strings.Contains(parsed.Host, "duckduckgo.com") {
		target := parsed.Query().Get("uddg")
		if target != "" {
			decoded, err := url.QueryUnescape(target)
			if err == nil && isHTTPURL(decoded) {
				return decoded
			}
		}
	}

	if parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}

	return parsed.String()
}

func normalizeBingLink(raw string) string {
	if raw == "" {
		return ""
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}

	if strings.Contains(parsed.Host, "bing.com") {
		if target := decodeBingTarget(parsed.Query().Get("u")); target != "" {
			return target
		}
	}

	if parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}

	return parsed.String()
}

func decodeBingTarget(encoded string) string {
	encoded = strings.TrimSpace(encoded)
	if encoded == "" {
		return ""
	}
	encoded = strings.TrimPrefix(encoded, "a1")

	decoded, err := base64.RawStdEncoding.DecodeString(encoded)
	if err != nil {
		decoded, err = base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return ""
		}
	}

	target := string(decoded)
	if isHTTPURL(target) {
		return target
	}

	return ""
}

func normalizeEngineName(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func withHTTPSchemeIfProtocolRelative(raw string) string {
	trimmed := strings.TrimPrefix(raw, "//")
	if trimmed == raw {
		return raw
	}

	return "https://" + trimmed
}

func isHTTPURL(value string) bool {
	return strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://")
}

func normalizeSpace(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}
