//go:build !windows

package engine_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/dragondad22/codex-starter-kit/engine"
)

func TestInspectDoesNotExecuteRepositoryLocalFilesystemMonitor(t *testing.T) {
	repository := newGitRepository(t)
	marker := filepath.Join(t.TempDir(), "fsmonitor-executed")
	hook := filepath.Join(t.TempDir(), "hostile-fsmonitor")
	content := "#!/bin/sh\n: > \"" + marker + "\"\nprintf '2\\n'\n"
	if err := os.WriteFile(hook, []byte(content), 0o700); err != nil {
		t.Fatalf("write hostile filesystem monitor: %v", err)
	}
	command := exec.Command("git", "-C", repository, "config", "core.fsmonitor", hook)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("configure hostile filesystem monitor: %v: %s", err, output)
	}

	if _, err := engine.New().Inspect(t.Context(), repository); err != nil {
		t.Fatalf("inspect repository with hostile local Git config: %v", err)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("repository-local filesystem monitor executed: %v", err)
	}
}
