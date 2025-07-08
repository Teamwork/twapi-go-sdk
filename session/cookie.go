package session

import (
	"context"
	"net/http"

	twapi "github.com/teamwork/twapi-go-sdk"
)

const cookieName = "tw-auth"

var _ twapi.Session = (*Cookie)(nil)

// Cookie represents a session for the Teamwork API using cookie-based
// authentication. It implements the Session interface, allowing it to be used
// with the Teamwork Engine for authenticated requests.
type Cookie struct {
	value  string
	server string
}

// NewCookie creates a new Cookie session with the provided value. The server
// parameter must identify the Teamwork installation, such as
// "https://yourcompany.teamwork.com".
func NewCookie(value string, server string) *Cookie {
	return &Cookie{
		value:  value,
		server: server,
	}
}

// Authenticate implements the Session interface for Cookie.
func (c *Cookie) Authenticate(_ context.Context, req *http.Request) error {
	req.AddCookie(&http.Cookie{
		Name:  cookieName,
		Value: c.value,
	})
	return nil
}

// Server returns the server URL for the Cookie session.
func (c *Cookie) Server() string {
	return c.server
}
