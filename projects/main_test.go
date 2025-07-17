package projects_test

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	twapi "github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
	"github.com/teamwork/twapi-go-sdk/session"
)

var engine *twapi.Engine

func TestMain(m *testing.M) {
	var exitCode int
	defer func() {
		os.Exit(exitCode)
	}()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	if engine = startEngine(); engine == nil {
		logger.Info("Missing setup environment variables, skipping tests")
		return
	}

	exitCode = m.Run()
}

func startEngine() *twapi.Engine {
	server, token := strings.TrimSuffix(os.Getenv("TWAPI_SERVER"), "/"), os.Getenv("TWAPI_TOKEN")
	if server == "" || token == "" {
		return nil
	}
	return twapi.NewEngine(session.NewBearerToken(token, server))
}

func createProject(t *testing.T) (int64, func(), error) {
	project, err := projects.ProjectCreate(t.Context(), engine, projects.ProjectCreateRequest{
		Name: fmt.Sprintf("Test Project %d", time.Now().UnixNano()),
	})
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create project for test: %w", err)
	}
	id := int64(project.ID)
	return id, func() {
		_, err := projects.ProjectDelete(t.Context(), engine, projects.NewProjectDeleteRequest(id))
		if err != nil {
			t.Errorf("failed to delete project after test: %s", err)
		}
	}, nil
}

func createTasklist(t *testing.T, projectID int64) (int64, func(), error) {
	tasklist, err := projects.TasklistCreate(t.Context(), engine, projects.TasklistCreateRequest{
		Path: projects.TasklistCreateRequestPath{
			ProjectID: projectID,
		},
		Name: fmt.Sprintf("Test Tasklist %d", time.Now().UnixNano()),
	})
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create tasklist for test: %w", err)
	}
	id := int64(tasklist.ID)
	return id, func() {
		_, err := projects.TasklistDelete(t.Context(), engine, projects.NewTasklistDeleteRequest(id))
		if err != nil {
			t.Errorf("failed to delete tasklist after test: %s", err)
		}
	}, nil
}

func createTask(t *testing.T, tasklistID int64) (int64, func(), error) {
	task, err := projects.TaskCreate(t.Context(), engine, projects.TaskCreateRequest{
		Path: projects.TaskCreateRequestPath{
			TasklistID: tasklistID,
		},
		Name: fmt.Sprintf("Test Task %d", time.Now().UnixNano()),
	})
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create task for test: %w", err)
	}
	id := task.Task.ID
	return id, func() {
		_, err := projects.TaskDelete(t.Context(), engine, projects.NewTaskDeleteRequest(id))
		if err != nil {
			t.Errorf("failed to delete task after test: %s", err)
		}
	}, nil
}
