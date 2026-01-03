package projects_test

import (
	"context"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestTimelogCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name  string
		input projects.TimelogCreateRequest
	}{{
		name: "only required fields for task",
		input: projects.NewTimelogCreateRequestInTask(
			testResources.TaskID,
			time.Now(),
			30*time.Minute,
		),
	}, {
		name: "only required fields for project",
		input: projects.NewTimelogCreateRequestInProject(
			testResources.ProjectID,
			time.Now(),
			30*time.Minute,
		),
	}, {
		name: "all fields for task",
		input: projects.TimelogCreateRequest{
			Path: projects.TimelogCreateRequestPath{
				TaskID: testResources.TaskID,
			},
			Description: twapi.Ptr("This is a test timelog"),
			Date:        twapi.Date(time.Now().UTC()),
			Time:        twapi.Time(time.Now().UTC()),
			IsUTC:       true,
			Hours:       2,
			Minutes:     30,
			Billable:    true,
			UserID:      &testResources.UserID,
			TagIDs:      []int64{testResources.TagID},
		},
	}, {
		name: "all fields for project",
		input: projects.TimelogCreateRequest{
			Path: projects.TimelogCreateRequestPath{
				ProjectID: testResources.ProjectID,
			},
			Description: twapi.Ptr("This is a test timelog"),
			Date:        twapi.Date(time.Now().UTC()),
			Time:        twapi.Time(time.Now().UTC()),
			IsUTC:       true,
			Hours:       2,
			Minutes:     30,
			Billable:    true,
			UserID:      &testResources.UserID,
			TagIDs:      []int64{testResources.TagID},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			timelogResponse, err := projects.TimelogCreate(ctx, engine, tt.input)
			defer func() {
				if err != nil {
					return
				}
				ctx = context.Background() // t.Context is always canceled in cleanup
				_, err := projects.TimelogDelete(ctx, engine, projects.NewTimelogDeleteRequest(timelogResponse.Timelog.ID))
				if err != nil {
					t.Errorf("failed to delete timelog after test: %s", err)
				}
			}()
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if timelogResponse.Timelog.ID == 0 {
				t.Error("expected a valid timelog ID but got 0")
			}
		})
	}
}

func TestTimelogUpdate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	timelogID, timelogCleanup, err := createTimelogInTask(t, testResources.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(timelogCleanup)

	tests := []struct {
		name  string
		input projects.TimelogUpdateRequest
	}{{
		name: "all fields",
		input: projects.TimelogUpdateRequest{
			Path: projects.TimelogUpdateRequestPath{
				ID: timelogID,
			},
			Description: twapi.Ptr("This is a test timelog"),
			Date:        twapi.Ptr(twapi.Date(time.Now().UTC())),
			Time:        twapi.Ptr(twapi.Time(time.Now().UTC())),
			IsUTC:       twapi.Ptr(true),
			Hours:       twapi.Ptr[int64](2),
			Minutes:     twapi.Ptr[int64](30),
			Billable:    twapi.Ptr(true),
			UserID:      &testResources.UserID,
			TagIDs:      []int64{testResources.TagID},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.TimelogUpdate(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

func TestTimelogDelete(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	timelogID, timelogCleanup, err := createTimelogInTask(t, testResources.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(timelogCleanup)

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.TimelogDelete(ctx, engine, projects.NewTimelogDeleteRequest(timelogID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestTimelogGet(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	timelogID, timelogCleanup, err := createTimelogInTask(t, testResources.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(timelogCleanup)

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.TimelogGet(ctx, engine, projects.NewTimelogGetRequest(timelogID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestTimelogList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	_, taskTimelogCleanup, err := createTimelogInTask(t, testResources.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(taskTimelogCleanup)

	_, projectTimelogCleanup, err := createTimelogInProject(t, testResources.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(projectTimelogCleanup)

	tests := []struct {
		name  string
		input projects.TimelogListRequest
	}{{
		name: "all timelogs",
	}, {
		name: "timelogs for task",
		input: projects.TimelogListRequest{
			Path: projects.TimelogListRequestPath{
				TaskID: testResources.TaskID,
			},
		},
	}, {
		name: "timelogs for project",
		input: projects.TimelogListRequest{
			Path: projects.TimelogListRequestPath{
				ProjectID: testResources.ProjectID,
			},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.TimelogList(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}
