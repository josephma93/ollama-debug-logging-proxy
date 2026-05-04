package tests

import (
	"testing"

	"github.com/joseph/ollama-logging-proxy/internal/proxy"
)

func TestIsTappedPath(t *testing.T) {
	testCases := []struct {
		path   string
		tapped bool
	}{
		{path: "/api/generate", tapped: true},
		{path: "/api/chat", tapped: true},
		{path: "/api/embeddings", tapped: true},
		{path: "/api/embed", tapped: true},
		{path: "/api/chat?trace=1", tapped: true},
		{path: "/api/chat/", tapped: false},
		{path: "/api/tags", tapped: false},
		{path: "/", tapped: false},
	}

	for _, tc := range testCases {
		got := proxy.IsTappedPath(tc.path)
		if got != tc.tapped {
			t.Fatalf("IsTappedPath(%q) = %v, expected %v", tc.path, got, tc.tapped)
		}
	}
}

func TestNormalizePath(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{input: "", expected: "/"},
		{input: "/api/chat?x=1", expected: "/api/chat"},
		{input: "/api/generate", expected: "/api/generate"},
		{input: "?", expected: "/"},
	}

	for _, tc := range testCases {
		got := proxy.NormalizePath(tc.input)
		if got != tc.expected {
			t.Fatalf("NormalizePath(%q) = %q, expected %q", tc.input, got, tc.expected)
		}
	}
}
