//go:build windows

package e2e_test

import (
	"os"

	"golang.org/x/sys/windows"
)

// Exit when the runtime under test dies without shutting down its Codex
// subprocess. Otherwise the orphan keeps Go's e2e.test.exe locked on Windows.
func watchRuntimeHelperParent() {
	parent := os.Getppid()
	handle, err := windows.OpenProcess(windows.SYNCHRONIZE, false, uint32(parent))
	if err != nil {
		os.Exit(0)
	}
	go func() {
		defer windows.CloseHandle(handle)
		_, _ = windows.WaitForSingleObject(handle, windows.INFINITE)
		os.Exit(0)
	}()
}
