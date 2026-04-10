# AGENTS.md — Teamwork API Go SDK

This file documents patterns and conventions for AI agents contributing to this codebase.

---

## Project Structure

```
twapi-go-sdk/
├── *.go                   # Root package (twapi): Engine, interfaces, shared types
├── projects/              # All API resource implementations (~88 .go files)
├── session/               # Authentication strategies
├── internal/browser/      # OAuth2 browser launcher
└── examples/              # Usage examples
```

- Module: `github.com/teamwork/twapi-go-sdk`
- Go 1.24+, minimal external dependencies (only `golang.org/x/sys`)
- Most work happens in the `projects` package.

---

## Adding a New Resource

Every resource follows the same structure. Use `projects/message.go` and `projects/message_reply.go` as reference.

### 1. File

Create `projects/{resource_name}.go`. Plural resource names use singular form in Go identifiers (e.g., messages → `Message`).

### 2. Interface assertions (top of file)

```go
var (
    _ twapi.HTTPRequester = (*MessageCreateRequest)(nil)
    _ twapi.HTTPResponser = (*MessageCreateResponse)(nil)
    _ twapi.HTTPRequester = (*MessageUpdateRequest)(nil)
    _ twapi.HTTPResponser = (*MessageUpdateResponse)(nil)
    _ twapi.HTTPRequester = (*MessageDeleteRequest)(nil)
    _ twapi.HTTPResponser = (*MessageDeleteResponse)(nil)
    _ twapi.HTTPRequester = (*MessageGetRequest)(nil)
    _ twapi.HTTPResponser = (*MessageGetResponse)(nil)
    _ twapi.HTTPRequester = (*MessageListRequest)(nil)
    _ twapi.HTTPResponser = (*MessageListResponse)(nil)
)
```

### 3. Type naming convention

| Purpose                  | Name pattern                              |
|--------------------------|-------------------------------------------|
| Create request/response  | `MessageCreateRequest` / `...Response`    |
| Update request/response  | `MessageUpdateRequest` / `...Response`    |
| Delete request/response  | `MessageDeleteRequest` / `...Response`    |
| Get request/response     | `MessageGetRequest` / `...Response`       |
| List request/response    | `MessageListRequest` / `...Response`      |
| Path parameters struct   | `MessageCreateRequestPath`, etc.          |
| List filters struct       | `MessageListRequestFilters`               |
| Status/enum constant type| `MessageStatus` (string typedef)          |

### 4. Constructor functions

Provide a `New{Resource}{Action}Request(...)` constructor for each operation, taking required fields as arguments:

```go
func NewMessageCreateRequest(projectID int64, title string, body string) MessageCreateRequest {
    return MessageCreateRequest{
        Path:  MessageCreateRequestPath{ProjectID: projectID},
        Title: title,
        Body:  body,
    }
}
```

### 5. Top-level operation functions

```go
func MessageCreate(ctx context.Context, engine *twapi.Engine, req MessageCreateRequest) (*MessageCreateResponse, error) {
    return twapi.Execute[MessageCreateRequest, *MessageCreateResponse](ctx, engine, req)
}
```

---

## Request / Response Patterns

### Path parameters

Always use a dedicated `{Resource}{Action}RequestPath` struct. Do NOT serialize path params (`json:"-"`):

```go
type MessageCreateRequestPath struct {
    ProjectID int64
}

type MessageCreateRequest struct {
    Path MessageCreateRequestPath `json:"-"`
    Title string `json:"title"`
    Body  string `json:"body"`
}
```

### Optional fields

Use pointers with `omitempty`:

```go
Description *string `json:"description,omitempty"`
Color       *string `json:"color,omitempty"`
Notify      *bool   `json:"notify-current-user,omitempty"`
```

### HTTPRequest implementation (POST/PUT/PATCH)

```go
func (r MessageCreateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
    uri := fmt.Sprintf("%s/projects/%d/posts.json", server, r.Path.ProjectID)

    payload := struct {
        Post MessageCreateRequest `json:"post"`
    }{Post: r}

    var body bytes.Buffer
    if err := json.NewEncoder(&body).Encode(payload); err != nil {
        return nil, fmt.Errorf("failed to encode create message request: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, &body)
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", "application/json")

    return req, nil
}
```

### HTTPRequest implementation (GET with query params)

```go
func (r MessageListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, server+"/projects/api/v3/messages.json", nil)
    if err != nil {
        return nil, err
    }

    query := req.URL.Query()
    if r.Filters.SearchTerm != "" {
        query.Set("searchTerm", r.Filters.SearchTerm)
    }
    if len(r.Filters.ProjectIDs) > 0 {
        ids := make([]string, len(r.Filters.ProjectIDs))
        for i, id := range r.Filters.ProjectIDs {
            ids[i] = strconv.FormatInt(id, 10)
        }
        query.Set("projectIds", strings.Join(ids, ","))
    }
    if r.Filters.Page > 0 {
        query.Set("page", strconv.FormatInt(r.Filters.Page, 10))
    }
    req.URL.RawQuery = query.Encode()

    return req, nil
}
```

### HandleHTTPResponse implementation

```go
func (r *MessageCreateResponse) HandleHTTPResponse(resp *http.Response) error {
    if resp.StatusCode != http.StatusCreated {
        return twapi.NewHTTPError(resp, "failed to create message")
    }
    if err := json.NewDecoder(resp.Body).Decode(r); err != nil {
        return fmt.Errorf("failed to decode create message response: %w", err)
    }
    if r.ID == 0 {
        return fmt.Errorf("create message response does not contain a valid identifier")
    }
    return nil
}
```

Expected status codes: `201 Created` for create, `200 OK` for get/update/list, `204 No Content` for delete.

### Pagination (list responses)

```go
type MessageListResponse struct {
    request  MessageListRequest  // unexported, set by Execute

    Meta struct {
        Page struct {
            HasMore bool `json:"hasMore"`
        } `json:"page"`
    } `json:"meta"`

    Messages []Message `json:"messages"`
}

func (r *MessageListResponse) SetRequest(req MessageListRequest) {
    r.request = req
}

func (r *MessageListResponse) Iterate() *MessageListRequest {
    if !r.Meta.Page.HasMore {
        return nil
    }
    req := r.request
    req.Filters.Page++
    return &req
}
```

---

## Common Types

Defined in the root `twapi` package:

| Type               | Use                                                             |
|--------------------|-----------------------------------------------------------------|
| `LegacyNumber`     | `int64` serialized as a quoted JSON string (`"12345"`)          |
| `LegacyDate`       | `time.Time` serialized as `"20060102"`                          |
| `LegacyUserGroups` | Comma-separated user/team/company IDs (`"123,t456,c789"`)       |
| `UserGroups`       | Modern format with explicit `userIds`/`teamIds`/`companyIds`    |
| `Date`             | `time.Time` formatted as `"2006-01-02"`                         |
| `Time`             | `time.Time` formatted as `"15:04:05"`                           |
| `OptionalDateTime` | `time.Time` that accepts empty strings                          |
| `Money`            | `int64` representing cents; use `NewMoney(float64)`             |
| `HTTPError`        | Structured API error; check with `errors.As(err, &httpErr)`     |

---

## Testing Patterns

### File naming

- Integration tests: `projects/{resource}_test.go` (package `projects_test`)
- Example tests: `projects/{resource}_example_test.go` (package `projects_test`)
- Shared setup: `projects/main_test.go`

### Integration test structure

```go
func TestMessageCreate(t *testing.T) {
    if engine == nil {
        t.Skip("Skipping test because the engine is not initialized")
    }

    tests := []struct {
        name  string
        input projects.MessageCreateRequest
    }{{
        name:  "only required fields",
        input: projects.NewMessageCreateRequest(testResources.ProjectID, "title", "body"),
    }, {
        name: "all fields",
        input: projects.MessageCreateRequest{/* ... */},
    }}

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
            t.Cleanup(cancel)

            response, err := projects.MessageCreate(ctx, engine, tt.input)
            t.Cleanup(func() {
                if err != nil {
                    return
                }
                _, err := projects.MessageDelete(context.Background(), engine,
                    projects.NewMessageDeleteRequest(int64(response.ID)))
                if err != nil {
                    t.Errorf("failed to delete message after test: %s", err)
                }
            })

            if err != nil {
                t.Errorf("unexpected error: %s", err)
            } else if response.ID == 0 {
                t.Error("expected a valid message ID but got 0")
            }
        })
    }
}
```

Key points:
- Skip with `t.Skip(...)` when `engine == nil` (no env vars configured).
- Always clean up created resources in `t.Cleanup`.
- Use `context.Background()` inside `t.Cleanup` (not `t.Context()`, which is already cancelled).
- Shared test fixtures live in `testResources` (set up in `TestMain`).

### Example tests

```go
func ExampleMessageCreate() {
    address, stop, err := startMessageServer()
    if err != nil {
        fmt.Printf("failed to start server: %s", err)
        return
    }
    defer stop()

    ctx := context.Background()
    eng := twapi.NewEngine(session.NewBearerToken("token", "http://"+address))

    response, err := projects.MessageCreate(ctx, eng,
        projects.NewMessageCreateRequest(777, "New Message", "Body text."))
    if err != nil {
        fmt.Printf("failed to create message: %s", err)
    } else {
        fmt.Printf("created message with identifier %d\n", response.ID)
    }

    // Output: created message with identifier 12345
}
```

Each example file has a `start{Resource}Server()` helper that returns `(address string, stop func(), err error)`.

---

## Documentation Conventions

- **Export everything** in request/response structs (PascalCase fields).
- **Comment every exported type and function** with purpose and a link to the API docs.
- **Comment struct fields** that are non-obvious.
- Reference format: `// https://apidocs.teamwork.com/docs/teamwork/...`

---

## JSON Tag Rules

| Scenario                       | Tag                         |
|--------------------------------|-----------------------------|
| Required field                 | `json:"fieldName"`          |
| Optional field                 | `json:"fieldName,omitempty"` |
| Path parameter (never encoded) | `json:"-"`                  |
| Hyphenated API key             | `json:"some-flag,omitempty"` |

---

## Nested Resources

For resources owned by a parent (e.g. message replies belong to messages), the path struct holds the parent ID:

```go
type MessageReplyCreateRequestPath struct {
    MessageID int64
}

// URI construction:
uri := fmt.Sprintf("%s/messages/%d/message-replies.json", server, r.Path.MessageID)
```

---

## Error Handling

- Wrap errors with context: `fmt.Errorf("failed to create message: %w", err)`
- Use `twapi.NewHTTPError(resp, "message")` for non-2xx responses.
- Callers can inspect: `errors.As(err, &twapi.HTTPError{})`.
