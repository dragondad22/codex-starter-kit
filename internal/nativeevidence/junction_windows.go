//go:build windows

package nativeevidence

import (
	"os"
	"os/exec"
	"path/filepath"
)

func probeJunction(root string) Capability {
	directory := filepath.Join(root, ".git", "phase1-junction-probe")
	target := filepath.Join(directory, "target")
	junction := filepath.Join(directory, "junction")
	if err := os.MkdirAll(target, 0o700); err != nil {
		return Capability{ID: "directory-junction", State: "needs-review", Details: "junction target could not be created"}
	}
	command := exec.Command("cmd.exe", "/d", "/c", "mklink", "/j", junction, target)
	if err := command.Run(); err != nil {
		return Capability{ID: "directory-junction", State: "not-configured", Details: "native runner did not grant directory-junction creation"}
	}
	return Capability{ID: "directory-junction", State: "supported", Details: "native directory-junction creation and rejection fixtures are available"}
}
