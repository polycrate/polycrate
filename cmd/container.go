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
	ctx := context.Background()
	cli, err := getDockerCLI()

	if err != nil {
		return "", err
	}

	buildOpts := types.ImageBuildOptions{
		Dockerfile: dockerfilePath,
		Tags:       tags,
	}
	buildCtx, _ := archive.TarWithOptions(workspace.LocalPath, &archive.TarOptions{})

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
	return nil
}

func runContainer(cli *client.Client, cc *container.Config, hc *container.HostConfig, name string) error {
	ctx := context.Background()
	inout := make(chan []byte)

	log.WithFields(log.Fields{
		"name": name,
	}).Debugf("Creating container")
	resp, err := cli.ContainerCreate(ctx, cc, hc, nil, nil, name)
	if err != nil {
		return err
	}

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
		log.WithFields(log.Fields{
			"id":   resp.ID,
			"name": name,
		}).Debugf("Attaching non-interactively")
		waiter, err = cli.ContainerAttach(ctx, resp.ID, containerAttachOptions)
	} else {
		log.WithFields(log.Fields{
			"id":   resp.ID,
			"name": name,
		}).Debugf("Attaching interactively (-it) - input will be forwarded")
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

	log.WithFields(log.Fields{
		"id":   resp.ID,
		"name": name,
	}).Debugf("Starting container")
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case <-statusCh:
	}

	log.WithFields(log.Fields{
		"id":   resp.ID,
		"name": name,
	}).Debugf("Removing container")
	// Stop and remove the container
	if err := pruneContainer(cli, resp.ID); err != nil {
		return err
	}
	return nil
}
