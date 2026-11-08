//go:build windows

package codexappserver

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBridgeStopReapsWindowsChild(t *testing.T) {
	if os.Getenv("CODEX_BRIDGE_STOP_HELPER") == "1" {
		// Deliberately ignore stdin so Stop must terminate this process.
		select {}
	}
	executable, err := os.Executable()
	require.NoError(t, err)
	cmd := exec.Command(executable, "-test.run=^TestBridgeStopReapsWindowsChild$")
	cmd.Env = append(os.Environ(), "CODEX_BRIDGE_STOP_HELPER=1")
	require.NoError(t, cmd.Start())
	done := make(chan struct{})
	go func() { _ = cmd.Wait(); close(done) }()
	bridge := NewBridge(nil, nil)
	bridge.mu.Lock()
	bridge.cmd = cmd
	bridge.waitDone = done
	bridge.mu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	require.NoError(t, bridge.Stop(ctx))
	select {
	case <-done:
	default:
		t.Fatal("child was not reaped")
	}
}
