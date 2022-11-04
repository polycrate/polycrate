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

func OCIImageExists(name string) bool {
	_, err := crane.Digest(name)
	return err == nil
}

func PullOCIImage(name string) (v1.Image, error) {
	img, err := crane.Pull(name)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func WrapOCIImage(path string, registryUrl string, imageName string, imageTag string, labels map[string]string) error {
	var tag name.Tag
	var latestTag name.Tag
	var err error

	log.WithFields(log.Fields{
		"registryUrl": registryUrl,
		"imageName":   imageName,
		"imageTag":    imageTag,
	}).Debugf("Preparing to push image")

	// if registryTag != "" {
	// 	tag, err = name.NewTag(strings.Join([]string{registryTag, imageTag}, ":"))
	// 	if err != nil {
	// 		return err
	// 	}

	// 	latestTag, err = name.NewTag(strings.Join([]string{registryTag, "latest"}, ":"))
	// 	if err != nil {
	// 		return err
	// 	}
	// } else {
	localImageTag := strings.Join([]string{imageName, imageTag}, ":")
	localImageTagLatest := strings.Join([]string{imageName, "latest"}, ":")
	tag, err = name.NewTag(strings.Join([]string{registryUrl, localImageTag}, "/"))
	if err != nil {
		return err
	}

	latestTag, err = name.NewTag(strings.Join([]string{registryUrl, localImageTagLatest}, "/"))
	if err != nil {
		return err
	}

	// Check if the image exists already; fail if it does
	if OCIImageExists(tag.String()) {
		return fmt.Errorf("tag %s already exists in the registry", tag.String())
	}

	//log.Debugf("Pulling base image %s", config.Registry.BaseImage)
	log.WithFields(log.Fields{
		"image": config.Registry.BaseImage,
	}).Debugf("Pulling base image")
	img, err := PullOCIImage(config.Registry.BaseImage)
	if err != nil {
		return err
	}

	//log.Debugf("Adding directory to image: %s", path)
	log.WithFields(log.Fields{
		"path": path,
	}).Debugf("Adding layer to image")
	addLayer, err := layerFromDir(path, "")
	if err != nil {
		return err
	}

	newImg, err := mutate.AppendLayers(img, addLayer)
	if err != nil {
		panic(err)
	}

	origConfig, err := newImg.ConfigFile()
	if err != nil {
		return err
	}
	origConfig = origConfig.DeepCopy()

	if labels != nil {
		// Set labels.
		if origConfig.Config.Labels == nil {
			origConfig.Config.Labels = map[string]string{}
		}

		// if err := validateKeyVals(labels); err != nil {
		// 	return err
		// }

		for k, v := range labels {
			log.WithFields(log.Fields{
				"key":   k,
				"value": v,
			}).Debugf("Adding label to image")
			origConfig.Config.Labels[k] = v
		}
	}

	newImg, err = mutate.Config(newImg, origConfig.Config)
	if err != nil {
		return err
	}

	//log.Debugf("Pushing image %s", tag.String())
	log.WithFields(log.Fields{
		"image": tag.String(),
	}).Debugf("Pushing image")
	if err := crane.Push(newImg, tag.String()); err != nil {
		return err
	}

	//log.Debugf("Pushing image %s", latestTag.String())
	log.WithFields(log.Fields{
		"image": latestTag.String(),
	}).Debugf("Pushing image")
	if err := crane.Push(newImg, latestTag.String()); err != nil {
		return err
	}

	return nil
}

func UnwrapOCIImage(path string, registryUrl string, imageName string, imageTag string) error {
	registryBase := config.Registry.Url

	if registryUrl != "" {
		registryBase = registryUrl
	}

	localImageTag := strings.Join([]string{imageName, imageTag}, ":")
	tag, err := name.NewTag(strings.Join([]string{registryBase, localImageTag}, "/"))
	if err != nil {
		return err
	}

	//log.Debugf("Pulling image %s", tag.String())
	log.WithFields(log.Fields{
		"image": tag.String(),
	}).Debugf("Pulling image")

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

	//log.Debugf("Saving temporary image to %s", f.Name())
	log.WithFields(log.Fields{
		"image": tag.String(),
		"path":  f.Name(),
	}).Debugf("Saving image")
	if err := crane.Export(img, f); err != nil {
		return err
	}

	//log.Debugf("Unpacking image to %s", path)
	log.WithFields(log.Fields{
		"image": tag.String(),
		"path":  path,
	}).Debugf("Unpacking image")
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
