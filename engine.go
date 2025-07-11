package twapi

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
)

// HTTPClient is an interface that defines the methods required for an HTTP
// client. This allows for adding middlewares and easier testing.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// HTTPClientFunc is a function type that implements the HTTPClient interface.
type HTTPClientFunc func(req *http.Request) (*http.Response, error)

// Do executes the HTTP request and returns the response.
func (f HTTPClientFunc) Do(req *http.Request) (*http.Response, error) {
	return f(req)
}

// HTTPRequester knows how to create an HTTP request for a specific entity.
type HTTPRequester interface {
	HTTPRequest(ctx context.Context, server string) (*http.Request, error)
}

// HTTPResponser knows how to handle an HTTP response for a specific entity.
type HTTPResponser interface {
	HandleHTTPResponse(*http.Response) error
}

// Session is an interface that defines the methods required for a session to
// authenticate requests to the Teamwork Engine.
type Session interface {
	Authenticate(context.Context, *http.Request) error
	Server() string
}

// httpClientMiddleware is a wrapper around an HTTP client that applies a
// middleware function to the client.
type httpClientMiddleware struct {
	client     HTTPClient
	middleware func(HTTPClient) HTTPClient
}

// Do executes the HTTP request with the middleware applied.
func (m *httpClientMiddleware) Do(req *http.Request) (*http.Response, error) {
	return m.middleware(m.client).Do(req)
}

// Engine is the main structure that handles communication with the Teamwork
// API.
type Engine struct {
	client  HTTPClient
	session Session
	logger  *slog.Logger
}

// EngineOption is a function that modifies the Engine configuration.
type EngineOption func(*Engine)

// WithHTTPClient sets the HTTP client for the Engine. By default, it uses
// http.DefaultClient. When setting the HTTP client, any middlewares that were
// added using WithMiddleware before this call will be ignored.
func WithHTTPClient(client HTTPClient) EngineOption {
	return func(e *Engine) {
		e.client = client
	}
}

// WithLogger sets the logger for the Engine. By default, it uses
// slog.Default().
func WithLogger(logger *slog.Logger) EngineOption {
	return func(e *Engine) {
		e.logger = logger
	}
}

// WithMiddleware adds a middleware to the Engine. Middlewares are applied in
// the order they are added.
func WithMiddleware(middleware func(HTTPClient) HTTPClient) EngineOption {
	return func(e *Engine) {
		e.client = &httpClientMiddleware{
			client:     e.client,
			middleware: middleware,
		}
	}
}

// NewEngine creates a new Engine instance with the provided HTTP client and
// session.
func NewEngine(session Session, opts ...EngineOption) *Engine {
	e := &Engine{
		client:  http.DefaultClient,
		session: session,
		logger:  slog.Default(),
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// Execute sends an HTTP request using the provided requester and handles the
// response using the provided responser.
func Execute[R HTTPRequester, T HTTPResponser](ctx context.Context, engine *Engine, requester R) (T, error) {
	var responser T
	if rt := reflect.TypeOf(responser); rt.Kind() == reflect.Ptr {
		responser = reflect.New(rt.Elem()).Interface().(T)
	}

	req, err := requester.HTTPRequest(ctx, engine.session.Server())
	if err != nil {
		return responser, fmt.Errorf("failed to create request: %w", err)
	}
	if err := engine.session.Authenticate(ctx, req); err != nil {
		return responser, fmt.Errorf("failed to authenticate request: %w", err)
	}

	resp, err := engine.client.Do(req)
	if err != nil {
		return responser, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			engine.logger.Error("failed to close response body",
				slog.String("error", err.Error()),
			)
		}
	}()

	if err := responser.HandleHTTPResponse(resp); err != nil {
		return responser, fmt.Errorf("failed to handle response: %w", err)
	}

	if paginated, ok := any(responser).(interface{ SetRequest(req R) }); ok {
		paginated.SetRequest(requester)
	}

	return responser, nil
}
