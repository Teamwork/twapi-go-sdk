<div align="center">

# 🚀 Teamwork.com API - Go SDK

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![Go Reference](https://img.shields.io/badge/Go-Reference-00ADD8?style=for-the-badge&logo=go)](https://pkg.go.dev/github.com/Teamwork/twapi-go-sdk)
[![License](https://img.shields.io/badge/License-MIT-blue?style=for-the-badge)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/teamwork/twapi-go-sdk?style=for-the-badge)](https://goreportcard.com/report/github.com/teamwork/twapi-go-sdk)

**The official Go SDK for the Teamwork.com API**

*Build powerful integrations with Teamwork's project management platform*

[📖 API Documentation](https://apidocs.teamwork.com/) • [🎯 Examples](examples/) • [🐛 Report Issues](https://github.com/teamwork/twapi-go-sdk/issues)

</div>

---

## ✨ Features

- � **Multiple Authentication Methods** - Bearer token, Basic auth, and OAuth2
- 🏗️ **Type-Safe API** - Fully typed requests and responses
- 🌐 **Context Support** - Built-in context.Context support for cancellation and timeouts
- 📦 **Zero Dependencies** - Minimal external dependencies
- 🧪 **Thoroughly Tested** - Comprehensive test coverage
- 📱 **Cross-Platform** - Works on Windows, macOS, and Linux

## 📦 Installation

Add this library as a dependency to your Go module:

```bash
go get github.com/teamwork/twapi-go-sdk
```

**Requirements:**
- Go 1.24 or later
- A Teamwork.com account with API access

## 🔐 Authentication

The SDK supports multiple authentication methods to suit different use cases:

### 🎫 Bearer Token (Recommended)
Perfect for server-to-server integrations and scripts:

```go
import "github.com/teamwork/twapi-go-sdk/session"

session := session.NewBearerToken("your_api_token", "https://yourdomain.teamwork.com")
```

### 🔑 Basic Authentication
Use with API tokens or user credentials:

```go
// With API token
session := session.NewBasicAuth("your_api_token", "", "https://yourdomain.teamwork.com")

// With username/password
session := session.NewBasicAuth("username", "password", "https://yourdomain.teamwork.com")
```

### 🌐 OAuth2
Ideal for user-facing applications (opens browser for authorization):

```go
session := session.NewOAuth2("client_id", "client_secret",
  session.WithOAuth2Server("https://teamwork.com"),
  session.WithOAuth2CallbackServerAddr("127.0.0.1:6275"),
)
```

> [!CAUTION]
> ⚠️ **Note:** OAuth2 opens a browser window and is not suitable for headless environments.

## 🏁 Quick Start

Here's a simple example to get you started:

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
  
  // Initialize the SDK with bearer token authentication
  engine := twapi.NewEngine(session.NewBearerToken("your_token", "https://yourdomain.teamwork.com"))

  // Create a new project
  project, err := projects.ProjectCreate(ctx, engine, projects.NewProjectCreateRequest("My Awesome Project"))
  if err != nil {
    log.Fatalf("Failed to create project: %v", err)
  }
  
  fmt.Printf("✅ Created project '%s' with ID: %d\n", project.Name, project.ID)
}
```

## 📚 Examples

### Working with Projects

```go
package main

import (
  "context"
  "fmt"
  "time"

  twapi "github.com/teamwork/twapi-go-sdk"
  "github.com/teamwork/twapi-go-sdk/projects"
  "github.com/teamwork/twapi-go-sdk/session"
)

func main() {
  ctx := context.Background()
  engine := twapi.NewEngine(session.NewBearerToken("your_token", "https://yourdomain.teamwork.com"))

  project, err := projects.ProjectCreate(ctx, engine, projects.ProjectCreateRequest{
    Name:        "Q1 Marketing Campaign",
    Description: twapi.Ptr("Marketing campaign for Q1 product launch"),
    StartAt:     twapi.Ptr(time.Now()),
    EndAt:       twapi.Ptr(time.Now().AddDate(0, 3, 0)), // 3 months from now
  })
  if err != nil {
    fmt.Fprintf(os.Stderr, "❌ Failed to create project: %v\n", err)
    os.Exit(1)
  }

  // Retrieve the project
  retrievedProject, err := projects.ProjectGet(ctx, engine, projects.NewProjectRetrieveRequest(int64(project.ID)))
  if err != nil {
    fmt.Fprintf(os.Stderr, "❌ Failed to retrieve project: %v\n", err)
    os.Exit(1)
  }

  fmt.Printf("✅ Project: %s (ID: %d)\n", retrievedProject.Name, retrievedProject.ID)
  
  // List all projects
  projectsList, err := projects.ProjectList(ctx, engine, projects.NewProjectRetrieveManyRequest())
  if err != nil {
    fmt.Fprintf(os.Stderr, "❌ Failed to list projects: %v\n", err)
    os.Exit(1)
  }
  
  fmt.Printf("✅ Found %d projects\n", len(projectsList.Projects))
  
  // Update the project
  updatedProject, err := projects.ProjectUpdate(ctx, engine, projects.ProjectUpdateRequest{
    Path:  projects.ProjectUpdateRequestPath{
      ID: int64(project.ID),
    },
    Name: "Q1 Marketing Campaign - Updated",
  })
  if err != nil {
    fmt.Fprintf(os.Stderr, "❌ Failed to update project: %v\n", err)
    os.Exit(1)
  }
  
  fmt.Printf("✅ Updated project name to: %s\n", updatedProject.Name)

  // Delete the project
  err = projects.ProjectDelete(ctx, engine, projects.NewProjectDeleteRequest(int64(project.ID)))
  if err != nil {
    fmt.Fprintf(os.Stderr, "❌ Failed to delete project: %v\n", err)
    os.Exit(1)
  }

  fmt.Println("✅ Project deleted successfully")
}
```

### OAuth2 Authentication Example

```go
package main

import (
  "context"
  "flag"
  "fmt"
  "os"

  twapi "github.com/teamwork/twapi-go-sdk"
  "github.com/teamwork/twapi-go-sdk/projects"
  "github.com/teamwork/twapi-go-sdk/session"
)

func main() {
  clientID := flag.String("client-id", "", "OAuth2 Client ID")
  clientSecret := flag.String("client-secret", "", "OAuth2 Client Secret")
  flag.Parse()

  if *clientID == "" || *clientSecret == "" {
    fmt.Fprintln(os.Stderr, "❌ client-id and client-secret are required")
    os.Exit(1)
  }

  // Create OAuth2 session (will open browser for authorization)
  session := session.NewOAuth2(*clientID, *clientSecret,
    session.WithOAuth2CallbackServerAddr("127.0.0.1:6275"),
  )
  
  engine := twapi.NewEngine(session)

  // Test the connection by creating a project
  project, err := projects.ProjectCreate(context.Background(), engine, projects.NewProjectCreateRequest("OAuth2 Test Project"))
  if err != nil {
    fmt.Fprintf(os.Stderr, "❌ Failed to create project: %v\n", err)
    os.Exit(1)
  }

  fmt.Printf("✅ OAuth2 authentication successful! Created project: %s (ID: %d)\n", project.Name, project.ID)
}
```

### Error Handling Best Practices

```go
package main

import (
  "context"
  "errors"
  "fmt"
  "net/http"

  twapi "github.com/teamwork/twapi-go-sdk"
  "github.com/teamwork/twapi-go-sdk/projects"
  "github.com/teamwork/twapi-go-sdk/session"
)

func main() {
  ctx := context.Background()
  engine := twapi.NewEngine(session.NewBearerToken("your_token", "https://yourdomain.teamwork.com"))

  project, err := projects.ProjectCreate(ctx, engine, projects.NewProjectCreateRequest("Test Project"))
  if err != nil {
    // Handle different types of errors
    var httpErr *twapi.HTTPError
    if errors.As(err, &httpErr) {
      switch httpErr.StatusCode {
      case http.StatusUnauthorized:
        fmt.Println("❌ Authentication failed - check your API token")
      case http.StatusForbidden:
        fmt.Println("❌ Access denied - insufficient permissions")
      case http.StatusTooManyRequests:
        fmt.Println("❌ Rate limit exceeded - please retry later")
      default:
        fmt.Printf("❌ HTTP error %d: %s\n", httpErr.StatusCode, httpErr.Message)
      }
    } else {
      fmt.Printf("❌ Unexpected error: %v\n", err)
    }
    return
  }

  fmt.Printf("✅ Success! Created project: %s\n", project.Name)
}
```

## 🔧 Configuration

### Context and Timeouts

The SDK supports Go's `context.Context` for request cancellation and timeouts:

```go
import "time"

// Create a context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Use the context in API calls
project, err := projects.ProjectCreate(ctx, engine, request)
```

### Custom HTTP Client

You can customize the underlying HTTP client:

```go
import (
  "net/http"
  "time"
)

// Create engine with custom HTTP client
httpClient := &http.Client{
  Timeout: 60 * time.Second,
  Transport: &http.Transport{
    MaxIdleConns:        10,
    IdleConnTimeout:     30 * time.Second,
    DisableCompression:  true,
  },
}

engine := twapi.NewEngine(session,
  twapi.WithHTTPClient(httpClient),
)
```

### Middleware

You can add custom middleware to intercept and modify HTTP requests/responses. Middlewares are executed in the order they are added:

```go
import (
  "fmt"
  "net/http"
  "time"

  twapi "github.com/teamwork/twapi-go-sdk"
  "github.com/teamwork/twapi-go-sdk/session"
)

// Logging middleware
func loggingMiddleware(next twapi.HTTPClient) twapi.HTTPClient {
  return twapi.HTTPClientFunc(func(req *http.Request) (*http.Response, error) {
    start := time.Now()
    fmt.Printf("➡️  %s %s", req.Method, req.URL)

    resp, err := next.Do(req)
    duration := time.Since(start)

    switch {
    case err != nil:
      fmt.Printf(" ❌ %s (took %v)\n", err.Error(), duration)
    case resp.StatusCode >= 400:
      fmt.Printf(" ❌ %s (took %v)\n", resp.Status, duration)
    default:
      fmt.Printf(" ✅ %s (took %v)\n", resp.Status, duration)
    }
    return resp, err
  })
}

// Rate limiting middleware
func rateLimitingMiddleware(next twapi.HTTPClient) twapi.HTTPClient {
  return twapi.HTTPClientFunc(func(req *http.Request) (*http.Response, error) {
    // Add rate limiting logic here
    time.Sleep(100 * time.Millisecond) // Simple delay example
    return next.Do(req)
  })
}

// Authentication header middleware
func authHeaderMiddleware(apiKey string) func(twapi.HTTPClient) twapi.HTTPClient {
  return func(next twapi.HTTPClient) twapi.HTTPClient {
    return twapi.HTTPClientFunc(func(req *http.Request) (*http.Response, error) {
      req.Header.Set("X-Custom-Auth", apiKey)
      return next.Do(req)
    })
  }
}

func main() {
  session := session.NewBearerToken("your_token", "https://yourdomain.teamwork.com")

  // Chain multiple middlewares
  engine := twapi.NewEngine(session,
    twapi.WithMiddleware(loggingMiddleware),
    twapi.WithMiddleware(rateLimitingMiddleware),
    twapi.WithMiddleware(authHeaderMiddleware("custom-key")),
  )

  // Now all requests will go through the middleware chain
  // ...use engine for API calls...
}
```

### Iterator for Paginated Results

The SDK provides an iterator function to easily handle paginated API responses:

```go
import (
  "context"
  "fmt"

  twapi "github.com/teamwork/twapi-go-sdk"
  "github.com/teamwork/twapi-go-sdk/projects"
  "github.com/teamwork/twapi-go-sdk/session"
)

func main() {
  ctx := context.Background()
  engine := twapi.NewEngine(session.NewBearerToken("your_token", "https://yourdomain.teamwork.com"))

  // Create an iterator for paginated project results
  next, err := twapi.Iterate[projects.ProjectListRequest, *projects.ProjectListResponse](
    ctx,
    engine,
    projects.NewProjectListRequest(),
  )
  if err != nil {
    fmt.Printf("Failed to create iterator: %v\n", err)
    return
  }

  // Iterate through all pages
  var iteration int
  for {
    iteration++
    fmt.Printf("📄 Page %d\n", iteration)

    response, hasNext, err := next()
    if err != nil {
      fmt.Printf("Error fetching page: %v\n", err)
      break
    }
    if response == nil {
      break
    }

    // Process projects from current page
    for _, project := range response.Projects {
      fmt.Printf("  ➢ %s (ID: %d)\n", project.Name, project.ID)
    }

    // Check if there are more pages
    if !hasNext {
      break
    }
  }
}
```

## 🐛 Error Handling

The SDK provides structured error handling:

```go
import "errors"

project, err := projects.ProjectCreate(ctx, engine, request)
if err != nil {
  var httpErr *twapi.HTTPError
  if errors.As(err, &httpErr) {
    fmt.Printf("HTTP %d: %s\n", httpErr.StatusCode, httpErr.Message)
    // Handle specific status codes
  }
}
```

## 🧪 Testing

Run the test suite:

```bash
go test ./...
```

Run integration tests:

```bash
TWAPI_SERVER=https://yourdomain.teamwork.com/ TWAPI_TOKEN=your_api_token go test ./...
```

Run tests with coverage:

```bash
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📋 Requirements

- **Go Version:** 1.24 or later
- **Dependencies:** Minimal external dependencies (see `go.mod`)
- **Teamwork Account:** Valid Teamwork.com account with API access

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- 📖 [API Documentation](https://apidocs.teamwork.com/)
- 🐛 [Report Issues](https://github.com/teamwork/twapi-go-sdk/issues)
- 💬 [Community Support](https://teamwork.com/support)

---

<div align="center">

**Made with ❤️ by the Teamwork.com team**

⭐ Star us on GitHub if this project helped you!

</div>