package cmd

import (
	"context"
	"io"
	"os"

	"github.com/moby/term"
	log "github.com/sirupsen/logrus"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
)

func getDockerCLI() (*client.Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	return cli, err
}

func buildContainerImage(ctx context.Context, path string, dockerfilePath string, tags []string) (string, error) {
	cli, err := getDockerCLI()

	if err != nil {
		return "", err
	}

	buildOpts := types.ImageBuildOptions{
		Dockerfile: dockerfilePath,
		Tags:       tags,
	}
	buildCtx, _ := archive.TarWithOptions(path, &archive.TarOptions{})

	resp, err := cli.ImageBuild(ctx, buildCtx, buildOpts)
	if err != nil {
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

func PullImageGo(ctx context.Context, image string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}
	reader, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()
	termFd, isTerm := term.GetFdInfo(os.Stderr)
	jsonmessage.DisplayJSONMessagesStream(reader, os.Stderr, termFd, isTerm, nil)
	// bar := progressbar.DefaultBytes(
	// 	-1,
	// 	"pulling image",
	// )
	//io.Copy(io.MultiWriter(os.Stdout, bar), reader)

	return nil
}

func PullImage(tx *PolycrateTransaction, image string) (int, string, error) {
	// Prepare container command
	var runCmd []string

	runCmd = append(runCmd, []string{"pull", image}...)

	// Pull image
	exitCode, output, err := RunCommandWithOutput(tx.Context, nil, "docker", runCmd...)

	return exitCode, output, err
}

// func pullContainerImage(image string) error {
// 	ctx := context.Background()
// 	cli, err := getDockerCLI()
// 	if err != nil {
// 		return err
// 	}

// 	reader, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
// 	if err != nil {
// 		return err
// 	}
// 	defer reader.Close()

// 	err = logDocker(reader)
// 	if err != nil {
// 		return err
// 	}

// 	//io.Copy(os.Stdout, reader)
// 	log.Debugf("Successfully pulled image %s", image)
// 	return nil
// }

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

// func getContainers(cli *client.Client, filterList map[string]string) ([]types.Container, error) {
// 	ctx := context.Background()

// 	f := filters.NewArgs()
// 	for key, value := range filterList {
// 		f.Add(key, value)
// 	}

// 	containerListOptions := types.ContainerListOptions{
// 		Size:    false,
// 		All:     false,
// 		Since:   "container",
// 		Limit:   1,
// 		Filters: f,
// 	}

// 	containers, err := cli.ContainerList(ctx, containerListOptions)
// 	if err != nil {
// 		return nil, fmt.Errorf("error listing container: %s", err)
// 	}

// 	return containers, nil
// }

// func pruneContainer(cli *client.Client, id string) error {
// 	ctx := context.Background()

// 	log.WithFields(log.Fields{
// 		"id": id,
// 	}).Debugf("Pruning container")
// 	// filters := filters.NewArgs()
// 	// filters.Add("id", id)

// 	// containerListOptions := types.ContainerListOptions{
// 	// 	Size:    false,
// 	// 	All:     false,
// 	// 	Since:   "container",
// 	// 	Limit:   1,
// 	// 	Filters: filters,
// 	// }

// 	// containers, err := cli.ContainerList(ctx, containerListOptions)
// 	// if err != nil {
// 	// 	return fmt.Errorf("error listing container: %s", err)
// 	// }

// 	removeOptions := types.ContainerRemoveOptions{
// 		RemoveVolumes: false,
// 		Force:         true, // we're force-removing so we do not need to worry about timeouts
// 	}

// 	if err := cli.ContainerRemove(ctx, id, removeOptions); err != nil {
// 		return fmt.Errorf("unable to remove container: %s", err)
// 	}
// 	workspace.containerID = ""
// 	return nil
// }

func PruneContainer(tx *PolycrateTransaction, filters []string) (int, string, error) {

	// Prepare container command
	var runCmd []string

	// https://stackoverflow.com/questions/16248241/concatenate-two-slices-in-go
	runCmd = append(runCmd, []string{"container", "prune", "--force"}...)

	// Filter
	for _, filter := range filters {
		runCmd = append(runCmd, []string{"--filter", filter}...)
	}

	// Prune container
	exitCode, output, err := RunCommandWithOutput(tx.Context, nil, "docker", runCmd...)

	return exitCode, output, err
}

func RunContainer(tx *PolycrateTransaction, image string, command []string, env []string, mounts []string, workdir string, ports []string, labels []string, name string) (int, string, error) {
	// Prepare container command
	var runCmd []string

	// https://stackoverflow.com/questions/16248241/concatenate-two-slices-in-go
	runCmd = append(runCmd, []string{"run", "--rm"}...)

	// Name
	runCmd = append(runCmd, []string{"--name", name}...)

	// Env
	for _, envVar := range env {
		runCmd = append(runCmd, []string{"-e", envVar}...)
	}

	// Mounts
	for _, bindMount := range mounts {
		runCmd = append(runCmd, []string{"-v", bindMount}...)
	}

	// Ports
	for _, port := range ports {
		runCmd = append(runCmd, []string{"-p", port}...)
	}

	// Labels
	for _, label := range labels {
		runCmd = append(runCmd, []string{"-l", label}...)
	}

	// Workdir
	if workdir != "" {
		runCmd = append(runCmd, []string{"--workdir", workdir}...)
	}

	// Interactive
	if interactive {
		tx.Log.Info("Running in interactive mode")
		runCmd = append(runCmd, []string{"-it"}...)
	} else {
		runCmd = append(runCmd, []string{"-t"}...)
	}

	// Entrypoint
	entrypointCmd := []string{"--entrypoint", "/bin/bash"}
	runCmd = append(runCmd, entrypointCmd...)

	// Image
	runCmd = append(runCmd, image)

	// Command
	runCmd = append(runCmd, command...)

	// Run container
	exitCode, output, err := RunCommand(tx.Context, env, "docker", runCmd...)

	return exitCode, output, err
}

func CreateContainer(tx *PolycrateTransaction, name string, image string, command []string, env []string, mounts []string, workdir string, ports []string, labels []string) (int, string, error) {
	// Prepare container command
	var runCmd []string

	// https://stackoverflow.com/questions/16248241/concatenate-two-slices-in-go
	runCmd = append(runCmd, []string{"container", "create"}...)

	runCmd = append(runCmd, []string{"--name", name}...)

	// Env
	for _, envVar := range env {
		runCmd = append(runCmd, []string{"-e", envVar}...)
	}

	// Mounts
	for _, bindMount := range mounts {
		runCmd = append(runCmd, []string{"-v", bindMount}...)
	}

	// Ports
	for _, port := range ports {
		runCmd = append(runCmd, []string{"-p", port}...)
	}

	// Labels
	for _, label := range labels {
		runCmd = append(runCmd, []string{"-l", label}...)
	}

	// Workdir
	if workdir != "" {

		runCmd = append(runCmd, []string{"--workdir", workdir}...)
	}

	// Entrypoint
	entrypointCmd := []string{"--entrypoint", "/bin/bash"}
	runCmd = append(runCmd, entrypointCmd...)

	// Image
	runCmd = append(runCmd, image)

	if len(command) > 0 {
		runCmd = append(runCmd, command...)
	}

	// Run container
	exitCode, _, err := RunCommandWithOutput(tx.Context, nil, "docker", runCmd...)

	return exitCode, name, err
}

// func CopyFromContainer(tx *PolycrateTransaction, container string, src string, dst string) error {
// 	log := tx.Log.log
// 	log = log.WithField("container", container)
// 	log = log.WithField("src", src)
// 	log = log.WithField("dst", dst)
// 	log.Debugf("Copying file from container")

// 	// Prepare container command
// 	var runCmd []string

// 	_src := strings.Join([]string{container, src}, ":")
// 	runCmd = append(runCmd, []string{"docker", "cp", _src, dst}...)

// 	// Copy
// 	_, _, err := RunCommandWithOutput(tx.Context, nil, "sudo", runCmd...)

// 	return err
// }

func RemoveContainer(tx *PolycrateTransaction, container string) error {
	tx.Log.Debugf("Removing container '%s'", container)
	// Prepare container command
	var runCmd []string

	runCmd = append(runCmd, []string{"rm", "--force", container}...)

	// Remove container
	_, _, err := RunCommandWithOutput(tx.Context, nil, "docker", runCmd...)

	return err
}

// func runContainer(cli *client.Client, cc *container.Config, hc *container.HostConfig, name string) error {
// 	//ctx := context.Background()
// 	ctx := context.Background()
// 	//inout := make(chan []byte, 1)
// 	quit := make(chan bool, 1)

// 	log.WithFields(log.Fields{
// 		"name": name,
// 	}).Debugf("Creating container")
// 	resp, err := cli.ContainerCreate(ctx, cc, hc, nil, nil, name)
// 	if err != nil {
// 		return err
// 	}

// 	// Save container id to the workspace
// 	workspace.containerID = resp.ID

// 	containerAttachOptions := types.ContainerAttachOptions{
// 		Stderr: true,
// 		Stdout: true,
// 		Stdin:  false,
// 		Stream: true,
// 	}

// 	var waiter types.HijackedResponse

// 	if !interactive {
// 		log.WithFields(log.Fields{
// 			"id":   resp.ID,
// 			"name": name,
// 		}).Debugf("Attaching non-interactively")
// 		waiter, err = cli.ContainerAttach(ctx, resp.ID, containerAttachOptions)
// 	} else {
// 		log.WithFields(log.Fields{
// 			"id":   resp.ID,
// 			"name": name,
// 		}).Debugf("Attaching interactively (-it) - input will be forwarded")
// 		containerAttachOptions.Stdin = true
// 		waiter, err = cli.ContainerAttach(ctx, resp.ID, containerAttachOptions)
// 		go io.Copy(waiter.Conn, os.Stdin)
// 	}
// 	go io.Copy(os.Stdout, waiter.Reader)
// 	go io.Copy(os.Stderr, waiter.Reader)

// 	if err != nil {
// 		return err
// 	}

// 	defer waiter.Close()

// 	// go func() {
// 	// 	scanner := bufio.NewScanner(os.Stdin)
// 	// 	for {
// 	// 		fmt.Println("inout")
// 	// 		select {
// 	// 		case <-quit:
// 	// 			fmt.Println("inout quit")

// 	// 			return
// 	// 		default:
// 	// 			fmt.Println("inout scan")
// 	// 			for scanner.Scan() {
// 	// 				inout <- []byte(scanner.Text())

// 	// 			}
// 	// 			return
// 	// 		}
// 	// 	}

// 	// }()

// 	// // Write to docker container
// 	// go func(w io.WriteCloser) {
// 	// 	for {
// 	// 		fmt.Println("writer")
// 	// 		select {
// 	// 		case <-quit:
// 	// 			fmt.Println("writer quit")
// 	// 			w.Close()
// 	// 			return
// 	// 		default:
// 	// 			data, ok := <-inout
// 	// 			if !ok {
// 	// 				fmt.Println("!ok")
// 	// 				w.Close()
// 	// 				return
// 	// 			}
// 	// 			w.Write(append(data, '\n'))
// 	// 		}

// 	// 	}
// 	// }(waiter.Conn)

// 	log.WithFields(log.Fields{
// 		"id":   resp.ID,
// 		"name": name,
// 	}).Debugf("Starting container")
// 	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
// 		return err
// 	}

// 	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
// 	select {
// 	case err := <-errCh:
// 		if err != nil {
// 			return err
// 		}
// 	case <-statusCh:

// 	}

// 	log.WithFields(log.Fields{
// 		"id":   resp.ID,
// 		"name": name,
// 	}).Debugf("Removing container")

// 	workspace.containerID = ""

// 	//os.Stdin.Close()
// 	quit <- true
// 	waiter.Close()

// 	// Stop and remove the container
// 	if err := pruneContainer(cli, resp.ID); err != nil {
// 		return err
// 	}
// 	//close(inout)
// 	return nil
// }
