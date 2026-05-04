package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	EnvListen       = "OLLAMA_PROXY_LISTEN"
	EnvTarget       = "OLLAMA_PROXY_TARGET"
	EnvMaxBodyBytes = "OLLAMA_PROXY_MAX_BODY_BYTES"

	DefaultListen       = "0.0.0.0:11434"
	DefaultTarget       = "http://127.0.0.1:11435"
	DefaultMaxBodyBytes = int64(10 * 1024 * 1024)
)

type Config struct {
	Listen       string
	Target       *url.URL
	MaxBodyBytes int64
}

func Load() (Config, error) {
	listen := getEnvOrDefault(EnvListen, DefaultListen)
	targetRaw := getEnvOrDefault(EnvTarget, DefaultTarget)
	target, err := parseTarget(targetRaw)
	if err != nil {
		return Config{}, err
	}

	maxBodyBytes, err := parseMaxBodyBytes(getEnvOrDefault(EnvMaxBodyBytes, ""))
	if err != nil {
		return Config{}, err
	}

	return Config{
		Listen:       listen,
		Target:       target,
		MaxBodyBytes: maxBodyBytes,
	}, nil
}

func parseTarget(raw string) (*url.URL, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return nil, fmt.Errorf("%s is invalid: %w", EnvTarget, err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("%s must include scheme and host", EnvTarget)
	}
	return parsed, nil
}

func parseMaxBodyBytes(raw string) (int64, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return DefaultMaxBodyBytes, nil
	}

	value, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s is invalid: %w", EnvMaxBodyBytes, err)
	}
	if value <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", EnvMaxBodyBytes)
	}
	return value, nil
}

func getEnvOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
