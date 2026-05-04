package retention

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseDateFromFilename(t *testing.T) {
	t.Parallel()

	valid, ok := ParseDateFromFilename("body-2026-05-04.jsonl")
	if !ok {
		t.Fatalf("expected valid body filename")
	}
	if got := valid.Format("2006-01-02"); got != "2026-05-04" {
		t.Fatalf("unexpected parsed date: %s", got)
	}

	cases := []string{
		"body-2026-05-04.log",
		"body-2026-5-04.jsonl",
		"body-2026-15-04.jsonl",
		"notes-2026-05-04.jsonl",
		"body-2026-05-04.jsonl.bak",
	}

	for _, name := range cases {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, ok := ParseDateFromFilename(name); ok {
				t.Fatalf("expected %q to be ignored", name)
			}
		})
	}
}

func TestCleanupDeletesOldMatchingFilesAndIgnoresNonMatching(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	mustWriteFile(t, filepath.Join(tmpDir, "body-2026-04-23.jsonl"))
	mustWriteFile(t, filepath.Join(tmpDir, "body-2026-04-24.jsonl"))
	mustWriteFile(t, filepath.Join(tmpDir, "body-2026-05-04.jsonl"))
	mustWriteFile(t, filepath.Join(tmpDir, "body-2026-13-01.jsonl"))
	mustWriteFile(t, filepath.Join(tmpDir, "body-2026-04-23.log"))
	mustWriteFile(t, filepath.Join(tmpDir, "notes.txt"))

	if err := os.Mkdir(filepath.Join(tmpDir, "body-2026-01-01.jsonl"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	cleaner := NewCleaner(tmpDir, 10)
	cleaner.Now = func() time.Time {
		return time.Date(2026, time.May, 4, 13, 30, 0, 0, time.Local)
	}

	result, err := cleaner.Cleanup(context.Background())
	if err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	assertNotExists(t, filepath.Join(tmpDir, "body-2026-04-23.jsonl"))
	assertExists(t, filepath.Join(tmpDir, "body-2026-04-24.jsonl"))
	assertExists(t, filepath.Join(tmpDir, "body-2026-05-04.jsonl"))
	assertExists(t, filepath.Join(tmpDir, "body-2026-13-01.jsonl"))
	assertExists(t, filepath.Join(tmpDir, "body-2026-04-23.log"))
	assertExists(t, filepath.Join(tmpDir, "notes.txt"))

	if result.Matched != 3 {
		t.Fatalf("expected 3 matched files, got %d", result.Matched)
	}
	if result.Deleted != 1 {
		t.Fatalf("expected 1 deleted file, got %d", result.Deleted)
	}
	if result.Kept != 2 {
		t.Fatalf("expected 2 kept files, got %d", result.Kept)
	}
	if result.Skipped != 4 {
		t.Fatalf("expected 4 skipped files, got %d", result.Skipped)
	}
}

func TestCleanupUsesFilenameDateOnly(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	oldByName := filepath.Join(tmpDir, "body-2020-01-01.jsonl")
	newByName := filepath.Join(tmpDir, "body-2026-05-04.jsonl")

	mustWriteFile(t, oldByName)
	mustWriteFile(t, newByName)

	now := time.Date(2026, time.May, 4, 9, 0, 0, 0, time.Local)
	oneYearAgo := now.AddDate(-1, 0, 0)

	if err := os.Chtimes(oldByName, now, now); err != nil {
		t.Fatalf("chtimes oldByName: %v", err)
	}
	if err := os.Chtimes(newByName, oneYearAgo, oneYearAgo); err != nil {
		t.Fatalf("chtimes newByName: %v", err)
	}

	cleaner := NewCleaner(tmpDir, 10)
	cleaner.Now = func() time.Time { return now }

	if _, err := cleaner.Cleanup(context.Background()); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	assertNotExists(t, oldByName)
	assertExists(t, newByName)
}

func mustWriteFile(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("test"), 0o644); err != nil {
		t.Fatalf("write file %q: %v", path, err)
	}
}

func assertExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected %q to exist: %v", path, err)
	}
}

func assertNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected %q to be removed, stat error: %v", path, err)
	}
}
