package session

import (
	"context"
	"net/http"

	twapi "github.com/teamwork/twapi-go-sdk"
)

var _ twapi.Session = (*BearerToken)(nil)

// BearerToken represents a session for the Teamwork API using Bearer Token
// authentication. It implements the Session interface, allowing it to be used
// with the Teamwork Engine for authenticated requests.
type BearerToken struct {
	token  string
	server string
}

// NewBearerToken creates a new BearerToken session with the provided token. The
// server parameter must identify the Teamwork installation, such as
// "https://yourcompany.teamwork.com".
func NewBearerToken(token string, server string) *BearerToken {
	return &BearerToken{
		token:  token,
		server: server,
	}
}

// Authenticate implements the Session interface for BearerToken.
func (b *BearerToken) Authenticate(_ context.Context, req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+b.token)
	return nil
}

// Server returns the server URL for the BearerToken session.
func (b *BearerToken) Server() string {
	return b.server
}
