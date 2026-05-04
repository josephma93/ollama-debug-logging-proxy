package retention

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

const (
	filenameDateLayout = "2006-01-02"
	// DefaultInterval is suitable for the PRD's startup + hourly retention cadence.
	DefaultInterval = time.Hour
)

var bodyFilenamePattern = regexp.MustCompile(`^body-(\d{4}-\d{2}-\d{2})\.jsonl$`)

// Result contains a summary of one cleanup pass.
type Result struct {
	Matched int
	Deleted int
	Kept    int
	Skipped int
}

// Cleaner removes old body log files based on filename date.
type Cleaner struct {
	Dir           string
	RetentionDays int
	Now           func() time.Time
}

func NewCleaner(dir string, retentionDays int) *Cleaner {
	return &Cleaner{
		Dir:           dir,
		RetentionDays: retentionDays,
		Now:           time.Now,
	}
}

// ParseDateFromFilename extracts YYYY-MM-DD from body-YYYY-MM-DD.jsonl.
func ParseDateFromFilename(name string) (time.Time, bool) {
	matches := bodyFilenamePattern.FindStringSubmatch(name)
	if len(matches) != 2 {
		return time.Time{}, false
	}

	parsed, err := time.ParseInLocation(filenameDateLayout, matches[1], time.UTC)
	if err != nil {
		return time.Time{}, false
	}

	return parsed, true
}

func (c *Cleaner) now() time.Time {
	if c.Now != nil {
		return c.Now()
	}

	return time.Now()
}

func normalizeContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}

	return ctx
}

// Cleanup deletes files older than the configured retention period.
// It only considers files named body-YYYY-MM-DD.jsonl and ignores all others.
func (c *Cleaner) Cleanup(ctx context.Context) (Result, error) {
	var result Result
	ctx = normalizeContext(ctx)

	if c == nil {
		return result, errors.New("retention: cleaner is nil")
	}
	if c.Dir == "" {
		return result, errors.New("retention: dir is required")
	}
	if c.RetentionDays < 0 {
		return result, fmt.Errorf("retention: retention days must be >= 0: %d", c.RetentionDays)
	}

	entries, err := os.ReadDir(c.Dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return result, nil
		}
		return result, fmt.Errorf("retention: read dir: %w", err)
	}

	now := c.now().In(time.Local)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	cutoff := today.AddDate(0, 0, -c.RetentionDays)

	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		if entry.IsDir() {
			result.Skipped++
			continue
		}

		fileDate, ok := ParseDateFromFilename(entry.Name())
		if !ok {
			result.Skipped++
			continue
		}

		result.Matched++
		fileDate = time.Date(fileDate.Year(), fileDate.Month(), fileDate.Day(), 0, 0, 0, 0, today.Location())
		if !fileDate.Before(cutoff) {
			result.Kept++
			continue
		}

		if err := os.Remove(filepath.Join(c.Dir, entry.Name())); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return result, fmt.Errorf("retention: remove %q: %w", entry.Name(), err)
		}

		result.Deleted++
	}

	return result, nil
}

// Start runs cleanup immediately, then on each interval tick until ctx is canceled.
// Use interval <= 0 to fall back to DefaultInterval (1 hour).
func (c *Cleaner) Start(ctx context.Context, interval time.Duration) <-chan error {
	errCh := make(chan error, 1)
	ctx = normalizeContext(ctx)
	if interval <= 0 {
		interval = DefaultInterval
	}

	go func() {
		defer close(errCh)

		if _, err := c.Cleanup(ctx); err != nil && !errors.Is(err, context.Canceled) {
			select {
			case errCh <- err:
			default:
			}
		}

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if _, err := c.Cleanup(ctx); err != nil && !errors.Is(err, context.Canceled) {
					select {
					case errCh <- err:
					default:
					}
				}
			}
		}
	}()

	return errCh
}
