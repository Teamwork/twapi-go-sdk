package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestClassify(t *testing.T) {
	tests := []struct {
		name       string
		title      string
		body       string
		wantLevel  bumpLevel
		wantPrefix string
		classified bool
	}{
		{
			name: "feature", title: "Feature: File uploads and attachments",
			wantLevel: bumpMinor, wantPrefix: "feature", classified: true,
		}, {
			name: "feat", title: "feat: add a thing",
			wantLevel: bumpMinor, wantPrefix: "feat", classified: true,
		}, {
			name: "fix", title: "Fix: LegacyNumericList - Accept a bare number",
			wantLevel: bumpPatch, wantPrefix: "fix", classified: true,
		}, {
			name: "enhancement", title: "Enhancement: Support search highlights",
			wantLevel: bumpPatch, wantPrefix: "enhancement", classified: true,
		}, {
			name: "scoped chore", title: "Chore(deps): Bump golang.org/x/sys from 0.46.0 to 0.47.0",
			wantLevel: bumpPatch, wantPrefix: "chore", classified: true,
		}, {
			name: "lowercase", title: "docs: add public-repo content rules to AGENTS.md",
			wantLevel: bumpPatch, wantPrefix: "docs", classified: true,
		}, {
			name: "bang is breaking", title: "Feature!: replace Engine.HTTPClient with Engine.Do",
			wantLevel: bumpMajor, wantPrefix: "feature", classified: true,
		}, {
			name: "scoped bang is breaking", title: "Fix(tasks)!: drop the untyped Predecessors field",
			wantLevel: bumpMajor, wantPrefix: "fix", classified: true,
		}, {
			name: "breaking footer", title: "Fix: rename a field",
			body:      "BREAKING CHANGE: Engine.HTTPClient is gone",
			wantLevel: bumpMajor, wantPrefix: "fix", classified: true,
		}, {
			name: "breaking footer hyphenated", title: "Fix: rename a field",
			body:      "BREAKING-CHANGE: Engine.HTTPClient is gone",
			wantLevel: bumpMajor, wantPrefix: "fix", classified: true,
		}, {
			name: "breaking footer without a prefix", title: "Rename a field",
			body:      "BREAKING CHANGE: Engine.HTTPClient is gone",
			wantLevel: bumpMajor, classified: true,
		},

		// Real unprefixed titles from the history. The first was a breaking
		// change and the second added API surface, and both shipped without a
		// prefix — which is why an unrecognised prefix is reported rather than
		// trusted.
		{
			name: "no prefix", title: "Replace Engine.HTTPClient and Logger with Engine.Do",
			wantLevel: bumpPatch,
		}, {
			name: "no prefix, was a feature", title: "Support adding colors to tags in the SDK.",
			wantLevel: bumpPatch,
		}, {
			name: "sentence with a colon later", title: "Add colors to tags. This allows: grouping",
			wantLevel: bumpPatch,
		}, {
			name: "unknown prefix", title: "Note: something happened",
			wantLevel: bumpPatch, wantPrefix: "note",
		}, {
			name: "empty", title: "",
			wantLevel: bumpPatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix, level, ok := classify(tt.title, tt.body)
			if level != tt.wantLevel {
				t.Errorf("level: want %s, got %s", tt.wantLevel, level)
			}
			if prefix != tt.wantPrefix {
				t.Errorf("prefix: want %q, got %q", tt.wantPrefix, prefix)
			}
			if ok != tt.classified {
				t.Errorf("classified: want %v, got %v", tt.classified, ok)
			}
		})
	}
}

// TestEveryContributingPrefixIsMapped keeps the prefix table and the prefixes
// CONTRIBUTING.md asks contributors to use from drifting apart: a documented
// prefix the tool does not know would be rejected by the PR lint that tells
// authors to read that very list.
func TestEveryContributingPrefixIsMapped(t *testing.T) {
	doc, err := os.ReadFile("../../CONTRIBUTING.md")
	if err != nil {
		t.Fatalf("read CONTRIBUTING.md: %v", err)
	}

	documented := regexp.MustCompile("(?m)^\\s+- `([A-Za-z]+):` for ").FindAllStringSubmatch(string(doc), -1)
	if len(documented) == 0 {
		t.Fatal("found no documented prefixes in CONTRIBUTING.md; has the list moved?")
	}

	for _, match := range documented {
		prefix := strings.ToLower(match[1])
		if _, ok := bumpByPrefix[prefix]; !ok {
			t.Errorf("CONTRIBUTING.md documents %q but bumpByPrefix does not map it", prefix)
		}
		// The PR lint has to accept what the docs tell authors to write.
		if _, err := checkTitle(match[1] + ": a subject"); err != nil {
			t.Errorf("CONTRIBUTING.md documents %q but the title check refuses it: %v", prefix, err)
		}
	}
}

// TestCheckTitle covers what the PR lint workflow enforces. The vocabulary is
// bumpByPrefix itself, so the check and the release can never disagree; what is
// worth pinning is the shapes accepted and refused at the door.
func TestCheckTitle(t *testing.T) {
	accepted := map[string]bumpLevel{
		"Feature: File uploads and attachments":            bumpMinor,
		"feat: support orderBy and orderMode":              bumpMinor,
		"Fix: LegacyNumericList - Accept a bare number":    bumpPatch,
		"Enhancement: Support search highlights":           bumpPatch,
		"Chore(deps): Bump golang.org/x/sys":               bumpPatch,
		"Chore(deps-dev): Bump a tool":                     bumpPatch,
		"Refactor: extract the pagination helper":          bumpPatch,
		"Feature!: replace Engine.HTTPClient with Do":      bumpMajor,
		"Fix(tasks)!: drop the untyped Predecessors field": bumpMajor,
	}
	for title, want := range accepted {
		got, err := checkTitle(title)
		if err != nil {
			t.Errorf("%q: want accepted, got %v", title, err)
			continue
		}
		if got != want {
			t.Errorf("%q: want a %s bump, got %s", title, want, got)
		}
	}

	refused := []string{
		// Real titles from the history. The first two added API surface; the
		// third was a behaviour change to OptionalDateTime.
		"Support adding colors to tags in the Teamwork API SDK.",
		"Replace Engine.HTTPClient and Logger with Engine.Do",
		"set zero value of OptionalDateTime to null",

		"Note: an unknown prefix",
		// A breaking marker must not smuggle an unknown prefix past the
		// vocabulary, even though classify() would read it as a major.
		"Whatever!: sneaking in",
		"Fix:",
		"Fix:   ",
		"",
	}
	for _, title := range refused {
		if _, err := checkTitle(title); err == nil {
			t.Errorf("%q: want refused, got accepted", title)
		}
	}
}

// TestBreakingChangeIsCappedAtMinor pins the module-path rule: a v2.0.0 tag on
// this path is one `go get` will not resolve, so a breaking change is reported
// as a major and tagged as a minor.
func TestBreakingChangeIsCappedAtMinor(t *testing.T) {
	level, breaking := capMajor(bumpMajor)
	if level != bumpMinor || !breaking {
		t.Errorf("a major should cap to minor and be flagged, got %s / %v", level, breaking)
	}

	for _, uncapped := range []bumpLevel{bumpMinor, bumpPatch, bumpNone} {
		level, breaking := capMajor(uncapped)
		if level != uncapped || breaking {
			t.Errorf("%s should pass through unflagged, got %s / %v", uncapped, level, breaking)
		}
	}

	// The reports have to say so: shipping a breaking change inside a minor is
	// only safe if whoever writes the release notes knows about it.
	rep := report{
		Previous: version{1, 21, 6},
		Next:     version{1, 22, 0},
		Level:    bumpMinor,
		Breaking: true,
		Changes: []change{
			{
				Ref: "#122", Title: "Feature!: replace Engine.HTTPClient",
				Prefix: "feature", Level: bumpMajor, Classified: true,
			},
		},
	}
	if text := rep.text(); !strings.Contains(text, "BREAKING") {
		t.Errorf("the text report hides the breaking change:\n%s", text)
	}
	markdown := rep.markdown()
	for _, want := range []string{"[!CAUTION]", "/v2", "go get -u"} {
		if !strings.Contains(markdown, want) {
			t.Errorf("the markdown report is missing %q:\n%s", want, markdown)
		}
	}
}

// TestAcceptedPrefixesHelpIsComplete checks the rejection message names every
// prefix that would have been accepted: it is the only guidance the author gets
// on a failed check.
func TestAcceptedPrefixesHelpIsComplete(t *testing.T) {
	help := acceptedPrefixes()
	for prefix := range bumpByPrefix {
		if !strings.Contains(help, prefix) {
			t.Errorf("the help does not mention the accepted prefix %q:\n%s", prefix, help)
		}
	}
}

// TestHistoricalRangesComputeTheRightBump replays real ranges from the history,
// classifying commit subjects the way the tool does when a pull-request lookup
// is unavailable. The range marked as under-versioned shipped a patch tag over
// a feat:; that is the tag this tool exists to prevent.
func TestHistoricalRangesComputeTheRightBump(t *testing.T) {
	tests := []struct {
		name           string
		previous       string
		subjects       []string
		wantVersion    string
		wantUnclassifd int
	}{
		{
			name:        "v1.21.6 shipped a feat as a patch",
			previous:    "v1.21.5",
			wantVersion: "v1.22.0",
			subjects: []string{
				// #120's body also declared a BREAKING CHANGE; read with the
				// body it earns a major, which caps to this same minor.
				"feat: support orderBy and orderMode on list endpoints (#120)",
			},
		},
		{
			name:        "v1.21.3 shipped new API surface under an unprefixed title",
			previous:    "v1.21.2",
			wantVersion: "v1.21.3", // a patch, and reported as unclassified
			subjects: []string{
				"Support adding colors to tags in the Teamwork API SDK. (#115)",
			},
			wantUnclassifd: 1,
		},
		{
			name:        "v1.22.0 shipped a feature, as a minor",
			previous:    "v1.21.6",
			wantVersion: "v1.22.0",
			subjects: []string{
				"Feature: File uploads and attachments",
				"Address review feedback",
				"Address review and simplification feedback",
				"Change file upload strategy to use pre-signed",
				"Replace Engine.HTTPClient and Logger with Engine.Do",
			},
			// The commits of the rebase-merged branch that carry no prefix of
			// their own. With a pull-request lookup they collapse into one
			// classified change; this is the subject-only fallback.
			wantUnclassifd: 4,
		},
		{
			name:        "enhancement plus docs stays a patch, as v1.21.5 shipped",
			previous:    "v1.21.4",
			wantVersion: "v1.21.5",
			subjects: []string{
				"Enhancement: Expose exact item counts on v3 list endpoints",
				"docs: add public-repo content rules to AGENTS.md",
			},
		},
		{
			name:        "fix-only stays a patch, as v1.20.8 shipped",
			previous:    "v1.20.7",
			wantVersion: "v1.20.8",
			subjects: []string{
				"Fix: CalendarEvent - Send sparse fields as fields[calendarsEvents]",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commits := make([]commit, 0, len(tt.subjects))
			for i, subject := range tt.subjects {
				commits = append(commits, commit{SHA: fmt.Sprintf("%040d", i), Subject: subject})
			}

			changes, _ := collectChanges(commits, nil, nil)

			previous, ok := parseVersion(tt.previous)
			if !ok {
				t.Fatalf("bad previous tag %q in the table", tt.previous)
			}
			if got := previous.next(highestLevel(changes)).String(); got != tt.wantVersion {
				t.Errorf("version: want %s, got %s", tt.wantVersion, got)
			}

			var unclassified int
			for _, c := range changes {
				if !c.Classified {
					unclassified++
				}
			}
			if unclassified != tt.wantUnclassifd {
				t.Errorf("unclassified: want %d, got %d", tt.wantUnclassifd, unclassified)
			}
		})
	}
}

func TestCollectChangesFoldsCommitsIntoPullRequests(t *testing.T) {
	commits := []commit{
		{SHA: "aaaaaaaa1", Subject: "Feature: File uploads and attachments"},
		{SHA: "aaaaaaaa2", Subject: "Address review feedback"},
		{SHA: "aaaaaaaa3", Subject: "Change file upload strategy to use pre-signed"},
		{SHA: "bbbbbbbb1", Subject: "feat: support orderBy and orderMode (#120)"},
		{SHA: "dddddddd1", Subject: "Enhancement: Support search highlights"},
	}

	prs := map[string]*pullRequest{
		"aaaaaaaa1": {Number: 122, Title: "Feature: File uploads and attachments", MergedAt: "2026-08-01T00:00:00Z"},
		"aaaaaaaa2": {Number: 122, Title: "Feature: File uploads and attachments", MergedAt: "2026-08-01T00:00:00Z"},
		"aaaaaaaa3": {Number: 122, Title: "Feature: File uploads and attachments", MergedAt: "2026-08-01T00:00:00Z"},
		"bbbbbbbb1": {Number: 120, Title: "feat: support orderBy and orderMode", MergedAt: "2026-08-02T00:00:00Z"},
	}

	changes, notes := collectChanges(commits, func(sha string) (*pullRequest, error) {
		return prs[sha], nil
	}, nil)
	if len(notes) != 0 {
		t.Errorf("unexpected notes: %v", notes)
	}

	// The three commits of #122 collapse into one change, and the commit with no
	// pull request keeps its own subject and short sha.
	want := []change{
		{
			Ref: "#122", Title: "Feature: File uploads and attachments",
			Prefix: "feature", Level: bumpMinor, Classified: true,
		},
		{Ref: "#120", Title: "feat: support orderBy and orderMode", Prefix: "feat", Level: bumpMinor, Classified: true},
		{
			Ref: "dddddddd", Title: "Enhancement: Support search highlights",
			Prefix: "enhancement", Level: bumpPatch, Classified: true,
		},
	}
	if len(changes) != len(want) {
		t.Fatalf("want %d changes, got %d: %+v", len(want), len(changes), changes)
	}
	for i, w := range want {
		if changes[i] != w {
			t.Errorf("change %d: want %+v, got %+v", i, w, changes[i])
		}
	}

	if got := highestLevel(changes); got != bumpMinor {
		t.Errorf("level: want minor, got %s", got)
	}
}

func TestCollectChangesFallsBackWhenLookupFails(t *testing.T) {
	commits := []commit{
		{SHA: "aaaaaaaa1", Subject: "Feature: File attachments"},
		{SHA: "bbbbbbbb1", Subject: "Fix: something"},
	}

	changes, notes := collectChanges(commits, func(string) (*pullRequest, error) {
		return nil, fmt.Errorf("403 Forbidden")
	}, nil)

	if len(changes) != 2 {
		t.Fatalf("want both commits classified by subject, got %+v", changes)
	}
	if changes[0].Ref != "aaaaaaaa" || changes[0].Level != bumpMinor {
		t.Errorf("first change should fall back to its subject: %+v", changes[0])
	}
	if len(notes) != 1 || !strings.Contains(notes[0], "403 Forbidden") {
		t.Errorf("want a note naming the failure, got %v", notes)
	}
}

func TestHighestVersionTagIgnoresNonVersions(t *testing.T) {
	tags := []string{"v1.9.0", "v1.22.0", "vnext", "v1.21.6", "v2", "v1.22.10", "v1.22.0-rc1", ""}
	if got := highestVersionTag(tags); got != "v1.22.10" {
		t.Errorf("want v1.22.10 (numeric, not lexical), got %q", got)
	}
	if got := highestVersionTag(nil); got != "" {
		t.Errorf("want no tag, got %q", got)
	}
}

func TestVersionNextResetsLowerComponents(t *testing.T) {
	v := version{1, 22, 0}
	tests := []struct {
		level bumpLevel
		want  string
	}{
		{bumpPatch, "v1.22.1"},
		{bumpMinor, "v1.23.0"},
		{bumpMajor, "v2.0.0"},
		{bumpNone, "v1.22.0"},
	}
	for _, tt := range tests {
		if got := v.next(tt.level).String(); got != tt.want {
			t.Errorf("%s bump: want %s, got %s", tt.level, tt.want, got)
		}
	}
}

func TestPickPullRequestPrefersMerged(t *testing.T) {
	open := pullRequest{Number: 1, Title: "open branch that contains the commit"}
	merged := pullRequest{Number: 2, Title: "Feature: the one that shipped", MergedAt: "2026-08-01T00:00:00Z"}

	if got := pickPullRequest([]pullRequest{open, merged}); got == nil || got.Number != 2 {
		t.Errorf("want the merged pull request, got %+v", got)
	}
	if got := pickPullRequest([]pullRequest{open}); got != nil {
		t.Errorf("want nil when nothing merged, got %+v", got)
	}
	if got := pickPullRequest(nil); got != nil {
		t.Errorf("want nil for no pull requests, got %+v", got)
	}
}

func TestRepoFromRemote(t *testing.T) {
	tests := map[string]string{
		"https://github.com/teamwork/mcp.git\n": "teamwork/mcp",
		"https://github.com/teamwork/mcp":       "teamwork/mcp",
		"git@github.com:teamwork/mcp.git":       "teamwork/mcp",
		"ssh://git@github.com/teamwork/mcp.git": "teamwork/mcp",
		"https://gitlab.com/teamwork/mcp.git":   "",
	}
	for remote, want := range tests {
		if got := repoFromRemote(remote); got != want {
			t.Errorf("%q: want %q, got %q", remote, want, got)
		}
	}
}

func TestReportReportsUnclassifiedChanges(t *testing.T) {
	rep := report{
		Previous: version{1, 22, 0},
		Next:     version{1, 22, 1},
		Level:    bumpPatch,
		Changes: []change{
			{Ref: "#470", Title: "Fix: a | in a title", Prefix: "fix", Level: bumpPatch, Classified: true},
			{Ref: "#471", Title: "Adds support for something", Level: bumpPatch},
		},
	}

	outputs := rep.outputs()
	for _, want := range []string{"version=v1.22.1", "previous_tag=v1.22.0", "bump=patch", "unclassified=1"} {
		if !strings.Contains(outputs, want) {
			t.Errorf("outputs missing %q:\n%s", want, outputs)
		}
	}

	markdown := rep.markdown()
	for _, want := range []string{"[!WARNING]", "bump: minor", `a \| in a title`, "⚠️ patch"} {
		if !strings.Contains(markdown, want) {
			t.Errorf("markdown missing %q:\n%s", want, markdown)
		}
	}

	if text := rep.text(); !strings.Contains(text, "WARNING: 1 change(s)") {
		t.Errorf("text missing the warning:\n%s", text)
	}
}

func TestParseBump(t *testing.T) {
	tests := []struct {
		in   string
		want bumpLevel
	}{
		{"auto", bumpNone},
		{"", bumpNone},
		{"patch", bumpPatch},
		{"MINOR", bumpMinor},
		{" major ", bumpMajor}, // the workflow interpolates the input as-is
	}
	for _, tt := range tests {
		got, err := parseBump(tt.in)
		if err != nil {
			t.Errorf("%q: unexpected error: %v", tt.in, err)
		}
		if got != tt.want {
			t.Errorf("%q: want %s, got %s", tt.in, tt.want, got)
		}
	}
	if _, err := parseBump("massive"); err == nil {
		t.Error("want an error for an unknown bump")
	}
}
