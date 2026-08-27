# AGENTS.md — Teamwork API Go SDK

This file documents patterns and conventions for AI agents contributing to this codebase.

---

## This Repository Is Public

**Everything written here is published.** Source, comments, generated files, test
fixtures, commit messages, PR titles and descriptions, review comments, issue
text — all of it is world-readable and permanent, since a rewritten history does
not remove what has already been cloned, cached, or indexed. This file is public
too, and is bound by its own rules.

Research done to write a change is *not* automatically publishable. Internal
sources may be used to get the behaviour right; only the resulting public-facing
behaviour goes into the repository.

### Never write

- **Internal system names.** Private repositories, internal services,
  deployments, hostnames, dashboards, monitoring or CI tooling, chat channels,
  wikis, runbooks, or the environments they run in.
- **Server-side implementation identifiers.** Handler, function, package, type,
  struct-field or query-argument names read from private backend code. The SDK
  describes an HTTP API: name routes, parameters and JSON keys, never the code
  behind them.
- **Internal tracking references.** Ticket keys, issue numbers, incident IDs,
  PR links, or commit hashes belonging to private repositories.
- **Unreleased or embargoed information.** Roadmap items, planned deprecations,
  dated fixes ("fixed in the 2026-08 release"), feature flags, capacity or
  traffic figures, or any endpoint not yet documented publicly.
- **Server defects described as such.** Do not explain a workaround by narrating
  a backend bug, its root cause, or when it will be fixed — that is a live
  exploitation hint for a service that is still running the old code. Document
  what the endpoint does today, from the outside.
- **Secrets and real data.** Tokens, API keys, passwords, cookies, real
  installation subdomains, real customer or employee names, emails, or IDs
  copied from a live account — including inside example output, test fixtures,
  and error strings.

### Instead

State only what a caller of the public API can observe, and cite public
documentation (`https://apidocs.teamwork.com/...`) as the source. Where a
private source was needed, credit it generically.

| Instead of                                                | Write                                                          |
|-----------------------------------------------------------|----------------------------------------------------------------|
| "confirm against `internal-api-repo`"                     | "confirm against the API's public documentation, or a live call" |
| "the `getFoo` handler drops the parsed fields"            | "this endpoint ignores the `fields[...]` selection"            |
| "broken until the 2026-08 fix"                            | "returns 422 in this case"                                     |
| "see PROJ-1234"                                           | omit, or describe the behaviour the change addresses           |
| "tested against `acme.teamwork.com` with token `tw_...`"  | "`example.teamwork.com`", `"token"`                            |

Fixtures follow the existing placeholders: server `http://` + a local test
address, token `"token"`, IDs like `777` and `12345`.

### Commit messages and PR descriptions

Same rules, plus: keep them to what the change does and why, in terms of the
public API. No internal links, no ticket keys, no "as discussed in
<internal channel>", no private reviewer names, no attribution to internal
services as the source of truth. A reader outside the company must be able to
understand the message completely.

### Before committing

Re-read the diff as an outsider would and ask whether any line only makes sense
to someone with access to private systems. If a comment cannot be rewritten
without an internal reference, drop the comment — the code ships, the reference
does not.

---

## Project Structure

```
twapi-go-sdk/
├── *.go                   # Root package (twapi): Engine, interfaces, shared types
├── projects/              # All API resource implementations (~88 .go files)
├── session/               # Authentication strategies
├── internal/browser/      # OAuth2 browser launcher
├── cmd/next-version/      # Computes the release version from the changes
└── examples/              # Usage examples
```

- Module: `github.com/teamwork/twapi-go-sdk`
- Go 1.24+, minimal external dependencies (only `golang.org/x/sys`)
- Most work happens in the `projects` package.

---

## Releases and version bumps

- A pull-request title needs one of the prefixes CONTRIBUTING.md lists, and it is load-bearing twice over: the `PR lint` workflow rejects a title without one, and `cmd/next-version` reads it to pick the release bump (`Feature:` → minor, everything else → patch). The *title* is what counts, not the commits: `cmd/next-version` resolves each commit to its merged pull request and folds a rebase-merged branch into one change, so an unprefixed `Address review feedback` inside a `Feature:` pull request is harmless while an unprefixed title silently costs the release its minor bump — which is how v1.21.6 shipped #120 under a patch tag — a `feat:` whose body declared a `BREAKING CHANGE` renaming six exported identifiers, so the release was two levels short.
- **A breaking change cannot be a major here.** `capMajor` reports the major and tags a minor, because v2 is reached by moving the module path to `/v2`, not by tagging `v2.0.0` — that tag on this path is one `go get` will not resolve. Both reports say so loudly, since consumers get the change on `go get -u`. `-bump=major` exists for a deliberate module-path migration; do not reach for it to "be correct" about semver. `TestBreakingChangeIsCappedAtMinor` pins this, and it is the one substantive difference from the copy of this tool in the MCP server repository — keep the divergence deliberate.
- `.github/workflows/pr_lint.yaml` enforces the prefix by running `go run ./cmd/next-version -check-title="$PR_TITLE"`, not a third-party action: the vocabulary is `bumpByPrefix` itself, so the door and the release cannot disagree, and no entry has to be added to the organisation's Actions allow-list. Two consequences to preserve: the title reaches the script through `env:`, never `${{ }}` interpolation — it is attacker-controlled text on a fork's pull request — and `checkTitle` validates the prefix against `bumpByPrefix` directly rather than through `classify`, which reports a major for *any* prefix carrying a `!` so a breaking change is never under-versioned. That leniency is right when reading history and wrong at the door: `Whatever!: …` must not pass.
- Releasing is `workflow_dispatch` on `.github/workflows/release.yaml` with `bump` (auto/patch/minor/major) and `dry_run` inputs; pushing a `v*.*.*` tag by hand still works and skips the computation. The two entry points meet in the `version` job, whose outputs (`version`, `previous_tag`, `sha`, `release`) the release job reads — do not reintroduce a `GITHUB_REF`-derived version, and check out `needs.version.outputs.sha` rather than `github.ref`, which on a dispatch is a branch. The dispatch path cannot create the tag and let the push path take over: a tag created with `GITHUB_TOKEN` does not trigger `on: push: tags`, so it would tag and release nothing.
- The prefixes CONTRIBUTING.md documents are guarded by `TestEveryContributingPrefixIsMapped`, which parses the `` - `Xxx:` for `` list and asserts each one is both mapped and accepted by the check — keep that list's shape. `.github/dependabot.yml` pins `commit-message.prefix` so its titles keep passing; without it Dependabot infers the prefix from recent history.

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

### 7. Count wiring (List endpoints)

Use `twapi.ListMeta` for the response's `Meta`, and give the filters a `CountMode twapi.ListCountMode` slot wired through `Apply` and `ResolveCount`. Nothing generates this, so see [Total counts](#total-counts-skipcounts-on-v3-list-endpoints) for the four steps and, importantly, for why `ResolveCount` is not optional.

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

### Ordering (`orderBy` / `orderMode`)

A list filter exposes ordering as two fields, placed together and documented with
the endpoint's default:

```go
// OrderBy is the field to sort the results by. Use the MessageOrderBy
// constants. The endpoint defaults to createdat.
OrderBy MessageOrderBy

// OrderMode is the direction to sort the results in. See twapi.OrderMode for
// the supported values. The endpoint defaults to ascending.
OrderMode twapi.OrderMode
```

Rules:

1. **`OrderBy` is a per-resource `{Resource}OrderBy` typedef**, one constant per
   value the endpoint's apidocs page lists under "Allowed values". Never a bare
   `string`.
2. **`OrderMode` is always `twapi.OrderMode`.** Ascending and descending are the
   only directions the API takes, so the shared type is the whole guard — there
   is no per-resource order-mode enum and no runtime validation to add.
3. **Only wire what the endpoint documents.** An unrecognised query key is
   silently ignored, so a `Fields`-style field on an endpoint that does not
   support ordering looks functional and does nothing. Some endpoints document
   `orderMode` but no `orderBy` (`jobrole.go`, `skill.go`, `custom_item_field.go`)
   — those get `OrderMode` alone. Several document neither (`project_category.go`,
   `timer.go`, `workflow.go`, the owner-scoped routes in `custom_field_value.go`,
   the v1 `link.go` list) and are deliberately left without ordering.
4. **Where `orderBy` accepts `customfield`, carry its companion ID.** Sorting by
   a custom field also needs `orderByCustomFieldId` (`orderByFieldId` on custom
   item records), so those filters expose `OrderByCustomFieldID` /
   `OrderByFieldID` next to `OrderBy`.
5. **v1 routes spell it differently.** The team list takes `sortBy`/`sortOrder`;
   the filter still exposes `OrderBy`/`OrderMode` for consistency with the rest
   of the package and maps them in `apply()`, noting the mapping in the field
   docs.

`projects/main_ordering_test.go` holds one table per direction — ordering is
applied, and ordering is absent when unset — and every filter that supports
ordering belongs in both.

### Total counts (`skipCounts`) on v3 list endpoints

The count is real, it is not going away, and the SDK is usually already paying for it. Expose it, but let the caller choose, through the shared `twapi.ListCountMode` — never a per-resource enum, and never a raw `skipCounts` literal in a filter's `apply()`.

`skipCounts` defaults to **false**, so a request that omits the parameter comes back with an exact `count` in `meta.page`. A list that drops the value from its `Meta` is therefore not saving the API any work — it is discarding something already paid for and returned.

What the count costs, in terms a caller can observe:

- It is computed by a dedicated query that applies the same filters as the request. The API caches the result per filter combination, independently of the requested page, so paginating over a result set pays for it once rather than once per page.
- The cached value is invalidated by writes to the entities being counted, and the narrower the filter the longer it survives. A count scoped to one project, tasklist or milestone outlives one scoped to the whole installation, which any write to any of those entities invalidates. An unscoped exact count is the expensive one to ask for, so say so in the `CountMode` doc comment.
- Skipping the count is not uniformly the faster choice: the API is also free to use a known total to plan how it loads large result sets. That is why `ListCountModeDefault` exists, leaves the parameter unset, and is the default.

See the public pagination guide for the `meta.page` contract: <https://apidocs.teamwork.com/guides/teamwork/pagination>. Confirm the behaviour of a specific endpoint with a live call — send the same filters twice, once with `skipCounts=true` and once without, and compare `meta.page.count`.

The wiring, four steps, mirroring the sparse-fieldsets recipe:

1. **`CountMode twapi.ListCountMode`** on `*ListRequestFilters`, immediately before the `Fields` slot (which stays last). Document the entity noun and the default.
2. **`<recv>.CountMode.Apply(query)`** in the filter's `apply()`, just before `req.URL.RawQuery = query.Encode()`.
3. **`<recv>.Meta.ResolveCount(req.Filters.CountMode)`** in the response's `SetRequest`. This is mandatory: with `skipCounts=true` the server still returns a `count`, but it is the lower bound `(page * pageSize) + 1` it needs to derive `HasMore`. `ResolveCount` drops it so a lower bound can never reach a caller as a total. `Execute` runs `HandleHTTPResponse` before `SetRequest`, which is why the reconciliation lives there and not in the decoder.
4. **A constructor default of `twapi.ListCountModeSkip`** if — and only if — the endpoint used to hard-code `skipCounts=true`, so existing behaviour is preserved (`custom_item.go`, `custom_item_field.go`, `custom_item_record.go`, `time_report.go`).

A response with no `SetRequest` has nowhere to reconcile the count, so it must not offer `CountMode` at all — `WorkloadResponse` is the one such case. If an endpoint needs the option, give it a `SetRequest` whose only job is the reconciliation (`ProjectBudgetListResponse`, `RateUserGetResponse` both do).

The list response's `Meta` is `twapi.ListMeta` — see [Pagination (list responses)](#pagination-list-responses). Never redeclare the anonymous struct, and never expose `Offset` or `Size`: both are derivable from the request, and `ListMetaPage` deliberately omits them.

A response with its own `UnmarshalJSON` needs one extra check, because the four steps above all pass and the count still never arrives. If the decoder's local envelope redeclares `meta` as an anonymous struct, it decodes only the fields that struct happens to name, and `TestCountWiring` cannot see the omission — the wiring is all present, the count dies in the decoder. Type the envelope's `Meta` field as `twapi.ListMeta` and assign it wholesale (`c.Meta = envelope.Meta`), so the envelope cannot drift from the shared type. `CustomFieldValueListResponse` is the one response in this position; its tests assert the count survives the decoder for every wrapper key the endpoint can return.

### v1 and v3 are not interchangeable for the same operation

Several resources are reachable through both API versions, and the newer route is not always the one that does more. `comment.go`, `link.go` and `message.go` already write through v1 while reading through v3, because that is where the working endpoint is.

Moving a task between tasklists is reachable three ways, and **all three carry the task's subtasks**, so "does it cascade?" is not what separates them:

- `PUT /projects/api/v3/tasks/{id}.json` (`TaskUpdateRequest.TasklistID`). Clears the moved task's own parent link, since a subtask may not live outside its parent's tasklist. A `parentTaskId` sent alongside must already be in the destination, or the call fails.
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
- **Which headers to send is decided by the URL, not by the environment.** `X-Amz-SignedHeaders` lists the headers the signature covers. One of them missing, or an unsigned `x-amz-*` header added, and the upload fails. The API signs a canned ACL unless the installation's bucket sets one of its own, so read the parameter and repeat `X-Amz-Acl: public-read` only when it is listed there. Three internal clients (`twe2e`, `projectsapitesting`, `deskapi`) each hardcode a different guess at the environment for this instead; do not copy them. `NewPendingFileUploadPlan` is the one implementation of this rule, and `PendingFileUpload` builds its request from it — a caller that sends the contents itself takes the plan rather than re-deriving the headers, which is how a fourth wrong guess would enter.
- **Set the content length.** A body the standard library cannot measure is sent chunked, which the storage service rejects, so `Size` is required and is assigned to `http.Request.ContentLength`.
- The media type the `PUT` declares is what the stored file keeps — the later copy to the permanent bucket preserves the temporary object's metadata — so it is derived from the file name, with an override for callers who know better.
- The composite operation (`PendingFileCreate`) is a plain function over both steps, so its request and response types deliberately do not implement `twapi.HTTPRequester`/`HTTPResponser`. The individual steps stay exported (`PendingFilePresignedURL`, `PendingFileUpload`) for callers that stream a large file or hand the URL to a browser. A caller that never holds the bytes at all — a server telling a client where to send a file it has on disk — reserves with `PendingFilePresignedURL` and describes the rest with `NewPendingFileUploadPlan`, which reports the method, headers and expiry without sending anything. `Content-Length` is deliberately absent from `PendingFileUploadPlan.Headers`: Go derives it from the request, and its value is the size already reserved.

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

Every page-paginated list response uses the shared `twapi.ListMeta`, which carries `Page.HasMore` and `Page.Count` (see [Total counts](#total-counts-skipcounts-on-v3-list-endpoints)). Do not redeclare it as an anonymous struct — 33 responses share the one type, so a local copy silently opts that endpoint out of anything added to it later.

```go
type MessageListResponse struct {
    request  MessageListRequest  // unexported, set by Execute

    Meta twapi.ListMeta `json:"meta"`

    Messages []Message `json:"messages"`
}

func (r *MessageListResponse) SetRequest(req MessageListRequest) {
    r.request = req
    r.Meta.ResolveCount(req.Filters.CountMode)
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

Cursor-paginated endpoints are the exception: they report `prevCursor`/`nextCursor`/`limit` instead of a `page` object, so they keep their own `Meta` (`calendar_event.go`, `search.go`).

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

Sparse fieldsets are not universal on single-entity routes, and an unsupported or
unrecognised parameter is silently ignored — the payload comes back complete, with
no error to signal that the selection was dropped. So do not assume symmetry with
the list endpoint. Confirm it: call the route with a deliberately narrow selection
and check that the response actually shrank.

Three ways a get route can fail to honour it:

- **The route ignores the selection outright.** Verified cases so far:
  `customfield.go`, `custom_field_value.go` and `tag.go` — all three are
  deliberately left without a `sparsefields:get` marker.
- **The key doesn't match the singular envelope.** The word matters (casing does
  not). `fields[skill]` is a recognised parameter but only filters the plural
  `skills` payload, so it cannot narrow the skill get's `{"skill": …}` response —
  `skill.go` is unmarked for that reason.
- **The route is v1** (`link.go`, `team.go` gets) — no sparse-fieldset support at
  all.

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
| `ListMeta`         | The `meta` object of a page-paginated v3 list response; holds `Page.HasMore` and `Page.Count` |
| `ListCountMode`    | Whether the endpoint computes an exact total; `Apply(query)` writes `skipCounts` |

---

## Testing Patterns

### File naming

Tests live in standalone `_test.go` files, and **every test file must be named
after an existing non-test `.go` file in the same directory**: `a.go` is tested
by `a_test.go`, and by `a_{suffix}_test.go` when one file is not enough. A
`{suffix}_test.go` with no `{suffix}.go` beside it is wrong — name it after the
code under test, not after the scenario.

```
message.go   → message_test.go, message_example_test.go   ✅
             → sparse_fields_test.go                       ❌ (no sparse_fields.go)
```

- Integration tests: `projects/{resource}_test.go` (package `projects_test`)
- Example tests: `projects/{resource}_example_test.go` (package `projects_test`)
- Shared setup: `projects/main_test.go` — the one exception, since Go's `TestMain`
  and the shared fixtures have no source file of their own.

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
- **Document a date filter's boundary rule** — whether the boundary moment itself matches. A response never reveals it, so a caller that is not told brackets the window and re-filters the rows by hand. It is not guessable, and not uniform even inside one filter struct: on `TaskListRequestFilters`, `CreatedAfter`/`CompletedAfter` are inclusive while `UpdatedAfter` is exclusive; on `AllocationListRequestFilters`, `UpdatedAfter` is exclusive and `DeletedAfter` inclusive. Two shapes need a sentence rather than a word — a bound whose rule depends on another (`DueAfter`/`DueBefore` are each exclusive alone, but both set switches the endpoint to an inclusive range) and bounds that are not symmetric (`StartedAfterDate` includes the named day, `EndedBeforeDate` excludes it). Say it too when a filter narrows the result set on its own: `CompletedAfter`/`CompletedBefore` restrict the response to completed tasks whatever `IncludeCompletedTasks` says.
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
