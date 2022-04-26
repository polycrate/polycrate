package cmd

import (
	"context"
	"io"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

// type ContainerConfig struct {
// 	Name    string
// 	Mounts  map[string]string
// 	Image   string
// 	Command []string
// 	Env     *map[string]string
// 	Flags   map[string]string // rm: "true"
// }

func doContainerStuff() {
	containerImage := strings.Join([]string{workspace.Config.Image.Reference, workspace.Config.Image.Version}, ":")
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	if err != nil {
		panic(err)
	}

	reader, err := cli.ImagePull(ctx, containerImage, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}

	defer reader.Close()
	io.Copy(os.Stdout, reader)

	cc := &container.Config{
		Image: containerImage,
		Cmd:   []string{"echo", "hello world"},
		Tty:   false,
	}

	// Setup mounts
	containerMounts := []mount.Mount{}
	for containerMount := range workspace.mounts {
		m := mount.Mount{
			Type:   mount.TypeBind,
			Source: containerMount,
			Target: workspace.mounts[containerMount],
		}
		containerMounts = append(containerMounts, m)
	}

	hc := &container.HostConfig{
		Mounts: containerMounts,
	}

	resp, err := cli.ContainerCreate(ctx, cc, hc, nil, nil, "")
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
}
