package proxy

import (
	"context"
	"io"
	"sync"
	"time"
)

type BodyCapture struct {
	Body      []byte
	Truncated bool
}

type CaptureMetadata struct {
	StartedAt time.Time
	Duration  time.Duration
	ClientIP  string
	Method    string
	Path      string
	Query     string
	UserAgent string
	Status    int
	Error     string
	Tapped    bool
}

type CaptureEvent struct {
	Metadata CaptureMetadata
	Request  BodyCapture
	Response BodyCapture
}

type CaptureHook func(context.Context, CaptureEvent)

type boundedCapture struct {
	limit int64

	mu        sync.Mutex
	data      []byte
	truncated bool
}

func newBoundedCapture(limit int64) *boundedCapture {
	if limit < 0 {
		limit = 0
	}

	return &boundedCapture{
		limit: limit,
	}
}

func (c *boundedCapture) Write(p []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.limit == 0 {
		if len(p) > 0 {
			c.truncated = true
		}
		return len(p), nil
	}

	remaining := c.limit - int64(len(c.data))
	if remaining <= 0 {
		if len(p) > 0 {
			c.truncated = true
		}
		return len(p), nil
	}

	toCopy := len(p)
	if int64(toCopy) > remaining {
		toCopy = int(remaining)
		c.truncated = true
	}

	c.data = append(c.data, p[:toCopy]...)
	return len(p), nil
}

func (c *boundedCapture) Snapshot() BodyCapture {
	c.mu.Lock()
	defer c.mu.Unlock()

	body := make([]byte, len(c.data))
	copy(body, c.data)

	return BodyCapture{
		Body:      body,
		Truncated: c.truncated,
	}
}

type teeReadCloser struct {
	source io.ReadCloser
	sink   io.Writer
}

func (t *teeReadCloser) Read(p []byte) (int, error) {
	n, err := t.source.Read(p)
	if n > 0 {
		_, _ = t.sink.Write(p[:n])
	}
	return n, err
}

func (t *teeReadCloser) Close() error {
	return t.source.Close()
}
