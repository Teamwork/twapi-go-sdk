package session

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"

	twapi "github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/internal/browser"
)

const (
	defaultOAuthServer  = "https://teamwork.com"
	defaultCallbackHost = "localhost:0"
)

var _ twapi.Session = (*OAuth2)(nil)

// OAuth2 represents a session for the Teamwork API using OAuth2 authentication.
// It implements the Session interface, allowing it to be used with the Teamwork
// Engine for authenticated requests.
type OAuth2 struct {
	client             *http.Client
	clientID           string
	clientSecret       string
	oauthServer        string
	callbackServerAddr string
	logger             *slog.Logger

	oauthMutex  sync.Mutex
	server      string
	bearerToken string
}

// OAuth2Option defines a function type that can modify the OAuth2 initial
// configuration.
type OAuth2Option func(*OAuth2)

// WithOAuth2Client sets the HTTP client for the OAuth2 session. By default, it
// uses http.DefaultClient.
func WithOAuth2Client(client *http.Client) OAuth2Option {
	return func(o *OAuth2) {
		o.client = client
	}
}

// WithOAuth2Server sets the OAuth2 authorization server URL. By default, it
// uses "https://teamwork.com".
func WithOAuth2Server(server string) OAuth2Option {
	return func(o *OAuth2) {
		server = strings.TrimSuffix(server, "/")
		if u, err := url.Parse(server); err == nil {
			if u.Scheme == "" {
				u.Scheme = "https"
			}
			o.oauthServer = u.String()
		}
	}
}

// WithOAuth2CallbackServerAddr sets the OAuth2 callback host. By default, it
// uses "localhost:0", which means the callback will be handled on a random port
// on localhost.
func WithOAuth2CallbackServerAddr(host string) OAuth2Option {
	return func(o *OAuth2) {
		if _, _, err := net.SplitHostPort(host); err == nil {
			o.callbackServerAddr = host
		}
	}
}

// WithOAuth2Logger sets the logger for the OAuth2 session. If not set, it uses
// a default logger that writes to the standard output.
func WithOAuth2Logger(logger *slog.Logger) OAuth2Option {
	return func(o *OAuth2) {
		if logger != nil {
			o.logger = logger
		} else {
			o.logger = slog.Default()
		}
	}
}

// NewOAuth2 creates a new OAuth2 session with the provided client ID and
// client secret.
func NewOAuth2(clientID, clientSecret string, opts ...OAuth2Option) *OAuth2 {
	oauth := &OAuth2{
		client:             http.DefaultClient,
		clientID:           clientID,
		clientSecret:       clientSecret,
		oauthServer:        defaultOAuthServer,
		callbackServerAddr: defaultCallbackHost,
		logger:             slog.Default(),
	}
	for _, opt := range opts {
		opt(oauth)
	}
	return oauth
}

// Authenticate implements the Session interface for OAuth2.
func (o *OAuth2) Authenticate(ctx context.Context, req *http.Request) error {
	if o.bearerToken == "" {
		if err := o.handshake(ctx); err != nil {
			return fmt.Errorf("failed to authenticate with oauth2: %w", err)
		}
	}

	if req.URL.Host == "" {
		serverURL, err := url.Parse(o.server)
		if err != nil {
			return fmt.Errorf("failed to parse server URL %q: %w", o.server, err)
		}
		req.URL.Scheme = serverURL.Scheme
		req.URL.Host = serverURL.Host
	}

	req.Header.Set("Authorization", "Bearer "+o.bearerToken)
	return nil
}

// Server returns the server URL for the OAuth2 session. If the authentication
// did not happen yet this may be empty.
func (o *OAuth2) Server() string {
	return o.server
}

// BearerToken returns the bearer token for the OAuth2 session. If the
// authentication did not happen yet this may be empty.
func (o *OAuth2) BearerToken() string {
	return o.bearerToken
}

func (o *OAuth2) handshake(ctx context.Context) error {
	o.oauthMutex.Lock()
	defer o.oauthMutex.Unlock()

	if o.bearerToken != "" && o.server != "" {
		return nil
	}

	serverInfo, err := o.serverInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get oauth2 server info: %w", err)
	}

	codeChallenge, err := o.generateCodeChallenge()
	if err != nil {
		return fmt.Errorf("failed to generate code challenge: %w", err)
	}

	code, redirectURL, err := o.retrieveCode(ctx, serverInfo, codeChallenge)
	if err != nil {
		return fmt.Errorf("failed to retrieve authorization code: %w", err)
	}

	if err := o.retrieveToken(ctx, serverInfo, redirectURL, codeChallenge, code); err != nil {
		return fmt.Errorf("failed to retrieve access token: %w", err)
	}

	return nil
}

func (o *OAuth2) retrieveCode(
	ctx context.Context,
	serverInfo *oauth2ServerInfo,
	codeChallenge string,
) (string, string, error) {
	authURL, err := url.Parse(serverInfo.AuthorizationEndpoint)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse authorization endpoint %q: %w", serverInfo.AuthorizationEndpoint, err)
	}

	listener, err := net.Listen("tcp", o.callbackServerAddr)
	if err != nil {
		return "", "", fmt.Errorf("failed to start listener on %q: %w", o.callbackServerAddr, err)
	}

	type serverResult struct {
		err  error
		code string
	}
	serverResultChannel := make(chan serverResult, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/oauth2/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			serverResultChannel <- serverResult{err: fmt.Errorf("method not allowed")}
			return
		}

		var code string
		for _, supported := range serverInfo.ResponseTypesSupported {
			if code = r.URL.Query().Get(supported); code != "" {
				break
			}
		}

		if code == "" {
			http.Error(w, "Missing code parameter", http.StatusBadRequest)
			serverResultChannel <- serverResult{err: fmt.Errorf("missing code parameter")}
			return
		}

		message := "OAuth2 authentication successful. You can close this window."
		if _, err := w.Write([]byte(message)); err != nil {
			o.logger.Error("failed to write response in oauth2 callback",
				slog.String("message", message),
				slog.String("error", err.Error()),
			)
		}
		serverResultChannel <- serverResult{code: code}
	})

	server := &http.Server{Handler: mux}
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			o.logger.Error("failed to start oauth2 callback server",
				slog.String("address", listener.Addr().String()),
				slog.String("error", err.Error()),
			)
		}
	}()

	redirectURL := fmt.Sprintf("http://%s/oauth2/callback", listener.Addr().String())

	query := authURL.Query()
	query.Set("client_id", o.clientID)
	query.Set("response_type", "code")
	query.Set("redirect_uri", redirectURL)
	query.Set("code_challenge", codeChallenge)
	query.Set("code_challenge_method", "S256")
	authURL.RawQuery = query.Encode()

	if err := browser.OpenURL(authURL.String()); err != nil {
		return "", "", fmt.Errorf("failed to open browser for oauth2 authentication: %w", err)
	}

	var code string

	select {
	case result := <-serverResultChannel:
		if err := server.Close(); err != nil {
			o.logger.Error("failed to close oauth2 callback server",
				slog.String("address", listener.Addr().String()),
				slog.String("error", err.Error()),
			)
		}
		code = result.code

	case <-ctx.Done():
		if err := server.Close(); err != nil {
			o.logger.Error("failed to close oauth2 callback server due to context cancellation",
				slog.String("address", listener.Addr().String()),
				slog.String("error", err.Error()),
			)
		}
		return "", "", ctx.Err()
	}

	if code == "" {
		return "", "", fmt.Errorf("failed to retrieve authorization code from callback")
	}

	return code, redirectURL, nil
}

func (o *OAuth2) retrieveToken(
	ctx context.Context,
	serverInfo *oauth2ServerInfo,
	redirectURL, codeChallenge, code string,
) error {
	if !slices.Contains(serverInfo.TokenEndpointAuthMethodsSupported, "client_secret_post") {
		return fmt.Errorf("unsupported token endpoint authentication methods: %v",
			serverInfo.TokenEndpointAuthMethodsSupported)
	}

	form := url.Values{}
	form.Add("grant_type", "authorization_code")
	form.Add("code", code)
	form.Add("client_id", o.clientID)
	form.Add("client_secret", o.clientSecret)
	form.Add("redirect_uri", redirectURL)
	form.Add("code_verifier", codeChallenge)

	body := bytes.NewBufferString(form.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, serverInfo.TokenEndpoint, body)
	if err != nil {
		return fmt.Errorf("failed to build request to %q: %w", serverInfo.TokenEndpoint, err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := o.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to %q: %w", serverInfo.TokenEndpoint, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			o.logger.Error("failed to close connection in oauth2 token retrieval",
				slog.String("url", serverInfo.TokenEndpoint),
				slog.String("error", err.Error()),
			)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		if len(body) == 0 {
			body = []byte("no response body")
		}
		return fmt.Errorf("unexpected status code %d from %q: %s", resp.StatusCode, serverInfo.TokenEndpoint, string(body))
	}

	var tokenResponse oauth2TokenResponse
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&tokenResponse); err != nil {
		return fmt.Errorf("failed to decode response from %q: %w", serverInfo.TokenEndpoint, err)
	}

	if tokenResponse.AccessToken == "" {
		return fmt.Errorf("missing access token in response from %q", serverInfo.TokenEndpoint)
	}
	if tokenResponse.TokenType != "Bearer" {
		return fmt.Errorf("unexpected token type %q from %q", tokenResponse.TokenType, serverInfo.TokenEndpoint)
	}

	o.bearerToken = tokenResponse.AccessToken
	o.server = tokenResponse.Installation.APIEndpoint
	return nil
}

func (o *OAuth2) serverInfo(ctx context.Context) (*oauth2ServerInfo, error) {
	url := o.oauthServer + "/.well-known/oauth-authorization-server"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request to %q: %w", url, err)
	}

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to %q: %w", url, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			o.logger.Error("failed to close connection in oauth2 server info",
				slog.String("url", url),
				slog.String("error", err.Error()),
			)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		if len(body) == 0 {
			body = []byte("no response body")
		}
		return nil, fmt.Errorf("unexpected status code %d from %q: %s", resp.StatusCode, url, string(body))
	}

	var info oauth2ServerInfo
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode response from %q: %w", url, err)
	}
	if info.AuthorizationEndpoint == "" || info.TokenEndpoint == "" {
		return nil, fmt.Errorf("incomplete server info from %q: %+v", url, info)
	}
	return &info, nil
}

func (o *OAuth2) generateCodeChallenge() (string, error) {
	verifier := make([]byte, 32)
	if _, err := rand.Read(verifier); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	challenge := sha256.Sum256(verifier)
	return base64.RawURLEncoding.EncodeToString(challenge[:]), nil
}

type oauth2ServerInfo struct {
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
}

type oauth2TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Installation struct {
		APIEndpoint string `json:"apiEndpoint"`
	} `json:"installation"`
}
