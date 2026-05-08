package isolation

import (
	"time"

	"github.com/docker/docker/client"
)

// Resource limits for sandbox containers.
const (
	memoryLimit = 256 * 1024 * 1024 // 256MB
	cpuLimit    = 500000000         // 0.5 CPU (in NanoCPUs)
	runTimeout  = 5 * time.Second   // Per-test-case execution timeout
)

// maxConcurrency caps parallel test case executions inside a single container.
// We set this to a small fixed number (e.g., 2) so that we don't overwhelm
// the container's 256MB memory and PID limits when evaluating large batches.
var maxConcurrency = 3

// ExecuteResult holds the stdout and stderr output from a single code execution.
type ExecuteResult struct {
	Output string
	Error  string
	TimeMs int64
}

// Client wraps the Docker API client for creating and managing sandbox containers.
type Client struct {
	cli *client.Client
}

// NewClient creates a new Docker client. Panics if Docker is unavailable since
// the entire application depends on it.
func NewClient() *Client {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	return &Client{cli: cli}
}
