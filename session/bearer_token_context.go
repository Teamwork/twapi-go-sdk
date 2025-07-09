package session

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	twapi "github.com/teamwork/twapi-go-sdk"
)

var _ twapi.Session = (*BearerTokenContext)(nil)

type bearerTokenContextKey struct{}

// BearerTokenContext represents a session that uses a bearer token for
// authentication extracted from the context. It's possible to inject the token
// into the context using the WithBearerTokenContext function.
type BearerTokenContext struct{}

// NewBearerTokenContext creates a new BearerTokenContext instance.
func NewBearerTokenContext() *BearerTokenContext {
	return &BearerTokenContext{}
}

// Authenticate implements the Session interface for BearerTokenContext.
func (b *BearerTokenContext) Authenticate(ctx context.Context, req *http.Request) error {
	bearerToken, ok := fromBearerTokenContext(ctx)
	if !ok || bearerToken == nil {
		return errors.New("missing bearer token")
	}
	req.Header.Set("Authorization", "Bearer "+bearerToken.token)

	if req.URL.Host == "" {
		serverURL, err := url.Parse(bearerToken.server)
		if err != nil {
			return fmt.Errorf("failed to parse server URL %q: %w", bearerToken.server, err)
		}
		req.URL.Scheme = serverURL.Scheme
		req.URL.Host = serverURL.Host
	}

	return nil
}

// Server returns the server URL for the BearerTokenContext session. This is a
// dummy implementation since it's not accessible at this point.
func (b *BearerTokenContext) Server() string {
	return ""
}

// WithBearerTokenContext returns a new context with the provided BearerToken.
func WithBearerTokenContext(ctx context.Context, bearerToken *BearerToken) context.Context {
	return context.WithValue(ctx, bearerTokenContextKey{}, bearerToken)
}

func fromBearerTokenContext(ctx context.Context) (*BearerToken, bool) {
	if ctx == nil {
		return nil, false
	}
	bearerToken, ok := ctx.Value(bearerTokenContextKey{}).(*BearerToken)
	return bearerToken, ok
}
