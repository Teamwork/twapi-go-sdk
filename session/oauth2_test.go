package session

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func TestGeneratePKCE(t *testing.T) {
	verifier, challenge, err := generatePKCE()
	if err != nil {
		t.Fatalf("generatePKCE: %v", err)
	}

	// RFC 7636 section 4.1: 43 to 128 characters from the unreserved set.
	if !regexp.MustCompile(`^[A-Za-z0-9\-._~]{43,128}$`).MatchString(verifier) {
		t.Errorf("verifier %q is not a valid RFC 7636 code verifier", verifier)
	}

	// RFC 7636 section 4.2: S256 challenge is BASE64URL(SHA256(verifier)).
	sum := sha256.Sum256([]byte(verifier))
	if want := base64.RawURLEncoding.EncodeToString(sum[:]); challenge != want {
		t.Errorf("challenge = %q, want %q", challenge, want)
	}

	if verifier == challenge {
		t.Error("verifier must not equal the challenge")
	}

	other, _, err := generatePKCE()
	if err != nil {
		t.Fatalf("generatePKCE: %v", err)
	}
	if other == verifier {
		t.Error("consecutive verifiers must differ")
	}
}

func TestRetrieveTokenSendsVerifier(t *testing.T) {
	verifier, _, err := generatePKCE()
	if err != nil {
		t.Fatalf("generatePKCE: %v", err)
	}

	var gotVerifier string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("parse form: %v", err)
		}
		gotVerifier = r.PostForm.Get("code_verifier")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"token","token_type":"Bearer",` +
			`"installation":{"apiEndpoint":"https://example.teamwork.com"}}`))
	}))
	defer server.Close()

	o := NewOAuth2("client", "secret", WithOAuth2Client(server.Client()))
	info := &oauth2ServerInfo{
		TokenEndpoint:                     server.URL,
		TokenEndpointAuthMethodsSupported: []string{"client_secret_post"},
	}
	redirectURL := "http://127.0.0.1/oauth2/callback"
	if err := o.retrieveToken(context.Background(), info, redirectURL, verifier, "code"); err != nil {
		t.Fatalf("retrieveToken: %v", err)
	}

	if gotVerifier != verifier {
		t.Errorf("code_verifier = %q, want %q", gotVerifier, verifier)
	}
}
