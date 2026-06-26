package isolation

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
)

// execInContainer runs a command inside a running container using docker exec.
//
// How it handles infinite loops / long-running code:
//  1. The command is wrapped with `timeout -s KILL <secs>` so the OS kills
//     the process if it exceeds the time limit.
//  2. A Go-side safety timeout (2s longer) acts as a fallback in case the
//     shell timeout doesn't fire.
//  3. Exit code 137 (SIGKILL) is detected and reported as "timeout".
func (c *Client) execInContainer(ctx context.Context, containerID, cmdStr, input string, timeout time.Duration) (*ExecuteResult, error) {
	start := time.Now()

	// Wrap command with shell `timeout` so the OS kills it on time limit.
	timeoutSecs := int(timeout.Seconds())
	wrappedCmd := fmt.Sprintf("timeout -s KILL %d %s", timeoutSecs, cmdStr)

	execConfig := container.ExecOptions{
		Cmd:          []string{"sh", "-c", wrappedCmd},
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		WorkingDir:   "/tmp",
	}

	execResp, err := c.cli.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create exec: %w", err)
	}

	hjResp, err := c.cli.ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to attach exec: %w", err)
	}
	defer hjResp.Close()

	// Write stdin and close so the process knows input is done.
	if input != "" {
		_, _ = hjResp.Conn.Write([]byte(input))
	}
	hjResp.CloseWrite()

	// Read stdout/stderr in a goroutine.
	var outBuf, errBuf bytes.Buffer
	done := make(chan error, 1)
	go func() {
		_, err := stdcopy.StdCopy(&outBuf, &errBuf, hjResp.Reader)
		done <- err
	}()

	// Go-side timeout as a safety net (slightly longer than shell timeout).
	safetyTimeout := timeout + 2*time.Second

	select {
	case <-done:
		// Check if the process was killed by the timeout command (exit code 137 = SIGKILL).
		inspect, inspectErr := c.cli.ContainerExecInspect(ctx, execResp.ID)
		if inspectErr == nil && inspect.ExitCode == 137 {
			return nil, fmt.Errorf("timeout")
		}
	case <-time.After(safetyTimeout):
		hjResp.Close()
		return nil, fmt.Errorf("timeout")
	}

	return &ExecuteResult{
		Output: outBuf.String(),
		Error:  errBuf.String(),
		TimeMs: time.Since(start).Milliseconds(),
	}, nil
}
