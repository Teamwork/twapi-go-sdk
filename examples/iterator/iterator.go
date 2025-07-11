package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
	"github.com/teamwork/twapi-go-sdk/session"
)

func main() {
	bearerToken := flag.String("bearer-token", "", "OAuth2 Bearer Token")
	server := flag.String("server", "https://teamwork.com", "Teamwork server URL")
	flag.Parse()

	if *bearerToken == "" {
		fmt.Fprintln(os.Stderr, "bearer-token is required")
		os.Exit(1)
	}

	session := session.NewBearerToken(*bearerToken, *server)
	engine := twapi.NewEngine(session)

	next, err := twapi.Iterate[projects.ProjectListRequest, *projects.ProjectListResponse](
		context.Background(),
		engine,
		projects.NewProjectListRequest(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create iterator: %v\n", err)
		os.Exit(1)
	}

	var iteration int
	for {
		iteration++
		fmt.Printf("🔍 Iteration %d\n", iteration)

		response, hasNext, err := next()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to list projects: %v\n", err)
			os.Exit(1)
		}
		if response == nil {
			break
		}
		for _, project := range response.Projects {
			fmt.Printf("  ➢ Project: %s (%d)\n", project.Name, project.ID)
		}
		if !hasNext {
			break
		}
	}
}
