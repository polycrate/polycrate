package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/moby/term"
	log "github.com/sirupsen/logrus"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
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

func getDockerCLI() (*client.Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	return cli, err
}

func buildContainerImage(dockerfilePath string, tags []string) (string, error) {
	ctx := context.Background()
	cli, err := getDockerCLI()

	if err != nil {
		return "", err
	}

	buildOpts := types.ImageBuildOptions{
		Dockerfile: dockerfilePath,
		Tags:       tags,
	}
	buildCtx, _ := archive.TarWithOptions(workspace.path, &archive.TarOptions{})

	resp, err := cli.ImageBuild(ctx, buildCtx, buildOpts)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	defer resp.Body.Close()
	termFd, isTerm := term.GetFdInfo(os.Stderr)
	jsonmessage.DisplayJSONMessagesStream(resp.Body, os.Stderr, termFd, isTerm, nil)

	log.Debug("Built " + tags[0])
	return tags[0], nil
}

func doContainerStuff() error {
	containerImage := strings.Join([]string{workspace.Config.Image.Reference, workspace.Config.Image.Version}, ":")
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	if err != nil {
		panic(err)
	}

	// Check if a Dockerfile is configured in the Workspace
	if workspace.Config.Dockerfile != "" {
		// Create the filepath
		dockerfilePath := filepath.Join(workspace.path, workspace.Config.Dockerfile)

		// Check if the file exists
		if _, err := os.Stat(dockerfilePath); !os.IsNotExist(err) {
			// We need to build and tag this
			log.Debugf("Found %s in Workspace. Building image.", workspace.Config.Dockerfile)
			tag := workspace.Metadata.Name + ":" + version
			log.Debugf("Building image for tag '%s'", tag)

			tags := []string{tag}
			containerImage, err = buildContainerImage(workspace.Config.Dockerfile, tags)
			if err != nil {
				return err
			}

			// Override image to use
			//containerImage = tag
		} else {
			if pull {
				_, err = cli.ImagePull(ctx, containerImage, types.ImagePullOptions{})
				if err != nil {
					log.Fatal(err)
				}
				log.Debugf("Pulled image %s", containerImage)
			}
		}
	}

	//defer reader.Close()
	//io.Copy(os.Stdout, reader)

	cc := &container.Config{
		Image:      containerImage,
		Cmd:        []string{"version"},
		Tty:        false,
		Env:        workspace.DumpEnv(),
		WorkingDir: workspace.containerPath,
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
		//AutoRemove: true,
	}

	resp, err := cli.ContainerCreate(ctx, cc, hc, nil, nil, "")
	if err != nil {
		log.Fatal(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		log.Fatal(err)
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			log.Fatal(err)
		}
	case status := <-statusCh:
		log.Debugf("Container exited with status code %d", status.StatusCode)
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	//reader, err = cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		log.Fatal(err)
	}

	if err := cli.ContainerStop(ctx, resp.ID, nil); err != nil {
		log.Fatalf("Unable to stop container %s: %s", resp.ID, err)
	}

	removeOptions := types.ContainerRemoveOptions{
		RemoveVolumes: false,
		Force:         true,
	}

	if err := cli.ContainerRemove(ctx, resp.ID, removeOptions); err != nil {
		log.Fatalf("Unable to remove container: %s", err)
		return err
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	return nil
}
