package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	EnvListen        = "OLLAMA_PROXY_LISTEN"
	EnvTarget        = "OLLAMA_PROXY_TARGET"
	EnvLogDir        = "OLLAMA_PROXY_LOG_DIR"
	EnvRetentionDays = "OLLAMA_PROXY_RETENTION_DAYS"
	EnvMaxBodyBytes  = "OLLAMA_PROXY_MAX_BODY_BYTES"

	DefaultListen        = "0.0.0.0:11434"
	DefaultTarget        = "http://127.0.0.1:11435"
	DefaultRetentionDays = 10
	DefaultMaxBodyBytes  = int64(10 * 1024 * 1024)
)

type Config struct {
	Listen        string
	Target        *url.URL
	LogDir        string
	RetentionDays int
	MaxBodyBytes  int64
}

func Load() (Config, error) {
	listen := getEnvOrDefault(EnvListen, DefaultListen)
	if _, _, err := net.SplitHostPort(listen); err != nil {
		return Config{}, fmt.Errorf("%s must be host:port: %w", EnvListen, err)
	}

	targetRaw := getEnvOrDefault(EnvTarget, DefaultTarget)
	target, err := parseTarget(targetRaw)
	if err != nil {
		return Config{}, err
	}

	logDir, err := resolveLogDir(os.Getenv(EnvLogDir))
	if err != nil {
		return Config{}, err
	}

	retentionDays, err := parseRetentionDays(os.Getenv(EnvRetentionDays))
	if err != nil {
		return Config{}, err
	}

	maxBodyBytes, err := parseMaxBodyBytes(os.Getenv(EnvMaxBodyBytes))
	if err != nil {
		return Config{}, err
	}

	return Config{
		Listen:        listen,
		Target:        target,
		LogDir:        logDir,
		RetentionDays: retentionDays,
		MaxBodyBytes:  maxBodyBytes,
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

func parseRetentionDays(raw string) (int, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return DefaultRetentionDays, nil
	}

	value, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0, fmt.Errorf("%s is invalid: %w", EnvRetentionDays, err)
	}
	if value < 0 {
		return 0, fmt.Errorf("%s must be zero or greater", EnvRetentionDays)
	}
	return value, nil
}

func resolveLogDir(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed != "" {
		return trimmed, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("%s could not be resolved: %w", EnvLogDir, err)
	}

	return filepath.Join(homeDir, "Library", "Logs", "ollama-proxy"), nil
}

func getEnvOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
