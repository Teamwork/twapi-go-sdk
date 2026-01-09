package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
	"github.com/teamwork/twapi-go-sdk/session"
)

func main() {
	clientID := flag.String("client-id", "", "OAuth2 Client ID")
	clientSecret := flag.String("client-secret", "", "OAuth2 Client Secret")
	callbackServerAddr := flag.String("callback-server-addr", "127.0.0.1:6275", "OAuth2 callback server address")
	verifyTLS := flag.Bool("verify-tls", true, "Verify TLS certificates")
	server := flag.String("server", "https://teamwork.com", "Teamwork server URL")
	flag.Parse()

	if *clientID == "" || *clientSecret == "" {
		fmt.Fprintln(os.Stderr, "client-id and client-secret are required")
		os.Exit(1)
	}

	fmt.Printf("📡 Server listening on http://%s/oauth2/callback\n", *callbackServerAddr)
	session := session.NewOAuth2(*clientID, *clientSecret,
		session.WithOAuth2Server(*server),
		session.WithOAuth2CallbackServerAddr(*callbackServerAddr),
		session.WithOAuth2Client(&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: !*verifyTLS,
				},
			},
		}),
	)
	engine := twapi.NewEngine(session)

	project, err := projects.ProjectCreate(context.Background(), engine, projects.ProjectCreateRequest{
		Name: fmt.Sprintf("New Project - %d", time.Now().Nanosecond()),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create project: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("🔐 Bearer token: %s\n", session.BearerToken())
	fmt.Printf("✅ Project created successfully (%d)\n", project.ID)
}
