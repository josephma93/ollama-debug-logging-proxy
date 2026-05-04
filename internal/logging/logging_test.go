package logging

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestTruncateBody(t *testing.T) {
	tests := []struct {
		name      string
		body      []byte
		maxBytes  int
		wantBody  string
		wantTrunc bool
	}{
		{name: "within limit", body: []byte("hello"), maxBytes: 5, wantBody: "hello", wantTrunc: false},
		{name: "below limit", body: []byte("hello"), maxBytes: 10, wantBody: "hello", wantTrunc: false},
		{name: "exceeds limit", body: []byte("hello"), maxBytes: 4, wantBody: "hell", wantTrunc: true},
		{name: "zero limit with body", body: []byte("hello"), maxBytes: 0, wantBody: "", wantTrunc: true},
		{name: "zero limit empty body", body: []byte(""), maxBytes: 0, wantBody: "", wantTrunc: false},
		{name: "negative limit disables truncation", body: []byte("hello"), maxBytes: -1, wantBody: "hello", wantTrunc: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			gotBody, gotTrunc := TruncateBody(tt.body, tt.maxBytes)
			if gotBody != tt.wantBody || gotTrunc != tt.wantTrunc {
				t.Fatalf("TruncateBody(...) = (%q, %t), want (%q, %t)", gotBody, gotTrunc, tt.wantBody, tt.wantTrunc)
			}
		})
	}
}

func TestBodyLogRecordJSONSchemaFields(t *testing.T) {
	record := BodyLogRecord{}
	raw, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	expectedOrder := []string{
		`"id":`,
		`"started_at":`,
		`"duration_ms":`,
		`"client_ip":`,
		`"method":`,
		`"path":`,
		`"query":`,
		`"user_agent":`,
		`"status":`,
		`"error":`,
		`"request_body":`,
		`"response_body":`,
		`"request_truncated":`,
		`"response_truncated":`,
		`"request_redacted":`,
		`"response_redacted":`,
	}

	rawString := string(raw)
	lastIdx := -1
	for _, token := range expectedOrder {
		idx := strings.Index(rawString, token)
		if idx == -1 {
			t.Fatalf("missing field token %s in %s", token, rawString)
		}
		if idx <= lastIdx {
			t.Fatalf("field order mismatch for %s in %s", token, rawString)
		}
		lastIdx = idx
	}
}

func TestDailyFilenameAndRotation(t *testing.T) {
	logDir := t.TempDir()
	times := []time.Time{
		time.Date(2026, 5, 4, 10, 0, 0, 0, time.UTC),
		time.Date(2026, 5, 4, 10, 30, 0, 0, time.UTC),
		time.Date(2026, 5, 5, 1, 0, 0, 0, time.UTC),
	}

	i := 0
	writer, err := NewBodyWriterWithClock(logDir, func() time.Time {
		return times[i]
	})
	if err != nil {
		t.Fatalf("NewBodyWriterWithClock: %v", err)
	}
	defer func() {
		if closeErr := writer.Close(); closeErr != nil {
			t.Fatalf("Close: %v", closeErr)
		}
	}()

	if got := DailyFilename(times[0]); got != "body-2026-05-04.jsonl" {
		t.Fatalf("DailyFilename = %q", got)
	}

	if err := writer.Write(testRecord("r1")); err != nil {
		t.Fatalf("write record 1: %v", err)
	}

	i = 1
	if err := writer.Write(testRecord("r2")); err != nil {
		t.Fatalf("write record 2: %v", err)
	}

	i = 2
	if err := writer.Write(testRecord("r3")); err != nil {
		t.Fatalf("write record 3: %v", err)
	}

	day1Path := DailyFilePath(logDir, times[0])
	day2Path := DailyFilePath(logDir, times[2])

	day1Lines := readJSONLLines(t, day1Path)
	day2Lines := readJSONLLines(t, day2Path)

	if len(day1Lines) != 2 {
		t.Fatalf("day1 line count=%d want=2", len(day1Lines))
	}
	if len(day2Lines) != 1 {
		t.Fatalf("day2 line count=%d want=1", len(day2Lines))
	}

	validateJSONL(t, day1Lines)
	validateJSONL(t, day2Lines)
}

func TestBodyWriterConcurrentWrites(t *testing.T) {
	logDir := t.TempDir()
	now := time.Date(2026, 5, 4, 12, 0, 0, 0, time.UTC)

	writer, err := NewBodyWriterWithClock(logDir, func() time.Time {
		return now
	})
	if err != nil {
		t.Fatalf("NewBodyWriterWithClock: %v", err)
	}
	defer func() {
		if closeErr := writer.Close(); closeErr != nil {
			t.Fatalf("Close: %v", closeErr)
		}
	}()

	const total = 250
	var wg sync.WaitGroup
	wg.Add(total)

	for i := 0; i < total; i++ {
		i := i
		go func() {
			defer wg.Done()
			if err := writer.Write(testRecord(fmt.Sprintf("id-%d", i))); err != nil {
				t.Errorf("write %d: %v", i, err)
			}
		}()
	}
	wg.Wait()

	path := DailyFilePath(logDir, now)
	lines := readJSONLLines(t, path)
	if len(lines) != total {
		t.Fatalf("line count=%d want=%d", len(lines), total)
	}
	validateJSONL(t, lines)
}

func TestNewBodyWriterFromEnv(t *testing.T) {
	logDir := t.TempDir()
	t.Setenv(LogDirEnv, logDir)

	writer, err := NewBodyWriterFromEnv()
	if err != nil {
		t.Fatalf("NewBodyWriterFromEnv: %v", err)
	}
	defer func() {
		if closeErr := writer.Close(); closeErr != nil {
			t.Fatalf("Close: %v", closeErr)
		}
	}()

	if err := writer.Write(testRecord("from-env")); err != nil {
		t.Fatalf("Write: %v", err)
	}

	entries, err := os.ReadDir(logDir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatalf("expected at least one log file in %s", logDir)
	}
	if ext := filepath.Ext(entries[0].Name()); ext != ".jsonl" {
		t.Fatalf("expected jsonl file, got %q", entries[0].Name())
	}
}

func testRecord(id string) BodyLogRecord {
	return BodyLogRecord{
		ID:                id,
		StartedAt:         "2026-05-04T12:00:00Z",
		DurationMS:        12,
		ClientIP:          "127.0.0.1",
		Method:            "POST",
		Path:              "/api/chat",
		Query:             "",
		UserAgent:         "test-agent",
		Status:            200,
		Error:             "",
		RequestBody:       `{"prompt":"hello"}`,
		ResponseBody:      `{"response":"hi"}`,
		RequestTruncated:  false,
		ResponseTruncated: false,
		RequestRedacted:   false,
		ResponseRedacted:  false,
	}
}

func readJSONLLines(t *testing.T, path string) []string {
	t.Helper()

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan %s: %v", path, err)
	}
	return lines
}

func validateJSONL(t *testing.T, lines []string) {
	t.Helper()
	for i, line := range lines {
		if !json.Valid([]byte(line)) {
			t.Fatalf("line %d is not valid JSON: %s", i, line)
		}
	}
}
