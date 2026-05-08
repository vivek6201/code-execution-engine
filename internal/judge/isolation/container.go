package isolation

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/errdefs"
)

// createContainer creates a long-lived sandbox container with resource limits.
// The container runs `sleep infinity` to stay alive for multiple exec calls.
func (c *Client) createContainer(ctx context.Context, img string) (string, error) {
	config := &container.Config{
		Image:           img,
		Cmd:             []string{"sleep", "infinity"},
		WorkingDir:      "/app",
		NetworkDisabled: true,
		User:            "1000:1000",
	}
	var pidsLimit int64 = 512
	hostConfig := &container.HostConfig{
		Resources: container.Resources{
			Memory:    memoryLimit,
			NanoCPUs:  cpuLimit,
			PidsLimit: &pidsLimit,
		},
		CapDrop:     []string{"ALL"},
		SecurityOpt: []string{"no-new-privileges:true"},
	}

	resp, err := c.cli.ContainerCreate(ctx, config, hostConfig, nil, nil, "")
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}
	return resp.ID, nil
}

// ensureImage checks if the Docker image exists locally and pulls it if not.
func (c *Client) ensureImage(ctx context.Context, img string) error {
	_, _, err := c.cli.ImageInspectWithRaw(ctx, img)
	if err != nil {
		if errdefs.IsNotFound(err) {
			reader, pullErr := c.cli.ImagePull(ctx, img, image.PullOptions{})
			if pullErr != nil {
				return fmt.Errorf("failed to pull image: %w", pullErr)
			}
			_, _ = io.Copy(io.Discard, reader)
			reader.Close()
		} else {
			return fmt.Errorf("failed to inspect image: %w", err)
		}
	}
	return nil
}

// copyCodeToContainer creates an in-memory tar archive with the code file
// and copies it into the container's /app directory.
func (c *Client) copyCodeToContainer(ctx context.Context, containerID, filename, code string) error {
	var tarBuf bytes.Buffer
	tw := tar.NewWriter(&tarBuf)
	hdr := &tar.Header{
		Name: filename,
		Mode: 0644,
		Size: int64(len(code)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return fmt.Errorf("failed to write tar header: %w", err)
	}
	if _, err := tw.Write([]byte(code)); err != nil {
		return fmt.Errorf("failed to write code to tar: %w", err)
	}
	if err := tw.Close(); err != nil {
		return fmt.Errorf("failed to close tar writer: %w", err)
	}
	return c.cli.CopyToContainer(ctx, containerID, "/app", &tarBuf, container.CopyToContainerOptions{})
}
