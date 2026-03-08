package tinyserp

import (
	"context"
	"errors"
	"io"
	"net/http"
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

	service := NewService(client)
	response, err := service.Search(context.Background(), "duckduckgo", "aws lambda")
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

	service := NewService(client)
	response, err := service.Search(context.Background(), "bing", "aws lambda")
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

func TestServiceSearchRejectsUnsupportedEngine(t *testing.T) {
	service := NewService(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		t.Fatal("unexpected outbound request")
		return nil, nil
	})})

	_, err := service.Search(context.Background(), "google", "aws lambda")
	if !errors.Is(err, ErrUnsupportedEngine) {
		t.Fatalf("expected ErrUnsupportedEngine, got %v", err)
	}
}

func TestServiceSearchRejectsEmptyQuery(t *testing.T) {
	service := NewService(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		t.Fatal("unexpected outbound request")
		return nil, nil
	})})

	_, err := service.Search(context.Background(), "duckduckgo", "")
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

	service := NewService(client)
	_, err := service.Search(context.Background(), "duckduckgo", "aws lambda")
	if !errors.Is(err, ErrUpstreamBlocked) {
		t.Fatalf("expected ErrUpstreamBlocked, got %v", err)
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
