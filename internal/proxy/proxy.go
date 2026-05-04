package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

const defaultMaxBodyBytes = int64(10 * 1024 * 1024)

type Options struct {
	Target       *url.URL
	MaxBodyBytes int64
	OnCapture    CaptureHook
	Now          func() time.Time
}

type handler struct {
	reverseProxy *httputil.ReverseProxy
	maxBodyBytes int64
	onCapture    CaptureHook
	now          func() time.Time
}

type requestState struct {
	tapped bool

	startedAt time.Time
	status    int
	errorText string

	clientIP  string
	method    string
	path      string
	query     string
	userAgent string

	requestBodyCapture  *boundedCapture
	responseBodyCapture *boundedCapture
}

type requestStateKey struct{}

func New(options Options) (http.Handler, error) {
	if options.Target == nil {
		return nil, errors.New("target is required")
	}
	if options.Target.Scheme == "" || options.Target.Host == "" {
		return nil, errors.New("target must include scheme and host")
	}

	if options.MaxBodyBytes <= 0 {
		options.MaxBodyBytes = defaultMaxBodyBytes
	}
	if options.Now == nil {
		options.Now = time.Now
	}

	h := &handler{
		maxBodyBytes: options.MaxBodyBytes,
		onCapture:    options.OnCapture,
		now:          options.Now,
	}

	rp := httputil.NewSingleHostReverseProxy(options.Target)
	rp.FlushInterval = -1
	rp.ModifyResponse = h.modifyResponse
	rp.ErrorHandler = h.errorHandler
	h.reverseProxy = rp

	return h, nil
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := NormalizePath(r.URL.RequestURI())
	if path == HealthPath {
		writeHealthResponse(w)
		return
	}

	tapped := IsTappedPath(r.URL.RequestURI())
	state := &requestState{
		tapped: tapped,

		startedAt: h.now().UTC(),
		clientIP:  clientIPFromRemoteAddr(r.RemoteAddr),
		method:    r.Method,
		path:      path,
		query:     r.URL.RawQuery,
		userAgent: r.UserAgent(),

		requestBodyCapture:  newBoundedCapture(h.maxBodyBytes),
		responseBodyCapture: newBoundedCapture(h.maxBodyBytes),
	}

	if tapped && hasBody(r) {
		r.Body = &teeReadCloser{
			source: r.Body,
			sink:   state.requestBodyCapture,
		}
	}

	ctx := context.WithValue(r.Context(), requestStateKey{}, state)
	h.reverseProxy.ServeHTTP(w, r.WithContext(ctx))

	if !tapped || h.onCapture == nil {
		return
	}

	if state.status == 0 {
		state.status = http.StatusOK
	}

	h.onCapture(ctx, CaptureEvent{
		Metadata: CaptureMetadata{
			StartedAt: state.startedAt,
			Duration:  h.now().UTC().Sub(state.startedAt),
			ClientIP:  state.clientIP,
			Method:    state.method,
			Path:      state.path,
			Query:     state.query,
			UserAgent: state.userAgent,
			Status:    state.status,
			Error:     state.errorText,
			Tapped:    state.tapped,
		},
		Request:  state.requestBodyCapture.Snapshot(),
		Response: state.responseBodyCapture.Snapshot(),
	})
}

func (h *handler) modifyResponse(resp *http.Response) error {
	state := requestStateFromContext(resp.Request.Context())
	if state == nil {
		return nil
	}

	state.status = resp.StatusCode
	if !state.tapped || resp.Body == nil || resp.Body == http.NoBody {
		return nil
	}

	resp.Body = &teeReadCloser{
		source: resp.Body,
		sink:   state.responseBodyCapture,
	}
	return nil
}

func (h *handler) errorHandler(w http.ResponseWriter, r *http.Request, err error) {
	state := requestStateFromContext(r.Context())
	if state != nil {
		state.status = http.StatusBadGateway
		if err != nil {
			state.errorText = err.Error()
		}
	}

	body := []byte("Bad Gateway\n")

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusBadGateway)
	_, _ = w.Write(body)

	if state != nil && state.tapped {
		_, _ = state.responseBodyCapture.Write(body)
	}
}

func requestStateFromContext(ctx context.Context) *requestState {
	state, ok := ctx.Value(requestStateKey{}).(*requestState)
	if !ok {
		return nil
	}
	return state
}

func hasBody(r *http.Request) bool {
	return r != nil && r.Body != nil && r.Body != http.NoBody
}

func clientIPFromRemoteAddr(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err == nil {
		return host
	}
	return remoteAddr
}

func writeHealthResponse(w http.ResponseWriter) {
	response := map[string]any{
		"ok":      true,
		"service": "ollama-logging-proxy",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}
