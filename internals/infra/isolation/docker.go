package isolation

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	"github.com/docker/docker/pkg/stdcopy"
)

type ExecuteResult struct {
	Output string
	Error  string
}

type Client struct {
	cli *client.Client
}

func NewClient() *Client {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err) // Initialization panic if Docker is not available
	}
	return &Client{cli: cli}
}

func (c *Client) Run(ctx context.Context, img string, filename string, code string, cmdStr string, input string) (*ExecuteResult, error) {
	// Auto-pull image if it doesn't exist
	_, _, err := c.cli.ImageInspectWithRaw(ctx, img)
	if err != nil {
		if errdefs.IsNotFound(err) {
			reader, err := c.cli.ImagePull(ctx, img, image.PullOptions{})
			if err != nil {
				return nil, fmt.Errorf("failed to pull image: %w", err)
			}
			_, _ = io.Copy(io.Discard, reader)
			reader.Close()
		} else {
			return nil, fmt.Errorf("failed to inspect image: %w", err)
		}
	}
	config := &container.Config{
		Image:           img,
		Cmd:             []string{"sh", "-c", cmdStr},
		WorkingDir:      "/app",
		OpenStdin:       true,
		StdinOnce:       true,
		AttachStdin:     true,
		AttachStdout:    true,
		AttachStderr:    true,
		NetworkDisabled: true, // Secure: No network access
	}

	hostConfig := &container.HostConfig{
		Resources: container.Resources{
			Memory:   256 * 1024 * 1024, // 256MB Limit
			NanoCPUs: 500000000,         // 0.5 CPU Limit
		},
		AutoRemove: true,
	}

	resp, err := c.cli.ContainerCreate(ctx, config, hostConfig, nil, nil, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	// Create an in-memory tar archive for the code file
	var tarBuf bytes.Buffer
	tw := tar.NewWriter(&tarBuf)
	hdr := &tar.Header{
		Name: filename,
		Mode: 0644,
		Size: int64(len(code)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return nil, fmt.Errorf("failed to write tar header: %w", err)
	}
	if _, err := tw.Write([]byte(code)); err != nil {
		return nil, fmt.Errorf("failed to write code to tar: %w", err)
	}
	if err := tw.Close(); err != nil {
		return nil, fmt.Errorf("failed to close tar writer: %w", err)
	}

	// Copy the tar archive into the container's /app directory
	err = c.cli.CopyToContainer(ctx, resp.ID, "/app", &tarBuf, container.CopyToContainerOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to copy code to container: %w", err)
	}

	attachOpts := container.AttachOptions{
		Stream: true,
		Stdin:  true,
		Stdout: true,
		Stderr: true,
	}
	hjResponse, err := c.cli.ContainerAttach(ctx, resp.ID, attachOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to attach: %w", err)
	}
	defer hjResponse.Close()

	if input != "" {
		_, _ = hjResponse.Conn.Write([]byte(input))
	}
	// Close stdin so the process knows input is done
	hjResponse.CloseWrite()

	if err := c.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	var outBuf, errBuf bytes.Buffer
	done := make(chan error, 1)

	go func() {
		_, err = stdcopy.StdCopy(&outBuf, &errBuf, hjResponse.Reader)
		done <- err
	}()

	waitCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	statusCh, errCh := c.cli.ContainerWait(waitCtx, resp.ID, container.WaitConditionNotRunning)

	select {
	case err := <-errCh:
		if err != nil {
			// If context timed out, kill the container immediately.
			if waitCtx.Err() != nil {
				_ = c.cli.ContainerKill(context.Background(), resp.ID, "SIGKILL")
				return nil, fmt.Errorf("timeout")
			}
			return nil, err
		}
	case <-statusCh:
		// Container finished naturally
	}

	<-done

	return &ExecuteResult{
		Output: outBuf.String(),
		Error:  errBuf.String(),
	}, nil
}
