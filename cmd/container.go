package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/moby/term"
	log "github.com/sirupsen/logrus"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
)

func getDockerCLI() (*client.Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	return cli, err
}

func buildContainerImage(dockerfilePath string, tags []string) (string, error) {
	log.Warnf("Building custom image %s", tags[0])
	ctx := context.Background()
	cli, err := getDockerCLI()

	if err != nil {
		return "", err
	}

	log.Debugf("Assembling docker context")
	buildOpts := types.ImageBuildOptions{
		Dockerfile: dockerfilePath,
		Tags:       tags,
	}
	buildCtx, _ := archive.TarWithOptions(workspace.Path, &archive.TarOptions{})

	log.Debugf("Building image")
	resp, err := cli.ImageBuild(ctx, buildCtx, buildOpts)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	defer resp.Body.Close()

	// We're using a special logger here that can be used as an io.Writer
	// As such it can be used by the docker log reader to dump logs directly
	// to our logger
	err = logDocker(resp.Body)
	if err != nil {
		return "", err
	}

	log.Debug("Built " + tags[0])
	return tags[0], nil
}

func pullContainerImage(image string) error {
	ctx := context.Background()
	cli, err := getDockerCLI()
	if err != nil {
		return err
	}

	reader, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()

	err = logDocker(reader)
	if err != nil {
		return err
	}

	//io.Copy(os.Stdout, reader)
	log.Debugf("Successfully pulled image %s", image)
	return nil
}

func logDocker(reader io.ReadCloser) error {
	dockerLogger := log.New()

	// Set the current loglevel
	dockerLogger.SetLevel(logrusLevel)

	// The log write writes with loglevel debug
	w := dockerLogger.WriterLevel(log.DebugLevel)
	defer w.Close()

	termFd, isTerm := term.GetFdInfo(w)

	err := jsonmessage.DisplayJSONMessagesStream(reader, w, termFd, isTerm, nil)
	if err != nil {
		return err
	}
	return nil
}

func pruneContainer(cli *client.Client, id string) error {
	ctx := context.Background()

	removeOptions := types.ContainerRemoveOptions{
		RemoveVolumes: false,
		Force:         true, // we're force-removing so we do not need to worry about timeouts
	}

	if err := cli.ContainerRemove(ctx, id, removeOptions); err != nil {
		return fmt.Errorf("unable to remove container: %s", err)
	}
	log.Debugf("Removed container with id %s", id)
	return nil
}

func runContainer(cli *client.Client, cc *container.Config, hc *container.HostConfig) error {
	ctx := context.Background()
	inout := make(chan []byte)

	resp, err := cli.ContainerCreate(ctx, cc, hc, nil, nil, "")
	if err != nil {
		return err
	}
	log.Debugf("Created container with ID %s", resp.ID)

	// Save container id to the workspace
	workspace.containerID = resp.ID

	containerAttachOptions := types.ContainerAttachOptions{
		Stderr: true,
		Stdout: true,
		Stdin:  false,
		Stream: true,
	}

	var waiter types.HijackedResponse

	if !interactive {
		log.Debugf("Attaching in non-interactive mode")
		waiter, err = cli.ContainerAttach(ctx, resp.ID, containerAttachOptions)
	} else {
		log.Warnf("Container in interactive mode - input will be forwarded")
		containerAttachOptions.Stdin = true
		waiter, err = cli.ContainerAttach(ctx, resp.ID, containerAttachOptions)
	}
	go io.Copy(os.Stdout, waiter.Reader)
	go io.Copy(os.Stderr, waiter.Reader)

	if err != nil {
		return err
	}

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			inout <- []byte(scanner.Text())
		}
	}()

	// Write to docker container
	go func(w io.WriteCloser) {
		for {
			data, ok := <-inout
			if !ok {
				fmt.Println("!ok")
				w.Close()
				return
			}

			w.Write(append(data, '\n'))
		}
	}(waiter.Conn)

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}
	log.Debugf("Started container with id %s", resp.ID)

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case <-statusCh:
	}

	// Stop and remove the container
	if err := pruneContainer(cli, resp.ID); err != nil {
		return err
	}
	return nil
}
