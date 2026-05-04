package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/joseph/ollama-logging-proxy/internal/config"
	"github.com/joseph/ollama-logging-proxy/internal/logging"
	"github.com/joseph/ollama-logging-proxy/internal/proxy"
	"github.com/joseph/ollama-logging-proxy/internal/redact"
	"github.com/joseph/ollama-logging-proxy/internal/retention"
)

var requestIDCounter atomic.Uint64

func main() {
	if err := run(os.Args[1:]); err != nil {
		log.Printf("%v", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	command := "serve"
	if len(args) > 0 {
		command = args[0]
		args = args[1:]
	}

	switch command {
	case "serve":
		if len(args) != 0 {
			return fmt.Errorf("serve does not accept arguments")
		}
		return runServe()
	case "health":
		if len(args) != 0 {
			return fmt.Errorf("health does not accept arguments")
		}
		return runHealth()
	case "tail":
		return runTail(args)
	case "purge":
		if len(args) != 0 {
			return fmt.Errorf("purge does not accept arguments")
		}
		return runPurge()
	default:
		return fmt.Errorf("unknown command %q (expected serve, health, tail, or purge)", command)
	}
}

func runServe() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	infoLogger := log.New(os.Stdout, "", log.LstdFlags)
	errLogger := log.New(os.Stderr, "", log.LstdFlags)

	bodyWriter, err := logging.NewBodyWriter(cfg.LogDir)
	if err != nil {
		return fmt.Errorf("configure body log writer: %w", err)
	}
	defer func() {
		if closeErr := bodyWriter.Close(); closeErr != nil {
			errLogger.Printf("failed to close body log writer: %v", closeErr)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cleaner := retention.NewCleaner(cfg.LogDir, cfg.RetentionDays)
	retentionErrCh := cleaner.Start(ctx, retention.DefaultInterval)
	go func() {
		for retentionErr := range retentionErrCh {
			errLogger.Printf("retention cleanup failed: %v", retentionErr)
		}
	}()

	handler, err := proxy.New(proxy.Options{
		Target:       cfg.Target,
		MaxBodyBytes: cfg.MaxBodyBytes,
		OnCapture:    buildCaptureHook(bodyWriter, cfg.MaxBodyBytes, errLogger),
	})
	if err != nil {
		return fmt.Errorf("configure proxy: %w", err)
	}

	infoLogger.Printf("ollama-logging-proxy listening on %s", cfg.Listen)
	infoLogger.Printf("forwarding upstream to %s", cfg.Target.String())
	infoLogger.Printf("body logs in %s", cfg.LogDir)

	server := &http.Server{
		Addr:    cfg.Listen,
		Handler: handler,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if shutdownErr := server.Shutdown(shutdownCtx); shutdownErr != nil && !errors.Is(shutdownErr, http.ErrServerClosed) {
			errLogger.Printf("shutdown failed: %v", shutdownErr)
		}
	}()

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("proxy server failed: %w", err)
	}

	return nil
}

func runHealth() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	healthURL, err := healthURL(cfg.Listen)
	if err != nil {
		return fmt.Errorf("build health URL: %w", err)
	}

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(healthURL)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: status=%d", resp.StatusCode)
	}

	var payload struct {
		OK      bool   `json:"ok"`
		Service string `json:"service"`
	}

	if decodeErr := json.NewDecoder(resp.Body).Decode(&payload); decodeErr != nil {
		return fmt.Errorf("health check failed: invalid JSON response: %w", decodeErr)
	}
	if !payload.OK {
		return errors.New("health check failed: ok=false")
	}

	fmt.Printf("ok: %s\n", payload.Service)
	return nil
}

func runTail(args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	lineCount := 100
	if len(args) > 1 {
		return fmt.Errorf("tail accepts at most one optional argument: number of lines")
	}
	if len(args) == 1 {
		parsed, parseErr := strconv.Atoi(args[0])
		if parseErr != nil || parsed <= 0 {
			return fmt.Errorf("tail line count must be a positive integer")
		}
		lineCount = parsed
	}

	filePath := logging.DailyFilePath(cfg.LogDir, time.Now())
	lines, err := tailLines(filePath, lineCount)
	if err != nil {
		return err
	}

	for _, line := range lines {
		fmt.Println(line)
	}

	return nil
}

func runPurge() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	cleaner := retention.NewCleaner(cfg.LogDir, cfg.RetentionDays)
	result, err := cleaner.Cleanup(context.Background())
	if err != nil {
		return fmt.Errorf("purge failed: %w", err)
	}

	fmt.Printf("matched=%d deleted=%d kept=%d skipped=%d\n", result.Matched, result.Deleted, result.Kept, result.Skipped)
	return nil
}

func buildCaptureHook(bodyWriter *logging.BodyWriter, maxBodyBytes int64, errLogger *log.Logger) proxy.CaptureHook {
	maxBytes := clampToInt(maxBodyBytes)

	return func(_ context.Context, event proxy.CaptureEvent) {
		requestBody, requestTruncated, requestRedacted := prepareBodyForLog(event.Request.Body, maxBytes)
		responseBody, responseTruncated, responseRedacted := prepareBodyForLog(event.Response.Body, maxBytes)

		record := logging.BodyLogRecord{
			ID:                newRequestID(),
			StartedAt:         event.Metadata.StartedAt.Format(time.RFC3339),
			DurationMS:        event.Metadata.Duration.Milliseconds(),
			ClientIP:          event.Metadata.ClientIP,
			Method:            event.Metadata.Method,
			Path:              event.Metadata.Path,
			Query:             event.Metadata.Query,
			UserAgent:         event.Metadata.UserAgent,
			Status:            event.Metadata.Status,
			Error:             event.Metadata.Error,
			RequestBody:       requestBody,
			ResponseBody:      responseBody,
			RequestTruncated:  event.Request.Truncated || requestTruncated,
			ResponseTruncated: event.Response.Truncated || responseTruncated,
			RequestRedacted:   requestRedacted,
			ResponseRedacted:  responseRedacted,
		}

		if err := bodyWriter.Write(record); err != nil {
			errLogger.Printf("body log write failed: %v", err)
		}
	}
}

func prepareBodyForLog(body []byte, maxBytes int) (string, bool, bool) {
	if len(body) == 0 {
		return "", false, false
	}

	redactedBody, redacted := redact.RedactImagesJSON(body)
	truncatedBody, truncated := logging.TruncateBody(redactedBody, maxBytes)
	return truncatedBody, truncated, redacted
}

func newRequestID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), requestIDCounter.Add(1))
}

func healthURL(listen string) (string, error) {
	host, port, err := net.SplitHostPort(listen)
	if err != nil {
		return "", fmt.Errorf("invalid listen address %q: %w", listen, err)
	}
	switch host {
	case "", "0.0.0.0", "::":
		host = "127.0.0.1"
	}
	return "http://" + net.JoinHostPort(host, port) + proxy.HealthPath, nil
}

func tailLines(path string, maxLines int) ([]string, error) {
	if maxLines <= 0 {
		return nil, fmt.Errorf("maxLines must be greater than zero")
	}

	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	lines := make([]string, 0, maxLines)
	scanner := bufio.NewScanner(file)
	const maxScanToken = 16 * 1024 * 1024
	scanner.Buffer(make([]byte, 0, 64*1024), maxScanToken)

	for scanner.Scan() {
		line := scanner.Text()
		if len(lines) < maxLines {
			lines = append(lines, line)
			continue
		}

		copy(lines, lines[1:])
		lines[len(lines)-1] = line
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	return lines, nil
}

func clampToInt(value int64) int {
	if value <= 0 {
		return 0
	}

	maxInt := int64(^uint(0) >> 1)
	if value > maxInt {
		return int(maxInt)
	}
	return int(value)
}
