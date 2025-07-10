# Teamwork.com API - Go SDK

This is the official Go SDK for the Teamwork.com API.

https://apidocs.teamwork.com/

## 📦 Installing

TO add this library as a dependency of youRun the following command inside your Go module:
```bash
go get github.com/teamwork/twapi-go-sdk
```

## 🔐 Authentication

The library supports the following authentication methods:

* Basic authentication using an API token or user credentials.
* Bearer token authentication.
* OAuth2 authentication using client ID and secret. This will open a browser
  window to authorize the application, so it is not suitable for headless
  environments.

## 🏁 Getting started

```go
package main

import (
  "context"
  "fmt"
  "log"

  twapi "github.com/teamwork/twapi-go-sdk"
  "github.com/teamwork/twapi-go-sdk/projects"
  "github.com/teamwork/twapi-go-sdk/session"
)

func main() {
  ctx := context.Background()
  engine := twapi.NewEngine(session.NewBearerToken("your_token", "https://yourdomain.teamwork.com"))

  project, err := projects.ProjectCreate(ctx, engine, projects.ProjectCreateRequest{
    Name: "New Project",
  })
  if err != nil {
    fmt.Printf("failed to create project: %s", err)
  } else {
    fmt.Printf("created project with identifier %d\n", project.ID)
  }
}
```