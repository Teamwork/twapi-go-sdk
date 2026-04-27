package projects_test

import (
	"context"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestLinkCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name  string
		input projects.LinkCreateRequest
	}{{
		name:  "only required fields",
		input: projects.NewLinkCreateRequest(testResources.ProjectID, "https://teamwork.com"),
	}, {
		name: "all fields",
		input: projects.LinkCreateRequest{
			Path: projects.LinkCreateRequestPath{
				ProjectID: testResources.ProjectID,
			},
			Title:             new("Teamwork.com Official Website"),
			Description:       new("Get news about the latest Teamwork.com features in the official website."),
			Code:              "https://teamwork.com",
			TagIDs:            projects.LegacyNumericList{testResources.TagID},
			NotifyCurrentUser: new(true),
			Notify:            projects.NewLinkNotifyAll(),
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			link, err := projects.LinkCreate(ctx, engine, tt.input)
			t.Cleanup(func() {
				if err != nil {
					return
				}
				ctx = context.Background() // t.Context is always canceled in cleanup
				_, err := projects.LinkDelete(ctx, engine, projects.NewLinkDeleteRequest(int64(link.ID)))
				if err != nil {
					t.Errorf("failed to delete link after test: %s", err)
				}
			})
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if link.ID == 0 {
				t.Error("expected a valid link ID but got 0")
			}
		})
	}
}

func TestLinkUpdate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	linkID, linkCleanup, err := createLink(t, testResources.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(linkCleanup)

	tests := []struct {
		name  string
		input projects.LinkUpdateRequest
	}{{
		name: "all fields",
		input: projects.LinkUpdateRequest{
			Path: projects.LinkUpdateRequestPath{
				ID: linkID,
			},
			Title:             new("Updated Teamwork.com Official Website"),
			Description:       new("Get news about the latest Teamwork.com features in the official website."),
			Code:              new("https://teamwork.com/"),
			TagIDs:            projects.LegacyNumericList{testResources.TagID},
			NotifyCurrentUser: new(true),
			Notify: projects.NewLinkNotifyGroup(projects.LegacyUserGroups{
				UserIDs:    []int64{testResources.UserID},
				CompanyIDs: []int64{testResources.CompanyID},
			}),
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.LinkUpdate(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

func TestLinkDelete(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	linkID, _, err := createLink(t, testResources.ProjectID)
	if err != nil {
		t.Fatal(err)
	}

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.LinkDelete(ctx, engine, projects.NewLinkDeleteRequest(linkID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestLinkGet(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	linkID, linkCleanup, err := createLink(t, testResources.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(linkCleanup)

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.LinkGet(ctx, engine, projects.NewLinkGetRequest(linkID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestLinkList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	_, linkCleanup, err := createLink(t, testResources.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(linkCleanup)

	tests := []struct {
		name  string
		input projects.LinkListRequest
	}{{
		name: "all links",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.LinkList(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}
