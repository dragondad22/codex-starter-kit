//go:build windows

package engine_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/dragondad22/codex-starter-kit/engine"
)

func TestCreateRejectsReservedDirectoryJunctionEscapeDuringPlanning(t *testing.T) {
	repository := newGitRepository(t)
	outside := t.TempDir()
	junction := filepath.Join(repository, ".starter-kit")
	command := exec.Command("cmd.exe", "/d", "/c", "mklink", "/j", junction, outside)
	if output, err := command.CombinedOutput(); err != nil {
		t.Skipf("native filesystem cannot create junction fixture: %v: %s", err, output)
	}

	if _, err := engine.New().Create(t.Context(), approvedCreate(repository)); err == nil {
		t.Fatal("create planned through reserved directory junction")
	}
	entries, err := os.ReadDir(outside)
	if err != nil {
		t.Fatalf("read junction target: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("planning wrote outside repository through junction: %#v", entries)
	}
}
