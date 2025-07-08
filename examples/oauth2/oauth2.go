package main

import (
	"context"
	"flag"
	"fmt"
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
	server := flag.String("server", "https://teamwork.com", "Teamwork server URL")
	flag.Parse()

	if *clientID == "" || *clientSecret == "" {
		fmt.Fprintln(os.Stderr, "client-id and client-secret are required")
		os.Exit(1)
	}

	session := session.NewOAuth2(*clientID, *clientSecret,
		session.WithOAuth2Server(*server),
		session.WithOAuth2CallbackServerAddr(*callbackServerAddr),
	)
	engine := twapi.NewEngine(session)

	project, err := projects.CreateProject(context.Background(), engine, projects.CreateProjectRequest{
		Name: fmt.Sprintf("New Project - %d", time.Now().Nanosecond()),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create project: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Project created successfully (%d)\n", project.ID)
}
