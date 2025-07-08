package projects_test

import (
	"log/slog"
	"os"
	"testing"

	twapi "github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/session"
)

var engine *twapi.Engine

func startEngine() *twapi.Engine {
	server, token := os.Getenv("TWAPI_SERVER"), os.Getenv("TWAPI_TOKEN")
	if server == "" || token == "" {
		return nil
	}
	return twapi.NewEngine(session.NewBearerToken(token, server))
}

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
