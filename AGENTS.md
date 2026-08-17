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

### 6. Sparse-fieldsets wiring (List and Get endpoints)

If your resource has a List endpoint, follow the [Sparse fieldsets (list responses)](#sparse-fieldsets-list-responses) recipe below to expose typed field selection. The pattern is mandatory for new v3 list endpoints — `go generate` plus a CI check in `.github/workflows/test.yaml` enforce that the wiring stays in sync.

Then wire the Get endpoint the same way via [Sparse fieldsets (get responses)](#sparse-fieldsets-get-responses), after checking the server actually honours the selection on that route — not every singular handler does.

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

### Typed enums over raw strings

Any field whose value is restricted to a fixed set the API accepts — sort keys (`OrderBy`), include/sideload lists, status values, type discriminators, etc. — **must** be exposed as a named `string` typedef with one exported constant per allowed value. Never expose those fields as `string` or `[]string`.

```go
// CustomItemSideload identifies the related entities that can be requested
// alongside a custom item type via the API's include mechanism.
type CustomItemSideload string

// Supported custom item sideloads.
const (
    CustomItemSideloadCreatedBy        CustomItemSideload = "createdBy"
    CustomItemSideloadCustomItemFields CustomItemSideload = "customItemFields"
    // ...
)

// CustomItemOrderBy identifies the attributes a custom item list can be
// ordered by.
type CustomItemOrderBy string

const (
    CustomItemOrderByName CustomItemOrderBy = "name"
)

type CustomItemListRequestFilters struct {
    OrderBy CustomItemOrderBy
    Include []CustomItemSideload
    // ...
}
```

In the filter's `apply()`, cast back to `string` when writing to the query — and for slice enums, convert via a small loop before `strings.Join`:

```go
if c.OrderBy != "" {
    query.Set("orderBy", string(c.OrderBy))
}
if len(c.Include) > 0 {
    includes := make([]string, len(c.Include))
    for i, include := range c.Include {
        includes[i] = string(include)
    }
    query.Set("include", strings.Join(includes, ","))
}
```

Why: surfaces typos at compile time, makes the legal value set self-documenting (and discoverable via `go doc`), and lets us evolve the API by adding constants without breaking callers. The naming convention is `{Resource}{Purpose}` for the type (e.g., `TaskRequestSideload`, `CustomItemRecordOrderBy`) and `{Type}{Value}` for each constant.

### `skipCounts` on v3 list endpoints

Always hard-code `skipCounts=true` in the filter's `apply()` rather than exposing it as an option. Total counts are derivable from `(page * pageSize) + 1` and `HasMore`, so leaving the option open just lets callers opt into a slower API call by mistake. The list response's `Meta` should mirror the pattern in [Pagination (list responses)](#pagination-list-responses) and expose only `HasMore` — never `Offset`, `Size`, or `Count`.

### v1 and v3 are not interchangeable for the same operation

Several resources are reachable through both API versions, and the newer route is not always the one that does more. `comment.go`, `link.go` and `message.go` already write through v1 while reading through v3, because that is where the working endpoint is.

Moving a task between tasklists is reachable three ways, and **all three carry the task's subtasks**, so "does it cascade?" is not what separates them:

- `PUT /projects/api/v3/tasks/{id}.json` (`TaskUpdateRequest.TasklistID`). Clears the moved task's own parent link, since a subtask may not live outside its parent's tasklist. A `parentTaskId` sent alongside must already be in the destination, or the call fails. Until an API fix in 2026-08 it failed for *any* subtask, sent parent or not.
- `PUT /tasks/{id}.json`, the v1 generic edit, envelope `{"task":…}` or `{"todo-item":…}`. Clears the parent link too, but silently ignores a `parentTaskId` repeating the current parent instead of failing.
- `PUT /tasks/{id}/move.json` (`TaskMoveRequest`), purpose-built, parameters at the top level of the body. Preserves the parent link when asked, and is the only one that moves a task to a tasklist in another project — v3 answers 422 for that. It also takes a board-column parameter, deliberately unmodelled: columns are superseded by workflows.

Rules that follow:

1. v1 and v3 routes for the "same" operation are separate implementations, not one behind two doors. A finding on one says nothing about the other; verify both.
2. When a write appears in more than one place, look for the purpose-built route before extending the generic one. A `move` or `complete` endpoint usually exists and does more.
3. Do not trust an `affected*Ids` field to be a manifest of what changed. On the v1 move, `affectedTaskIds` repeats entries and covers the moved task and its direct children only — a grandchild moves without appearing. On the v3 update it is the subtasks that moved, but only two levels deep. Both cascade to the whole subtree regardless. `affectedTaskListIds` and `affectedProjectIds` are the source and destination, and are reliable.
4. Legacy payloads spell things differently and the difference is load-bearing: `taskListId` versus v3's `tasklistId`, IDs returned as a comma-separated string (`LegacyNumericList`) rather than a JSON array, and `0` rather than `null` as the value that clears a relationship (`TaskDetachFromParent`). Kebab and camel keys are interchangeable (`todo-item`, `todoItem`), but snake case is not.
5. A live call confirms what happened once. It does not show which parameter was ignored, or which of several plausible causes produced an error — so treat one successful response as weak evidence for how an endpoint behaves in general.

### File uploads are two requests, and only the first reaches the API

Uploading a file is a pre-signed flow (`pending_file.go`):

1. `GET /projects/api/v1/pendingfiles/presignedurl.json?fileName=…&fileSize=…` → `{"ref","url"}`, 200. The reference exists from here on; the URL expires after ten minutes.
2. `PUT` the bytes to that URL. It addresses the storage service, not the API.

`POST /projects/api/v1/pendingfiles.json` accepts a multipart body and does the same job, but the file passes through the API instead of going straight to storage, so the pre-signed route is the one to use.

What follows from that:

- **The upload must not be authenticated with the Teamwork session.** The URL carries its own credentials in its query, and a request that also carries an `Authorization` header is rejected as two auth mechanisms. This is why step 2 cannot go through `twapi.Execute`, and why `Engine.Do` exists — it sends an already built request through the caller's client and middlewares without authenticating it or deriving its URL from the session.
- **Which headers to send is decided by the URL, not by the environment.** `X-Amz-SignedHeaders` lists the headers the signature covers. One of them missing, or an unsigned `x-amz-*` header added, and the upload fails. The API signs a canned ACL unless the installation's bucket sets one of its own, so read the parameter and repeat `X-Amz-Acl: public-read` only when it is listed there. Three internal clients (`twe2e`, `projectsapitesting`, `deskapi`) each hardcode a different guess at the environment for this instead; do not copy them.
- **Set the content length.** A body the standard library cannot measure is sent chunked, which the storage service rejects, so `Size` is required and is assigned to `http.Request.ContentLength`.
- The media type the `PUT` declares is what the stored file keeps — the later copy to the permanent bucket preserves the temporary object's metadata — so it is derived from the file name, with an override for callers who know better.
- The composite operation (`PendingFileCreate`) is a plain function over both steps, so its request and response types deliberately do not implement `twapi.HTTPRequester`/`HTTPResponser`. The individual steps stay exported (`PendingFilePresignedURL`, `PendingFileUpload`) for callers that stream a large file or hand the URL to a browser.

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

### Sparse fieldsets (list responses)

Every v3 list endpoint supports sparse fieldsets — clients restrict the attributes returned per entity via `fields[entity]=attr1,attr2,...`. The SDK exposes this as compile-time-checked typed slices, generated by `internal/sparsefieldsgen` from markers on the source. Contributors write four small bits of wiring; the generator emits everything else (`<Entity>Field` constants, the `<Resource>ListFields` container, its `apply` method, and tests).

Most single-entity endpoints support it too, through the same machinery — see [Sparse fieldsets (get responses)](#sparse-fieldsets-get-responses).

#### The recipe — four steps

1. **Mark the entity struct** with `// sparsefields:gen` as the last paragraph of its doc comment:

   ```go
   // Message is a project-scoped announcement…
   //
   // sparsefields:gen
   type Message struct { ... }
   ```

2. **Mark the `*ListResponse` struct** with `// sparsefields:list`:

   ```go
   // MessageListResponse contains messages matching the filters…
   //
   // sparsefields:list
   type MessageListResponse struct { ... }
   ```

3. **Add `Fields <Resource>ListFields`** as the last field of `*ListRequestFilters`. Document it consistently with other filters:

   ```go
   type MessageListRequestFilters struct {
       // ...existing fields...

       Page     int64
       PageSize int64

       // Fields restricts the attributes returned for the message and each of
       // its sideloads. Each slot of MessageListFields is a separate
       // `fields[entity]=…` selection; populated slots restrict the response,
       // empty slots return the API default. Use the generated MessageField
       // constants to ensure values match real attributes.
       Fields MessageListFields
   }
   ```

4. **Call `t.Fields.apply(query)`** immediately before `req.URL.RawQuery = query.Encode()` in the filter's `apply()` method:

   ```go
   func (m MessageListRequestFilters) apply(req *http.Request) {
       query := req.URL.Query()
       // ...existing filter wiring...
       m.Fields.apply(query)
       req.URL.RawQuery = query.Encode()
   }
   ```

Then run `go generate ./...` from the package directory (or the repo root). The output lands in `projects/sparse_fields_gen.go` plus `projects/sparse_fields_gen_test.go`.

#### What the generator emits

- `type <Entity>Field string` plus one constant per JSON-tagged attribute of the entity (same-package embedded structs are flattened; an outer struct's tag shadows the embed's).
- `type <Resource>ListFields struct { ... }` with one typed slice per slot — the main list slice **and** each map field inside the response's `Included` struct. Slot names mirror the response's Go field names; entity keys come from the response's JSON tags unless overridden with `sparsefields:key` (see below).
- `func (f <Resource>ListFields) apply(query url.Values)` writing every populated slot via `twapi.ApplySparseFields`.
- A pair of generated tests per wired list: `Test<Resource>ListFieldsApply` and `Test<Resource>ListFieldsZeroValue`.

#### When the envelope key isn't the entity name

Entity keys default to the JSON tag of the response field, which is correct whenever the response envelope key matches the entity name the API recognises — the usual case (`{"tasks":[…]}` ↔ `fields[tasks]`). A few endpoints diverge: `/projects/api/v3/projects/budgets.json` wraps its payload in `budgets` while the entity is `projectBudgets`. Declare the divergence with a `sparsefields:key=<entityName>` marker on the response field — as a doc comment or a trailing line comment:

```go
type ProjectBudgetListResponse struct {
    // The endpoint wraps its payload in "budgets", but the entity name the API
    // recognises for sparse fieldsets and sideloads is "projectBudgets", so the
    // fields[...] key can't be derived from the json tag here.
    //
    // sparsefields:key=projectBudgets
    Budgets []ProjectBudget `json:"budgets"`
```

The marker only changes the emitted `fields[...]` key; the JSON tag still drives decoding. It works on `Included` sideload fields too. When in doubt, check the endpoint's apidocs page for the `fields[…]` parameter name rather than assuming it matches the envelope — an unrecognised key is silently ignored by the API, so the whole selection degrades to "return everything" with no error.

#### When *not* to mark a response

- **Responses that wrap anonymous structs** (e.g. some `rates.go` list responses use `[]struct{...}` for their main slice, or `any` for sideloads). The generator can't introspect anonymous types; leave the response unmarked.
- **List endpoints with no filters struct** (e.g. `IndustryListRequest struct{}`). Without a filter to host the `Fields` slot there's no place for the wiring; leave the response unmarked.

In both cases the SDK still works — the endpoint simply can't expose typed sparse fields.

#### Safety nets

- The generator fails fast if a marked response references an entity type that *isn't* marked with `sparsefields:gen` — every slot must resolve to a `<Entity>Field` enum.
- The generator also fails if a filter (or a get request) declares `Fields <Container>` but no method on that type ever calls `<receiver>.Fields.apply(...)` — that catches the easy mistake of adding the slot but forgetting to wire it into the request.
- A `sparsefields:key` marker with no value, or one containing characters that can't appear in an entity name (` `, `,`, `[`, `]`, `=`), is a generate-time error rather than a bad query parameter.
- Two slots of the same response resolving to the same `fields[...]` key is a generate-time error — otherwise one would silently overwrite the other in `apply`.
- CI runs `go generate ./...` on a clean checkout and fails if anything changes, so stale generated code can't merge.

### Sparse fieldsets (get responses)

Single-entity v3 endpoints share their query bindings with the plural endpoint, so most of them accept the same `fields[entity]=…` selection. Marking a `*GetResponse` with `// sparsefields:get` emits a `<Resource>GetFields` container built exactly like the list one — one slot for the retrieved entity, one per map inside `Included`.

Two things differ from the list recipe:

1. **The get request hosts the `Fields` slot**, because there is no filters struct. Add it as the last field of `<Resource>GetRequest` and apply it in `HTTPRequest`:

   ```go
   type MessageGetRequest struct {
       Path MessageGetRequestPath

       // Fields restricts the attributes returned for the message. …
       Fields MessageGetFields
   }

   func (m MessageGetRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
       // ...build req...
       query := req.URL.Query()
       m.Fields.apply(query)
       req.URL.RawQuery = query.Encode()
       return req, nil
   }
   ```

2. **The main slot's entity key is inherited from the entity's list, not from the JSON tag.** The envelope is singular (`{"message": …}`) but the parameter is plural (`fields[messages]`), so deriving the key from the tag would produce a parameter the API ignores. The generator instead reuses the key of the entity's marked `*ListResponse`, which fails at generate time if that list doesn't exist. Sideload keys in `Included` still come from their JSON tags, same as for lists.

   `sparsefields:key` overrides the inherited key when a singular endpoint genuinely diverges. `UserGetResponse` is the one such case today: the server maps `fields[people]` to the list envelope only, and the singular payload is filtered by `fields[person]`.

#### Verifying a get endpoint before wiring it

Sparse fieldsets are not universal on single-entity routes, and an unrecognised or unsupported parameter is silently ignored, so confirm against the server (`projectsapigo`) rather than assuming symmetry with the list:

- The handler must pass the parsed fields to `jsonfilter.Marshal(response, fields)`. Several singular handlers pass `nil` and therefore ignore any selection — `getCustomField`, `getCustomFieldValue` and `getTag` all do, which is why `customfield.go`, `custom_field_value.go` and `tag.go` have no `sparsefields:get` marker.
- The key must reach the response's envelope. `jsonfilter` lowercases both sides, so casing doesn't matter, but the word must match: `arg.SparseFieldsSkills` (`fields[skill]`) maps only to `skills`, so it can't filter the skill get's `{"skill": …}` payload — `skill.go` is unmarked for that reason.
- v1 endpoints (`link.go`, `team.go` gets) have no sparse-fieldset support at all.

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
| `NullableInt64`    | Tri-state int (unset / null / value); use `NewNullableInt64` / `NullInt64` |
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
