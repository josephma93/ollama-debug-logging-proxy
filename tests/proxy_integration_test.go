package tests

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/joseph/ollama-logging-proxy/internal/proxy"
)

type upstreamRequest struct {
	Method string
	Path   string
	Query  string
	Body   string
}

func TestProxyForwardsNonTappedRequests(t *testing.T) {
	requests := make(chan upstreamRequest, 1)
	var upstreamCalls atomic.Int32
	hookCalls := make(chan proxy.CaptureEvent, 1)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamCalls.Add(1)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed reading request body: %v", err)
		}

		requests <- upstreamRequest{
			Method: r.Method,
			Path:   r.URL.Path,
			Query:  r.URL.RawQuery,
			Body:   string(body),
		}

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("upstream-ok"))
	}))
	defer upstream.Close()

	proxyServer := newProxyServer(t, upstream.URL, 1024, func(_ context.Context, event proxy.CaptureEvent) {
		hookCalls <- event
	})
	defer proxyServer.Close()

	req, err := http.NewRequest(http.MethodPost, proxyServer.URL+"/api/tags?source=test", strings.NewReader(`{"hello":"world"}`))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("proxy request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed reading proxy response body: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}
	if string(body) != "upstream-ok" {
		t.Fatalf("expected upstream response body, got %q", string(body))
	}

	select {
	case got := <-requests:
		if got.Method != http.MethodPost {
			t.Fatalf("expected method %s, got %s", http.MethodPost, got.Method)
		}
		if got.Path != "/api/tags" {
			t.Fatalf("expected path /api/tags, got %s", got.Path)
		}
		if got.Query != "source=test" {
			t.Fatalf("expected query source=test, got %s", got.Query)
		}
		if got.Body != `{"hello":"world"}` {
			t.Fatalf("expected forwarded body, got %q", got.Body)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for forwarded upstream request")
	}

	select {
	case <-hookCalls:
		t.Fatal("did not expect capture hook for non-tapped endpoint")
	case <-time.After(200 * time.Millisecond):
	}

	if upstreamCalls.Load() != 1 {
		t.Fatalf("expected 1 upstream call, got %d", upstreamCalls.Load())
	}
}

func TestProxyHealthEndpointIsNotForwarded(t *testing.T) {
	var upstreamCalls atomic.Int32

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		upstreamCalls.Add(1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("unexpected"))
	}))
	defer upstream.Close()

	proxyServer := newProxyServer(t, upstream.URL, 1024, nil)
	defer proxyServer.Close()

	resp, err := http.Get(proxyServer.URL + proxy.HealthPath)
	if err != nil {
		t.Fatalf("health request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("health response was not valid JSON: %v", err)
	}
	if ok, _ := payload["ok"].(bool); !ok {
		t.Fatalf("expected ok=true, got %#v", payload["ok"])
	}
	if service, _ := payload["service"].(string); service != "ollama-logging-proxy" {
		t.Fatalf("expected service=ollama-logging-proxy, got %#v", payload["service"])
	}
	if upstreamCalls.Load() != 0 {
		t.Fatalf("health endpoint should not be forwarded, got %d upstream calls", upstreamCalls.Load())
	}
}

func TestProxyTappedCaptureWithQueryAndBounds(t *testing.T) {
	const (
		requestBody  = "request-body-long"
		responseBody = "response-body-long"
	)

	events := make(chan proxy.CaptureEvent, 1)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(responseBody))
	}))
	defer upstream.Close()

	proxyServer := newProxyServer(t, upstream.URL, 8, func(_ context.Context, event proxy.CaptureEvent) {
		events <- event
	})
	defer proxyServer.Close()

	resp, err := http.Post(proxyServer.URL+"/api/chat?trace=1", "application/json", strings.NewReader(requestBody))
	if err != nil {
		t.Fatalf("tapped request failed: %v", err)
	}
	defer resp.Body.Close()

	fullResponse, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed reading response body: %v", err)
	}
	if string(fullResponse) != responseBody {
		t.Fatalf("expected full response body %q, got %q", responseBody, string(fullResponse))
	}

	select {
	case event := <-events:
		if !event.Metadata.Tapped {
			t.Fatal("expected tapped metadata to be true")
		}
		if event.Metadata.Path != "/api/chat" {
			t.Fatalf("expected metadata path /api/chat, got %q", event.Metadata.Path)
		}
		if event.Metadata.Query != "trace=1" {
			t.Fatalf("expected metadata query trace=1, got %q", event.Metadata.Query)
		}
		if event.Metadata.Status != http.StatusOK {
			t.Fatalf("expected metadata status %d, got %d", http.StatusOK, event.Metadata.Status)
		}

		if !bytes.Equal(event.Request.Body, []byte("request-")) {
			t.Fatalf("expected truncated request capture %q, got %q", "request-", string(event.Request.Body))
		}
		if !event.Request.Truncated {
			t.Fatal("expected request capture to be truncated")
		}

		if !bytes.Equal(event.Response.Body, []byte("response")) {
			t.Fatalf("expected truncated response capture %q, got %q", "response", string(event.Response.Body))
		}
		if !event.Response.Truncated {
			t.Fatal("expected response capture to be truncated")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for capture event")
	}
}

func TestProxyStreamsIncrementally(t *testing.T) {
	const streamDelay = 250 * time.Millisecond

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("upstream response writer does not implement http.Flusher")
		}

		w.Header().Set("Content-Type", "application/x-ndjson")
		_, _ = w.Write([]byte("chunk-1\n"))
		flusher.Flush()

		time.Sleep(streamDelay)

		_, _ = w.Write([]byte("chunk-2\n"))
		flusher.Flush()
	}))
	defer upstream.Close()

	proxyServer := newProxyServer(t, upstream.URL, 1024, nil)
	defer proxyServer.Close()

	startedAt := time.Now()
	resp, err := http.Get(proxyServer.URL + "/api/generate")
	if err != nil {
		t.Fatalf("streaming request failed: %v", err)
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)

	firstChunk, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("failed reading first chunk: %v", err)
	}
	firstChunkLatency := time.Since(startedAt)

	secondChunk, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("failed reading second chunk: %v", err)
	}
	totalLatency := time.Since(startedAt)

	if firstChunk != "chunk-1\n" {
		t.Fatalf("unexpected first chunk %q", firstChunk)
	}
	if secondChunk != "chunk-2\n" {
		t.Fatalf("unexpected second chunk %q", secondChunk)
	}
	if firstChunkLatency >= 180*time.Millisecond {
		t.Fatalf("expected first chunk before buffering delay; got latency %s", firstChunkLatency)
	}
	if totalLatency < streamDelay {
		t.Fatalf("expected total latency >= %s, got %s", streamDelay, totalLatency)
	}
}

func newProxyServer(t *testing.T, upstreamURL string, maxBodyBytes int64, hook proxy.CaptureHook) *httptest.Server {
	t.Helper()

	target, err := url.Parse(upstreamURL)
	if err != nil {
		t.Fatalf("failed to parse upstream URL: %v", err)
	}

	handler, err := proxy.New(proxy.Options{
		Target:       target,
		MaxBodyBytes: maxBodyBytes,
		OnCapture:    hook,
	})
	if err != nil {
		t.Fatalf("failed to create proxy handler: %v", err)
	}

	return httptest.NewServer(handler)
}
