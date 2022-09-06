package cmd

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	log "github.com/sirupsen/logrus"
)

func PullOCIImage(name string) (v1.Image, error) {
	img, err := crane.Pull(name)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func WrapOCIImage(path string, imageName string, imageTag string) error {
	// goal:
	// 1. download nginx
	// 2. /usr/share/nginx/html <- delete this dir (new layer, appended on top of nginx)
	// 3. copy my blog there (new layer, appended on top of nginx)

	log.Debugf("Pulling base image %s", config.Registry.BaseImage)
	img, err := PullOCIImage(config.Registry.BaseImage)
	if err != nil {
		return err
	}

	log.Debugf("Adding directory to image: %s", path)
	addLayer, err := layerFromDir(path, "")
	if err != nil {
		return err
	}

	newImg, err := mutate.AppendLayers(img, addLayer)
	if err != nil {
		panic(err)
	}

	registryBase := strings.Join([]string{config.Registry.Url, config.Registry.BlockNamespace}, "/")
	localImageTag := strings.Join([]string{imageName, imageTag}, ":")
	localImageTagLatest := strings.Join([]string{imageName, "latest"}, ":")
	tag, err := name.NewTag(strings.Join([]string{registryBase, localImageTag}, "/"))
	if err != nil {
		return err
	}

	latestTag, err := name.NewTag(strings.Join([]string{registryBase, localImageTagLatest}, "/"))
	if err != nil {
		return err
	}

	log.Debugf("Pushing image %s", tag.String())
	if err := crane.Push(newImg, tag.String()); err != nil {
		return err
	}

	log.Debugf("Pushing image %s", latestTag.String())
	if err := crane.Push(newImg, latestTag.String()); err != nil {
		return err
	}

	return nil
}

func UnwrapOCIImage(path string, imageName string, imageTag string) error {
	registryBase := strings.Join([]string{config.Registry.Url, config.Registry.BlockNamespace}, "/")
	localImageTag := strings.Join([]string{imageName, imageTag}, ":")
	tag, err := name.NewTag(strings.Join([]string{registryBase, localImageTag}, "/"))
	if err != nil {
		return err
	}

	log.Debugf("Pulling image %s", tag.String())

	img, err := PullOCIImage(tag.String())
	if err != nil {
		return err
	}

	f, err := workspace.getTempFile(strings.Replace(imageName, "/", "-", -1))
	if err != nil {
		return err
	}
	defer f.Close()
	defer os.Remove(f.Name())

	log.Debugf("Saving temporary image to %s", f.Name())
	if err := crane.Export(img, f); err != nil {
		return err
	}

	log.Debugf("Unpacking image to %s", path)
	err = Untar(f.Name(), path)
	if err != nil {
		return err
	}

	return nil
}

func layerFromDir(root string, targetPath string) (v1.Layer, error) {
	var b bytes.Buffer
	tw := tar.NewWriter(&b)

	err := filepath.Walk(root, func(fp string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		rel, err := filepath.Rel(root, fp)
		if err != nil {
			return fmt.Errorf("failed to calculate relative path: %w", err)
		}

		hdr := &tar.Header{
			Name: path.Join(targetPath, filepath.ToSlash(rel)),
			Mode: int64(info.Mode()),
		}

		if !info.IsDir() {
			hdr.Size = info.Size()
		}

		if info.Mode().IsDir() {
			hdr.Typeflag = tar.TypeDir
		} else if info.Mode().IsRegular() {
			hdr.Typeflag = tar.TypeReg
		} else {
			return fmt.Errorf("not implemented archiving file type %s (%s)", info.Mode(), rel)
		}

		if err := tw.WriteHeader(hdr); err != nil {
			return fmt.Errorf("failed to write tar header: %w", err)
		}
		if !info.IsDir() {
			f, err := os.Open(fp)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tw, f); err != nil {
				return fmt.Errorf("failed to read file into the tar: %w", err)
			}
			f.Close()
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scan files: %w", err)
	}
	if err := tw.Close(); err != nil {
		return nil, fmt.Errorf("failed to finish tar: %w", err)
	}
	return tarball.LayerFromReader(&b)
}
