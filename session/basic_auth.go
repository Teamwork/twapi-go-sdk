package session

import (
	"context"
	"net/http"

	twapi "github.com/teamwork/twapi-go-sdk"
)

var _ twapi.Session = (*BasicAuth)(nil)

// BasicAuth represents a basic authentication session for the Teamwork API.
type BasicAuth struct {
	username string
	password string
	server   string
}

// NewBasicAuth creates a new BasicAuth session with the provided username and
// password. The server parameter must identify the Teamwork installation, such
// as "https://yourcompany.teamwork.com".
func NewBasicAuth(username, password, server string) *BasicAuth {
	return &BasicAuth{
		username: username,
		password: password,
		server:   server,
	}
}

// Authenticate implements the Session interface for BasicAuth.
func (b *BasicAuth) Authenticate(_ context.Context, req *http.Request) error {
	req.SetBasicAuth(b.username, b.password)
	return nil
}

// Server returns the server URL for the BasicAuth session.
func (b *BasicAuth) Server() string {
	return b.server
}
