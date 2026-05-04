package logging

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	LogDirEnv = "OLLAMA_PROXY_LOG_DIR"
)

// BodyLogRecord is the canonical JSONL schema for MVP body logs (PRD 7.3.1).
type BodyLogRecord struct {
	ID                string `json:"id"`
	StartedAt         string `json:"started_at"`
	DurationMS        int64  `json:"duration_ms"`
	ClientIP          string `json:"client_ip"`
	Method            string `json:"method"`
	Path              string `json:"path"`
	Query             string `json:"query"`
	UserAgent         string `json:"user_agent"`
	Status            int    `json:"status"`
	Error             string `json:"error"`
	RequestBody       string `json:"request_body"`
	ResponseBody      string `json:"response_body"`
	RequestTruncated  bool   `json:"request_truncated"`
	ResponseTruncated bool   `json:"response_truncated"`
	RequestRedacted   bool   `json:"request_redacted"`
	ResponseRedacted  bool   `json:"response_redacted"`
}

// BodyWriter writes body log records to daily JSONL files.
// It is safe for concurrent use.
type BodyWriter struct {
	logDir string
	now    func() time.Time

	mu          sync.Mutex
	currentDate string
	file        *os.File
}

func NewBodyWriter(logDir string) (*BodyWriter, error) {
	return NewBodyWriterWithClock(logDir, time.Now)
}

func NewBodyWriterWithClock(logDir string, now func() time.Time) (*BodyWriter, error) {
	if strings.TrimSpace(logDir) == "" {
		return nil, errors.New("log directory is required")
	}
	if now == nil {
		now = time.Now
	}
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, fmt.Errorf("create log directory: %w", err)
	}
	return &BodyWriter{
		logDir: logDir,
		now:    now,
	}, nil
}

// NewBodyWriterFromEnv resolves OLLAMA_PROXY_LOG_DIR and creates a writer.
// If the env var is empty, it falls back to ~/Library/Logs/ollama-proxy.
func NewBodyWriterFromEnv() (*BodyWriter, error) {
	logDir, err := ResolveLogDir(os.Getenv(LogDirEnv))
	if err != nil {
		return nil, err
	}
	return NewBodyWriter(logDir)
}

// ResolveLogDir returns an explicit path or the PRD default log directory.
func ResolveLogDir(path string) (string, error) {
	if trimmed := strings.TrimSpace(path); trimmed != "" {
		return trimmed, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, "Library", "Logs", "ollama-proxy"), nil
}

// DailyFilename returns the PRD-required daily body log filename.
func DailyFilename(now time.Time) string {
	return fmt.Sprintf("body-%s.jsonl", now.Format("2006-01-02"))
}

// DailyFilePath returns the full daily JSONL path under the provided directory.
func DailyFilePath(logDir string, now time.Time) string {
	return filepath.Join(logDir, DailyFilename(now))
}

// Write appends one JSON object line to the current daily body log file.
func (w *BodyWriter) Write(record BodyLogRecord) error {
	if w == nil {
		return errors.New("body writer is nil")
	}

	line, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal body log record: %w", err)
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.rotateLocked(w.now()); err != nil {
		return err
	}
	if _, err := w.file.Write(line); err != nil {
		return fmt.Errorf("write body log line: %w", err)
	}
	if _, err := w.file.Write([]byte("\n")); err != nil {
		return fmt.Errorf("write body log newline: %w", err)
	}

	return nil
}

// Close releases the active file descriptor.
func (w *BodyWriter) Close() error {
	if w == nil {
		return nil
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		return nil
	}

	err := w.file.Close()
	w.file = nil
	w.currentDate = ""
	if err != nil {
		return fmt.Errorf("close body log file: %w", err)
	}
	return nil
}

func (w *BodyWriter) rotateLocked(now time.Time) error {
	date := now.Format("2006-01-02")
	if w.file != nil && w.currentDate == date {
		return nil
	}

	if w.file != nil {
		if err := w.file.Close(); err != nil {
			return fmt.Errorf("close previous body log file: %w", err)
		}
		w.file = nil
	}

	path := DailyFilePath(w.logDir, now)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open body log file: %w", err)
	}

	w.file = file
	w.currentDate = date
	return nil
}

// TruncateBody converts body bytes into a log string and marks whether truncation
// occurred due to maxBytes.
func TruncateBody(body []byte, maxBytes int) (string, bool) {
	if maxBytes < 0 {
		return string(body), false
	}
	if len(body) <= maxBytes {
		return string(body), false
	}
	return string(body[:maxBytes]), true
}
