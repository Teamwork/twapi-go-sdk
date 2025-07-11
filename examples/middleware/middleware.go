package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
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
	engine := twapi.NewEngine(session,
		twapi.WithMiddleware(func(next twapi.HTTPClient) twapi.HTTPClient {
			return twapi.HTTPClientFunc(func(req *http.Request) (*http.Response, error) {
				// logging middleware

				fmt.Printf("➡️  %s %s", req.Method, req.URL)

				resp, err := next.Do(req)
				switch {
				case err != nil:
					fmt.Printf(" ❌ %s\n", err.Error())
				case resp.StatusCode >= 400:
					fmt.Printf(" ❌ %s\n", resp.Status)
				default:
					fmt.Printf(" ✅ %s\n", resp.Status)
				}
				return resp, err
			})
		}),
	)

	_, err := projects.ProjectList(context.Background(), engine, projects.NewProjectListRequest())
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to list projects: %v\n", err)
		os.Exit(1)
	}
}
