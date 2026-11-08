//go:build windows

package e2e_test

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

// Closing this job terminates the runtime and any surviving descendants.
func runtimeArtifactJob(pid int) (func(), error) {
	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return nil, err
	}
	fail := func(e error) (func(), error) { _ = windows.CloseHandle(job); return nil, e }
	var info windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION
	info.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	if _, err = windows.SetInformationJobObject(job, windows.JobObjectExtendedLimitInformation, uintptr(unsafe.Pointer(&info)), uint32(unsafe.Sizeof(info))); err != nil {
		return fail(err)
	}
	process, err := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE, false, uint32(pid))
	if err != nil {
		return fail(err)
	}
	defer windows.CloseHandle(process)
	if err = windows.AssignProcessToJobObject(job, process); err != nil {
		return fail(fmt.Errorf("assign runtime pid %d to job: %w", pid, err))
	}
	return func() { _ = windows.CloseHandle(job) }, nil
}

func TestRuntimeArtifactJobTerminatesChild(t *testing.T) {
	if os.Getenv("RUNTIME_JOB_HELPER") == "1" {
		for {
			time.Sleep(time.Second)
		}
	}
	exe, err := os.Executable()
	require.NoError(t, err)
	cmd := exec.Command(exe, "-test.run=^TestRuntimeArtifactJobTerminatesChild$")
	cmd.Env = append(os.Environ(), "RUNTIME_JOB_HELPER=1")
	require.NoError(t, cmd.Start())
	defer func() { _ = cmd.Process.Kill() }()
	closeJob, err := runtimeArtifactJob(cmd.Process.Pid)
	require.NoError(t, err)
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	closeJob()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("job did not terminate child")
	}
}
