package tinyserp

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
)

func TestServiceSearchDuckDuckGo(t *testing.T) {
	html := readFixture(t, "testdata/duckduckgo.html")
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", req.Method)
		}
		if req.URL.Host != "html.duckduckgo.com" {
			t.Fatalf("unexpected host: %s", req.URL.Host)
		}
		if got := req.URL.Query().Get("q"); got != "aws lambda" {
			t.Fatalf("unexpected query: %s", got)
		}
		if req.Header.Get("User-Agent") == "" {
			t.Fatal("user-agent header was not set")
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(html)),
		}, nil
	})}

	service := NewService(DuckDuckGoEngine{}, client)
	response, err := service.Search(context.Background(), "aws lambda")
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}

	if response.SearchInformation.Query != "aws lambda" {
		t.Fatalf("unexpected query: %s", response.SearchInformation.Query)
	}
	if response.SearchInformation.Engine != "duckduckgo" {
		t.Fatalf("unexpected engine: %s", response.SearchInformation.Engine)
	}
	if response.SearchInformation.ResultsReturned != 2 {
		t.Fatalf("unexpected resultsReturned: %d", response.SearchInformation.ResultsReturned)
	}
	if len(response.Items) != 2 {
		t.Fatalf("unexpected number of items: %d", len(response.Items))
	}

	first := response.Items[0]
	if first.Rank != 1 {
		t.Fatalf("unexpected rank: %d", first.Rank)
	}
	if first.Title != "AWS Lambda - Amazon Web Services" {
		t.Fatalf("unexpected title: %s", first.Title)
	}
	if first.Link != "https://aws.amazon.com/lambda/" {
		t.Fatalf("unexpected link: %s", first.Link)
	}
	if first.Snippet != "Run code without thinking about servers." {
		t.Fatalf("unexpected snippet: %s", first.Snippet)
	}
}

func TestServiceSearchBing(t *testing.T) {
	html := readFixture(t, "testdata/bing.html")
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host != "www.bing.com" {
			t.Fatalf("unexpected host: %s", req.URL.Host)
		}
		if got := req.URL.Query().Get("q"); got != "aws lambda" {
			t.Fatalf("unexpected query: %s", got)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(html)),
		}, nil
	})}

	service := NewService(BingEngine{}, client)
	response, err := service.Search(context.Background(), "aws lambda")
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}

	if response.SearchInformation.ResultsReturned != 2 {
		t.Fatalf("unexpected resultsReturned: %d", response.SearchInformation.ResultsReturned)
	}
	if len(response.Items) != 2 {
		t.Fatalf("unexpected number of items: %d", len(response.Items))
	}
	if response.Items[0].Link != "https://aws.amazon.com/lambda/" {
		t.Fatalf("unexpected decoded link: %s", response.Items[0].Link)
	}
	if response.Items[0].Snippet != "Run code without provisioning or managing servers." {
		t.Fatalf("unexpected snippet: %s", response.Items[0].Snippet)
	}
}

func TestServiceSearchRejectsEmptyQuery(t *testing.T) {
	service := NewService(DuckDuckGoEngine{}, &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		t.Fatal("unexpected outbound request")
		return nil, nil
	})})

	_, err := service.Search(context.Background(), "")
	if !errors.Is(err, ErrQueryRequired) {
		t.Fatalf("expected ErrQueryRequired, got %v", err)
	}
}

func TestServiceSearchDetectsDuckDuckGoChallenge(t *testing.T) {
	html := readFixture(t, "testdata/duckduckgo_challenge.html")
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(html)),
		}, nil
	})}

	service := NewService(DuckDuckGoEngine{}, client)
	_, err := service.Search(context.Background(), "aws lambda")
	if !errors.Is(err, ErrUpstreamBlocked) {
		t.Fatalf("expected ErrUpstreamBlocked, got %v", err)
	}
}

func TestServiceSearchRejectsNilEngine(t *testing.T) {
	service := NewService(nil, &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		t.Fatal("unexpected outbound request")
		return nil, nil
	})})

	_, err := service.Search(context.Background(), "aws lambda")
	if !errors.Is(err, ErrEngineRequired) {
		t.Fatalf("expected ErrEngineRequired, got %v", err)
	}
}

func TestServiceSearchSupportsCustomEngine(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host != "example.com" {
			t.Fatalf("unexpected host: %s", req.URL.Host)
		}
		if got := req.URL.Query().Get("q"); got != "aws lambda" {
			t.Fatalf("unexpected query: %s", got)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("<html></html>")),
		}, nil
	})}

	service := NewService(customEngine{
		name: "custom",
		items: []SearchItem{{
			Title:   "Custom result",
			Link:    "https://example.com/result",
			Snippet: "Custom snippet",
		}},
	}, client)

	response, err := service.Search(context.Background(), "aws lambda")
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if response.SearchInformation.Engine != "custom" {
		t.Fatalf("unexpected engine: %s", response.SearchInformation.Engine)
	}
	if len(response.Items) != 1 {
		t.Fatalf("unexpected items length: %d", len(response.Items))
	}
	if response.Items[0].Rank != 1 {
		t.Fatalf("unexpected rank: %d", response.Items[0].Rank)
	}
}

func TestDefaultEnginesReturnsCopy(t *testing.T) {
	first := DefaultEngines()
	if _, ok := first["bing"]; !ok {
		t.Fatal("expected bing engine to be registered")
	}
	if _, ok := first["duckduckgo"]; !ok {
		t.Fatal("expected duckduckgo engine to be registered")
	}

	first["bing"] = nil

	second := DefaultEngines()
	if second["bing"] == nil {
		t.Fatal("expected default engines to be isolated from caller mutations")
	}
}

func TestNormalizeDuckDuckGoLinkRejectsLookalikeHost(t *testing.T) {
	link := normalizeDuckDuckGoLink("https://notduckduckgo.com/l/?uddg=https%3A%2F%2Faws.amazon.com%2Flambda%2F")
	if link != "https://notduckduckgo.com/l/?uddg=https%3A%2F%2Faws.amazon.com%2Flambda%2F" {
		t.Fatalf("unexpected normalized link: %s", link)
	}
}

func TestNormalizeDuckDuckGoLinkHandlesProtocolRelativeURLs(t *testing.T) {
	link := normalizeDuckDuckGoLink("//duckduckgo.com/l/?uddg=https%3A%2F%2Faws.amazon.com%2Flambda%2F")
	if link != "https://aws.amazon.com/lambda/" {
		t.Fatalf("unexpected normalized link: %s", link)
	}
}

func readFixture(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", path, err)
	}

	return string(content)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type customEngine struct {
	name  string
	items []SearchItem
}

func (e customEngine) Name() string {
	return e.name
}

func (e customEngine) BuildRequest(ctx context.Context, query string) (*http.Request, error) {
	target := &url.URL{
		Scheme: "https",
		Host:   "example.com",
		Path:   "/search",
	}
	values := target.Query()
	values.Set("q", query)
	target.RawQuery = values.Encode()

	return http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
}

func (e customEngine) Parse(r io.Reader) ([]SearchItem, error) {
	return append([]SearchItem(nil), e.items...), nil
}
