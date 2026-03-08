package tinyserp

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Engine describes a search engine implementation.
type Engine interface {
	Name() string
	BuildRequest(ctx context.Context, query string) (*http.Request, error)
	Parse(io.Reader) ([]SearchItem, error)
}

// DuckDuckGoEngine searches DuckDuckGo's HTML endpoint.
type DuckDuckGoEngine struct{}

// BingEngine searches Bing's result page.
type BingEngine struct{}

// DefaultEngines returns the built-in engine registry keyed by engine name.
func DefaultEngines() map[string]Engine {
	return map[string]Engine{
		"duckduckgo": DuckDuckGoEngine{},
		"bing":       BingEngine{},
	}
}

func (DuckDuckGoEngine) Name() string {
	return "duckduckgo"
}

func (DuckDuckGoEngine) BuildRequest(ctx context.Context, query string) (*http.Request, error) {
	return buildSearchRequest(ctx, "https://html.duckduckgo.com/html/", query)
}

func (DuckDuckGoEngine) Parse(r io.Reader) ([]SearchItem, error) {
	return parseDuckDuckGo(r)
}

func (BingEngine) Name() string {
	return "bing"
}

func (BingEngine) BuildRequest(ctx context.Context, query string) (*http.Request, error) {
	return buildSearchRequest(ctx, "https://www.bing.com/search", query)
}

func (BingEngine) Parse(r io.Reader) ([]SearchItem, error) {
	return parseBing(r)
}

func buildSearchRequest(ctx context.Context, endpoint, query string) (*http.Request, error) {
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("parse endpoint: %w", err)
	}

	values := parsed.Query()
	values.Set("q", query)
	parsed.RawQuery = values.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	return req, nil
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
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	if parsed.Scheme == "" && parsed.Host != "" {
		parsed.Scheme = "https"
	}

	if hasHostSuffix(parsed.Hostname(), "duckduckgo.com") {
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
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}

	if hasHostSuffix(parsed.Hostname(), "bing.com") {
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

func hasHostSuffix(host, suffix string) bool {
	return host == suffix || strings.HasSuffix(host, "."+suffix)
}

func isHTTPURL(value string) bool {
	return strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://")
}

func normalizeSpace(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}
