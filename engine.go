package twapi

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
)

// HTTPRequester knows how to create an HTTP request for a specific entity.
type HTTPRequester interface {
	HTTPRequest(ctx context.Context, server string) (*http.Request, error)
}

// HTTPResponser knows how to handle an HTTP response for a specific entity.
type HTTPResponser interface {
	HandleHTTPResponse(resp *http.Response) error
}

// Session is an interface that defines the methods required for a session to
// authenticate requests to the Teamwork Engine.
type Session interface {
	Authenticate(ctx context.Context, req *http.Request) error
	Server() string
}

// Engine is the main structure that handles communication with the Teamwork
// API.
type Engine struct {
	client  *http.Client
	session Session
	logger  *slog.Logger
}

// EngineOption is a function that modifies the Engine configuration.
type EngineOption func(*Engine)

// WithHTTPClient sets the HTTP client for the Engine. By default, it uses
// http.DefaultClient.
func WithHTTPClient(client *http.Client) EngineOption {
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
func Execute[T HTTPResponser](ctx context.Context, engine *Engine, requester HTTPRequester) (T, error) {
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
	return responser, nil
}
