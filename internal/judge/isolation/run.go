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
func (c *Client) Run(ctx context.Context, compilerImg, runtimeImg, filename, code, compileCmd, runCmd string, outputsToCopy []string, input string, memLimitBytes int64, timeLimitMS int64) (*ExecuteResult, int64, error) {
	results, memKB, err := c.RunBatch(ctx, compilerImg, runtimeImg, filename, code, compileCmd, runCmd, outputsToCopy, []string{input}, memLimitBytes, timeLimitMS)
	if err != nil {
		return nil, 0, err
	}
	return &results[0], memKB, nil
}

// RunBatch creates a compilation container (if needed), compiles the code,
// extracts the output binary/classes, and executes it inside a clean runtime container.
//
// Container lifecycle:
//  1. Pull compiler/runtime image (if needed)
//  2. If compiled language, create and start compilation container
//  3. Copy source code into compilation container
//  4. Compile source code
//  5. Extract compiled artifacts from compilation container
//  6. Destroy compilation container
//  7. Create and start runtime container
//  8. Copy compiled artifacts into runtime container (or source code if interpreted)
//  9. Run the program for each input concurrently (bounded by maxConcurrency)
//  10. Measure peak memory via cgroups of runtime container
//  11. Cleanup: force remove runtime container
//
// The ctx parameter supports cancellation — when cancelled, remaining test cases
// are skipped and marked as "cancelled".
func (c *Client) RunBatch(ctx context.Context, compilerImg, runtimeImg, filename, code, compileCmd, runCmd string, outputsToCopy []string, inputs []string, memLimitBytes int64, timeLimitMS int64) ([]ExecuteResult, int64, error) {
	if compilerImg != "" {
		if err := c.ensureImage(ctx, compilerImg); err != nil {
			return nil, 0, err
		}
	}
	if err := c.ensureImage(ctx, runtimeImg); err != nil {
		return nil, 0, err
	}

	var compContainerID string
	var err error

	// If separate compiler is specified, compile first in compiler container
	if compilerImg != "" && compileCmd != "" {
		compContainerID, err = c.createContainer(ctx, compilerImg, memLimitBytes)
		if err != nil {
			return nil, 0, err
		}
		defer func() {
			_ = c.cli.ContainerRemove(context.Background(), compContainerID, container.RemoveOptions{Force: true})
		}()

		if err := c.copyCodeToContainer(ctx, compContainerID, filename, code); err != nil {
			return nil, 0, err
		}

		if err := c.cli.ContainerStart(ctx, compContainerID, container.StartOptions{}); err != nil {
			return nil, 0, fmt.Errorf("failed to start compiler container: %w", err)
		}

		compileErr, err := c.compile(ctx, compContainerID, compileCmd)
		if err != nil {
			return nil, 0, err
		}
		if compileErr != "" {
			results := make([]ExecuteResult, len(inputs))
			for i := range results {
				results[i] = ExecuteResult{Error: compileErr}
			}
			return results, 0, nil
		}
	}

	// Create and start runtime container
	runContainerID, err := c.createContainer(ctx, runtimeImg, memLimitBytes)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		_ = c.cli.ContainerRemove(context.Background(), runContainerID, container.RemoveOptions{Force: true})
	}()

	// If compiled in separate container, copy compiled artifacts to runtime container.
	// Otherwise, copy the source code directly to the runtime container.
	if compilerImg != "" && compileCmd != "" && len(outputsToCopy) > 0 {
		for _, file := range outputsToCopy {
			tarStream, err := c.copyFromContainer(ctx, compContainerID, "/tmp/"+file)
			if err != nil {
				return nil, 0, fmt.Errorf("failed to extract compiled file %s: %w", file, err)
			}
			err = c.copyArchiveToContainer(ctx, runContainerID, "/tmp", tarStream)
			_ = tarStream.Close() // Ensure stream is closed
			if err != nil {
				return nil, 0, fmt.Errorf("failed to copy compiled file %s to runtime container: %w", file, err)
			}
		}
	} else {
		// Non-compiled or single-container fallback
		if err := c.copyCodeToContainer(ctx, runContainerID, filename, code); err != nil {
			return nil, 0, err
		}
		// If compileCmd is provided but no separate compiler image, compile inside runtime container
		if compileCmd != "" {
			if err := c.cli.ContainerStart(ctx, runContainerID, container.StartOptions{}); err != nil {
				return nil, 0, fmt.Errorf("failed to start container: %w", err)
			}
			compileErr, err := c.compile(ctx, runContainerID, compileCmd)
			if err != nil {
				return nil, 0, err
			}
			if compileErr != "" {
				results := make([]ExecuteResult, len(inputs))
				for i := range results {
					results[i] = ExecuteResult{Error: compileErr}
				}
				return results, 0, nil
			}
		}
	}

	// Ensure runtime container is started (if it wasn't already started during compilation fallback)
	inspect, err := c.cli.ContainerInspect(ctx, runContainerID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to inspect runtime container: %w", err)
	}
	if !inspect.State.Running {
		if err := c.cli.ContainerStart(ctx, runContainerID, container.StartOptions{}); err != nil {
			return nil, 0, fmt.Errorf("failed to start runtime container: %w", err)
		}
	}

	results, err := c.runAll(ctx, runContainerID, runCmd, inputs, timeLimitMS)

	// Measure peak memory using cgroups
	var memKB int64
	memRes, memErr := c.execInContainer(ctx, runContainerID, "cat /sys/fs/cgroup/memory.peak 2>/dev/null || cat /sys/fs/cgroup/memory/memory.max_usage_in_bytes 2>/dev/null", "", 5*time.Second)
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
func (c *Client) runAll(ctx context.Context, containerID, runCmd string, inputs []string, timeLimitMS int64) ([]ExecuteResult, error) {
	results := make([]ExecuteResult, len(inputs))
	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	timeout := runTimeout
	if timeLimitMS > 0 {
		timeout = time.Duration(timeLimitMS) * time.Millisecond
	}

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

			res, err := c.execInContainer(ctx, containerID, runCmd, inp, timeout)
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
