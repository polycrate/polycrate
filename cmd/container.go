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
	buildCtx, _ := archive.TarWithOptions(workspace.Path, &archive.TarOptions{})

	resp, err := cli.ImageBuild(ctx, buildCtx, buildOpts)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	defer resp.Body.Close()

	buildLogger := log.New()
	buildLogger.SetLevel(logrusLevel)

	w := buildLogger.WriterLevel(log.DebugLevel)
	defer w.Close()

	termFd, isTerm := term.GetFdInfo(w)

	jsonmessage.DisplayJSONMessagesStream(resp.Body, w, termFd, isTerm, nil)

	log.Debug("Built " + tags[0])
	return tags[0], nil
}

func runContainer(cli *client.Client, cc *container.Config, hc *container.HostConfig) error {
	ctx := context.Background()
	inout := make(chan []byte)

	resp, err := cli.ContainerCreate(ctx, cc, hc, nil, nil, "")
	if err != nil {
		return err
	}
	log.Debugf("Created container with ID %s", resp.ID)

	if !interactive {
		// statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
		// select {
		// case err := <-errCh:
		// 	if err != nil {
		// 		return err
		// 	}
		// case status := <-statusCh:
		// 	log.Debugf("Container exited with status code %d", status.StatusCode)
		// }
		// out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
		// //reader, err = cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
		// if err != nil {
		// 	return err
		// }
		// stdcopy.StdCopy(os.Stdout, os.Stderr, out)
		log.Debugf("Attaching in non-interactive mode")
		waiter, err := cli.ContainerAttach(ctx, resp.ID, types.ContainerAttachOptions{
			Stderr: true,
			Stdout: true,
			Stdin:  false,
			Stream: true,
		})

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
				//log.Println("Received to send to docker", string(data))
				if !ok {
					fmt.Println("!ok")
					w.Close()
					return
				}

				w.Write(append(data, '\n'))
			}
		}(waiter.Conn)

		// statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
		// select {
		// case err := <-errCh:
		// 	if err != nil {
		// 		return err
		// 	}
		// case <-statusCh:
		// }
	} else {
		log.Debugf("Attaching in interactive mode")
		log.Warnf("Container in interactive mode - input will be forwarded")
		waiter, err := cli.ContainerAttach(ctx, resp.ID, types.ContainerAttachOptions{
			Stderr: true,
			Stdout: true,
			Stdin:  true,
			Stream: true,
		})

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
				//log.Println("Received to send to docker", string(data))
				if !ok {
					fmt.Println("!ok")
					w.Close()
					return
				}

				w.Write(append(data, '\n'))
			}
		}(waiter.Conn)

	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}
	log.Debugf("Started container with ID %s", resp.ID)

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case <-statusCh:
	}

	if err := cli.ContainerStop(ctx, resp.ID, nil); err != nil {
		return fmt.Errorf("unable to stop container %s: %s", resp.ID, err)
	}
	log.Debugf("Stopped container with ID %s", resp.ID)

	removeOptions := types.ContainerRemoveOptions{
		RemoveVolumes: false,
		Force:         true,
	}

	if err := cli.ContainerRemove(ctx, resp.ID, removeOptions); err != nil {
		return fmt.Errorf("unable to remove container: %s", err)
	}
	log.Debugf("Removed container with ID %s", resp.ID)

	//stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	return nil
}
