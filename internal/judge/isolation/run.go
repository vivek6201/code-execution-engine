package isolation

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
)

// Run executes code in a container with a single input. This is a convenience
// wrapper around RunBatch for single executions (no test cases).
func (c *Client) Run(ctx context.Context, img, filename, code, compileAndRunCmd, input string) (*ExecuteResult, int64, error) {
	results, memKB, err := c.RunBatch(ctx, img, filename, code, "", compileAndRunCmd, []string{input})
	if err != nil {
		return nil, 0, err
	}
	return &results[0], memKB, nil
}

// RunBatch creates a single long-lived container, compiles once, and runs the
// program for each input concurrently using docker exec.
//
// Container lifecycle:
//  1. Pull image (if needed)
//  2. Create container with `sleep infinity` (keeps it alive)
//  3. Copy code file into /app
//  4. Start container
//  5. Compile (if compileCmd is provided)
//  6. Run the program for each input concurrently (bounded by maxConcurrency)
//  7. Cleanup: force remove container
//
// The ctx parameter supports cancellation — when cancelled, remaining test cases
// are skipped and marked as "cancelled".
func (c *Client) RunBatch(ctx context.Context, img, filename, code, compileCmd, runCmd string, inputs []string) ([]ExecuteResult, int64, error) {
	if err := c.ensureImage(ctx, img); err != nil {
		return nil, 0, err
	}

	containerID, err := c.createContainer(ctx, img)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		_ = c.cli.ContainerRemove(context.Background(), containerID, container.RemoveOptions{Force: true})
	}()

	if err := c.copyCodeToContainer(ctx, containerID, filename, code); err != nil {
		return nil, 0, err
	}

	if err := c.cli.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return nil, 0, fmt.Errorf("failed to start container: %w", err)
	}

	// Compile step (if needed).
	if compileCmd != "" {
		compileErr, err := c.compile(ctx, containerID, compileCmd)
		if err != nil {
			return nil, 0, err // Infrastructure failure
		}
		if compileErr != "" {
			// Return compile error for all test cases
			results := make([]ExecuteResult, len(inputs))
			for i := range results {
				results[i] = ExecuteResult{Error: compileErr}
			}
			return results, 0, nil
		}
	}

	results, err := c.runAll(ctx, containerID, runCmd, inputs)

	var memKB int64
	memRes, memErr := c.execInContainer(ctx, containerID, "cat /sys/fs/cgroup/memory.peak 2>/dev/null || cat /sys/fs/cgroup/memory/memory.max_usage_in_bytes 2>/dev/null", "", 5*time.Second)
	if memErr == nil && memRes != nil && memRes.Output != "" {
		var bytes int64
		if _, err := fmt.Sscanf(memRes.Output, "%d", &bytes); err == nil {
			memKB = bytes / 1024
		}
	}

	return results, memKB, err
}

// compile runs the compile command. Returns the compile error output (if any)
// and a Go error only on infrastructure failures.
func (c *Client) compile(ctx context.Context, containerID, compileCmd string) (compileErr string, err error) {
	result, execErr := c.execInContainer(ctx, containerID, compileCmd, "", 15*time.Second)
	if execErr != nil {
		return "", fmt.Errorf("compilation failed: %w", execErr)
	}
	return result.Error, nil
}

// runAll executes the run command for each input concurrently with bounded
// parallelism. Respects context cancellation for early exit.
func (c *Client) runAll(ctx context.Context, containerID, runCmd string, inputs []string) ([]ExecuteResult, error) {
	results := make([]ExecuteResult, len(inputs))
	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	for i, input := range inputs {
		// Check if context was cancelled (early exit by caller).
		select {
		case <-ctx.Done():
			results[i] = ExecuteResult{Error: "cancelled"}
			continue
		default:
		}

		// Acquire semaphore slot (bounded concurrency).
		sem <- struct{}{}
		wg.Add(1)

		go func(idx int, inp string) {
			defer wg.Done()
			defer func() { <-sem }()

			// Recheck context inside goroutine.
			select {
			case <-ctx.Done():
				results[idx] = ExecuteResult{Error: "cancelled"}
				return
			default:
			}

			res, err := c.execInContainer(ctx, containerID, runCmd, inp, runTimeout)
			if err != nil {
				results[idx] = ExecuteResult{Error: err.Error()}
			} else {
				results[idx] = *res
			}
		}(i, input)
	}

	wg.Wait()
	return results, nil
}
