package proxy

import "strings"

const HealthPath = "/__ollama_logging_proxy/health"

var tappedEndpointSet = map[string]struct{}{
	"/api/generate":   {},
	"/api/chat":       {},
	"/api/embeddings": {},
	"/api/embed":      {},
}

var skipResponseBodySet = map[string]struct{}{
	"/api/embeddings": {},
	"/api/embed":      {},
}

func NormalizePath(rawPath string) string {
	if rawPath == "" {
		return "/"
	}

	if idx := strings.IndexByte(rawPath, '?'); idx >= 0 {
		rawPath = rawPath[:idx]
	}

	if rawPath == "" {
		return "/"
	}

	return rawPath
}

func IsTappedPath(rawPath string) bool {
	_, ok := tappedEndpointSet[NormalizePath(rawPath)]
	return ok
}

func ShouldLogResponseBody(rawPath string) bool {
	_, skip := skipResponseBodySet[NormalizePath(rawPath)]
	return !skip
}
