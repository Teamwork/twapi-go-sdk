package projects_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestCommentCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name  string
		input projects.CommentCreateRequest
	}{{
		name: "only required fields for milestone",
		input: projects.NewCommentCreateRequestInMilestone(
			testResources.MilestoneID,
			fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
		),
	}, {
		name: "only required fields for task",
		input: projects.NewCommentCreateRequestInTask(
			testResources.TaskID,
			fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
		),
	}, {
		name: "all fields for milestone",
		input: projects.CommentCreateRequest{
			Path: projects.CommentCreateRequestPath{
				MilestoneID: testResources.MilestoneID,
			},
			Body:        "<h1>This is a test comment</h1>",
			ContentType: twapi.Ptr("HTML"),
		},
	}, {
		name: "all fields for task",
		input: projects.CommentCreateRequest{
			Path: projects.CommentCreateRequestPath{
				TaskID: testResources.TaskID,
			},
			Body:        "<h1>This is a test comment</h1>",
			ContentType: twapi.Ptr("HTML"),
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			t.Cleanup(cancel)

			comment, err := projects.CommentCreate(ctx, engine, tt.input)
			defer func() {
				if err != nil {
					return
				}
				ctx = context.Background() // t.Context is always canceled in cleanup
				_, err := projects.CommentDelete(ctx, engine, projects.NewCommentDeleteRequest(int64(comment.ID)))
				if err != nil {
					t.Errorf("failed to delete comment after test: %s", err)
				}
			}()
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if comment.ID == 0 {
				t.Error("expected a valid comment ID but got 0")
			}
		})
	}
}

func TestCommentUpdate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	commentID, commentCleanup, err := createCommentInTask(t, testResources.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(commentCleanup)

	tests := []struct {
		name  string
		input projects.CommentUpdateRequest
	}{{
		name: "all fields",
		input: projects.CommentUpdateRequest{
			Path: projects.CommentUpdateRequestPath{
				ID: commentID,
			},
			Body: "<h1>Updated comment</h1>",
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.CommentUpdate(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

func TestCommentDelete(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	commentID, _, err := createCommentInTask(t, testResources.TaskID)
	if err != nil {
		t.Fatal(err)
	}

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.CommentDelete(ctx, engine, projects.NewCommentDeleteRequest(commentID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestCommentGet(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	commentID, commentCleanup, err := createCommentInTask(t, testResources.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(commentCleanup)

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.CommentGet(ctx, engine, projects.NewCommentGetRequest(commentID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestCommentList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	_, milestoneCommentCleanup, err := createCommentInMilestone(t, testResources.MilestoneID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(milestoneCommentCleanup)

	_, taskCommentCleanup, err := createCommentInTask(t, testResources.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(taskCommentCleanup)

	tests := []struct {
		name  string
		input projects.CommentListRequest
	}{{
		name: "all comments",
	}, {
		name: "comments for milestone",
		input: projects.CommentListRequest{
			Path: projects.CommentListRequestPath{
				MilestoneID: testResources.MilestoneID,
			},
		},
	}, {
		name: "comments for task",
		input: projects.CommentListRequest{
			Path: projects.CommentListRequestPath{
				TaskID: testResources.TaskID,
			},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.CommentList(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}
