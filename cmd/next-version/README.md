# next-version

Computes the next release version from the changes since the last tag, so the
bump is derived from what shipped rather than picked by hand.

```bash
go run ./cmd/next-version                 # report the next version
go run ./cmd/next-version -bump=minor     # force a bump level
go run ./cmd/next-version -from=v1.21.0   # diff from an explicit tag
go run ./cmd/next-version -no-pr-lookup   # classify commit subjects only

go run ./cmd/next-version -check-title="Feature: File uploads"
```

## How a change is classified

Each commit since the last tag is resolved to the pull request it merged
through, and the bump comes from that pull request's title. All commits of one
pull request count once, so a rebase-merged branch is judged by its title, not
by commits like `Fix comment`. A commit with no pull request falls back to its
own subject.

| Title | Bump |
| --- | --- |
| `Feature:`, `Feat:` | minor |
| `Fix:`, `Enhancement:`, `Chore:`, `Docs:`, `Test:`, `Refactor:`, `Perf:`, `Build:`, `CI:`, `Style:`, `Revert:` | patch |
| any prefix with `!` (`Feature!:`), or a `BREAKING CHANGE:` body | major — see below |
| anything else | patch, **reported as unclassified** |

The release takes the highest bump any single change asks for. Prefixes are
case-insensitive and may carry a scope: `Chore(deps):` is a patch.

`Enhancement:` is a patch because every Enhancement-only release so far shipped
as one (v1.20.7, v1.21.2, v1.21.5).

## Breaking changes cannot be a major here

A breaking change is *reported* as a major and *tagged* as a minor. This is a Go
module: v2 is reached by moving the module path to `/v2`, not by tagging
`v2.0.0`, so a `v2.0.0` tag on this path is one `go get` will not resolve.

That is also what this repository has always done — v1.22.0 replaced
`Engine.HTTPClient` and `Logger` with `Engine.Do`, as a minor. Both reports say
so loudly, because consumers pick the change up on `go get -u` and it belongs in
the release notes. `-bump=major` stays available, for a deliberate module-path
migration.

## Unclassified changes

A title with no known prefix counts as a patch and is listed as unclassified,
in the terminal and as a warning in the workflow summary. That is not a
formality: v1.21.3 shipped new API surface under the unprefixed "Support adding
colors to tags". **Read the unclassified list before releasing** — if one of them
is a feature, re-run with `bump: minor`.

The prefix matters even when it is present: v1.21.6 was a patch tag over #120,
whose title was `feat: support orderBy and orderMode` and whose body declared a
`BREAKING CHANGE` renaming six exported identifiers. This tool reads that body,
so the same release now computes a minor and prints the breaking-change
banner.

## Checking a title

`-check-title` validates one title against the table above and prints the bump
it earns, exiting non-zero with the accepted prefixes if it has none. The `PR
lint` workflow runs exactly that, so the check and the release read a title the
same way — there is no second list of prefixes anywhere.

```console
$ go run ./cmd/next-version -check-title="Feature: File uploads and attachments"
Title accepted; this pull request earns a minor bump.
```

Use it locally before opening a pull request, or to see why the check failed on
one.

## Where it runs

`.github/workflows/release.yaml` calls it on the `workflow_dispatch` path,
where it writes `version`, `previous_tag`, `bump` and `unclassified` to
`$GITHUB_OUTPUT` and a per-change table to the run summary. The workflow then
creates that tag and releases it.

Use `-bump` to override the computation, and the workflow's `dry_run` input to
see the version and the table without tagging anything.

Locally the pull-request lookup needs `GH_TOKEN` or `GITHUB_TOKEN`
(`GH_TOKEN=$(gh auth token)`); without one it falls back to commit subjects and
says so.
