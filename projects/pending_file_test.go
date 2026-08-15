package projects_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestPendingFileCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	name := fmt.Sprintf("test%d%d.txt", time.Now().UnixNano(), rand.Intn(100))

	// There is nothing to clean up: a pending file that is never attached is not
	// visible anywhere, and the API offers no way to discard one.
	pendingFile, err := projects.PendingFileCreate(ctx, engine,
		projects.NewPendingFileCreateRequest(name, []byte("This is a test file")))
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	} else if pendingFile.PendingFile.Ref == "" {
		t.Error("expected a valid pending file reference")
	}
}
