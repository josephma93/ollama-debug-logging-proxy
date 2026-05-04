package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/joseph/ollama-logging-proxy/internal/logging"
	"github.com/joseph/ollama-logging-proxy/internal/proxy"
)

func TestHealthURL(t *testing.T) {
	testCases := []struct {
		name   string
		listen string
		want   string
	}{
		{
			name:   "wildcard host maps to localhost",
			listen: "0.0.0.0:11434",
			want:   "http://127.0.0.1:11434" + proxy.HealthPath,
		},
		{
			name:   "specific host is preserved",
			listen: "127.0.0.1:18080",
			want:   "http://127.0.0.1:18080" + proxy.HealthPath,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := healthURL(tc.listen)
			if err != nil {
				t.Fatalf("healthURL returned error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("healthURL(%q) = %q, expected %q", tc.listen, got, tc.want)
			}
		})
	}
}

func TestPrepareBodyForLog(t *testing.T) {
	input := []byte(`{"model":"llava","images":["abc123"],"prompt":"describe"}`)
	body, truncated, redacted := prepareBodyForLog(input, 1024)

	if truncated {
		t.Fatal("did not expect truncation")
	}
	if !redacted {
		t.Fatal("expected redacted=true")
	}
	if want := `{"images":"[redacted]","model":"llava","prompt":"describe"}`; body != want {
		t.Fatalf("unexpected redacted body %q, expected %q", body, want)
	}
}

func TestTailLines(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "body-2026-05-04.jsonl")

	content := "line-1\nline-2\nline-3\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	got, err := tailLines(path, 2)
	if err != nil {
		t.Fatalf("tailLines returned error: %v", err)
	}

	expected := []string{"line-2", "line-3"}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("tailLines returned %v, expected %v", got, expected)
	}
}

func TestTailLinesMissingFile(t *testing.T) {
	got, err := tailLines(filepath.Join(t.TempDir(), "missing.jsonl"), 10)
	if err != nil {
		t.Fatalf("tailLines returned error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected no lines for missing file, got %v", got)
	}
}

func TestBuildCaptureHookWritesRecord(t *testing.T) {
	logDir := t.TempDir()
	now := time.Date(2026, 5, 4, 10, 0, 0, 0, time.UTC)
	writer, err := logging.NewBodyWriterWithClock(logDir, func() time.Time { return now })
	if err != nil {
		t.Fatalf("NewBodyWriterWithClock failed: %v", err)
	}
	defer writer.Close()

	hook := buildCaptureHook(writer, 12, log.New(io.Discard, "", 0))
	hook(context.Background(), proxy.CaptureEvent{
		Metadata: proxy.CaptureMetadata{
			StartedAt: now,
			Duration:  1250 * time.Millisecond,
			ClientIP:  "127.0.0.1",
			Method:    "POST",
			Path:      "/api/chat",
			Query:     "x=1",
			UserAgent: "test-agent",
			Status:    200,
		},
		Request: proxy.BodyCapture{
			Body:      []byte(`{"images":["abc"],"prompt":"hello"}`),
			Truncated: false,
		},
		Response: proxy.BodyCapture{
			Body:      []byte(`{"response":"hello world","done":true}`),
			Truncated: true,
		},
	})

	lines, err := tailLines(filepath.Join(logDir, "body-2026-05-04.jsonl"), 1)
	if err != nil {
		t.Fatalf("tailLines failed: %v", err)
	}
	if len(lines) != 1 {
		t.Fatalf("expected 1 log line, got %d", len(lines))
	}

	var record logging.BodyLogRecord
	if err := json.Unmarshal([]byte(lines[0]), &record); err != nil {
		t.Fatalf("unmarshal log line: %v", err)
	}

	if record.Path != "/api/chat" {
		t.Fatalf("unexpected path %q", record.Path)
	}
	if !record.RequestRedacted {
		t.Fatal("expected request_redacted=true")
	}
	if !record.RequestTruncated {
		t.Fatal("expected request_truncated=true due max body bytes")
	}
	if !record.ResponseTruncated {
		t.Fatal("expected response_truncated=true")
	}
}
