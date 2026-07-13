//go:build windows

package engine

import (
	"errors"
	"syscall"
)

func processAlive(pid int) bool {
	const processQueryLimitedInformation = 0x1000
	handle, err := syscall.OpenProcess(processQueryLimitedInformation, false, uint32(pid))
	if err == nil {
		_ = syscall.CloseHandle(handle)
		return true
	}
	return errors.Is(err, syscall.ERROR_ACCESS_DENIED)
}
