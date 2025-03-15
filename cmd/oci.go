package cmd

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
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

func PullOCIImage(ctx context.Context, name string) (v1.Image, error) {
	img, err := crane.Pull(name)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func WrapOCIImage(ctx context.Context, path string, registryUrl string, imageName string, imageTag string, labels map[string]string) error {

	var tag name.Tag
	var latestTag name.Tag
	var err error

	log := log.WithField("registry", registryUrl)
	log = log.WithField("image", imageName)
	log = log.WithField("tag", imageTag)

	log.Debugf("Preparing to push image")

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
	log = log.WithField("base_image", polycrate.Config.Registry.BaseImage)
	log.Debugf("Pulling base image")

	img, err := PullOCIImage(ctx, polycrate.Config.Registry.BaseImage)
	if err != nil {
		return err
	}

	//log.Debugf("Adding directory to image: %s", path)
	log = log.WithField("path", path)
	log.Debugf("Adding layer to image")

	addLayer, err := layerFromDir(path, "")
	if err != nil {
		return err
	}

	newImg, err := mutate.AppendLayers(img, addLayer)
	if err != nil {
		return err
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
			log = log.WithField("key", k)
			log = log.WithField("value", v)
			log.Debugf("Adding label to image")
			origConfig.Config.Labels[k] = v
		}
	}

	newImg, err = mutate.Config(newImg, origConfig.Config)
	if err != nil {
		return err
	}

	//log.Debugf("Pushing image %s", tag.String())
	log.Debugf("Pushing image")
	if err := crane.Push(newImg, tag.String()); err != nil {
		return err
	}

	// Don't push the latest tag if the global development flag has been enabled
	if !dev {
		//log.Debugf("Pushing image %s", latestTag.String())
		log = log.WithField("tag", latestTag.String())
		log.Debugf("Pushing image")
		if err := crane.Push(newImg, latestTag.String()); err != nil {
			return err
		}
	}

	return nil
}

const whiteoutPrefix = ".wh."

func inWhiteoutDir(fileMap map[string]bool, file string) bool {
	for {
		if file == "" {
			break
		}
		dirname := filepath.Dir(file)
		if file == dirname {
			break
		}
		if val, ok := fileMap[dirname]; ok && val {
			return true
		}
		file = dirname
	}
	return false
}

func extract(img v1.Image, w io.Writer) error {
	tarWriter := tar.NewWriter(w)
	defer tarWriter.Close()

	fileMap := map[string]bool{}

	layers, err := img.Layers()
	if err != nil {
		return fmt.Errorf("retrieving image layers: %w", err)
	}

	// we iterate through the layers in reverse order because it makes handling
	// whiteout layers more efficient, since we can just keep track of the removed
	// files as we see .wh. layers and ignore those in previous layers.
	for i := len(layers) - 1; i >= 0; i-- {
		layer := layers[i]
		layerReader, err := layer.Uncompressed()
		if err != nil {
			return fmt.Errorf("reading layer contents: %w", err)
		}
		defer layerReader.Close()
		tarReader := tar.NewReader(layerReader)

		for {
			header, err := tarReader.Next()
			if errors.Is(err, io.EOF) {
				break
			}

			if err != nil {
				return fmt.Errorf("reading tar: %w", err)
			}

			// Some tools prepend everything with "./", so if we don't Clean the
			// name, we may have duplicate entries, which angers tar-split.
			header.Name = filepath.Clean(header.Name)
			// force PAX format to remove Name/Linkname length limit of 100 characters
			// required by USTAR and to not depend on internal tar package guess which
			// prefers USTAR over PAX
			// REMOVED BY FP
			//header.Format = tar.FormatPAX

			basename := filepath.Base(header.Name)
			dirname := filepath.Dir(header.Name)
			tombstone := strings.HasPrefix(basename, whiteoutPrefix)
			if tombstone {
				basename = basename[len(whiteoutPrefix):]
			}

			// check if we have seen value before
			// if we're checking a directory, don't filepath.Join names
			var name string
			if header.Typeflag == tar.TypeDir {
				name = header.Name
			} else {
				name = filepath.Join(dirname, basename)
			}

			if _, ok := fileMap[name]; ok {
				continue
			}

			// check for a whited out parent directory
			if inWhiteoutDir(fileMap, name) {
				continue
			}

			// mark file as handled. non-directory implicitly tombstones
			// any entries with a matching (or child) name
			fileMap[name] = tombstone || !(header.Typeflag == tar.TypeDir)
			if !tombstone {
				if err := tarWriter.WriteHeader(header); err != nil {
					return err
				}
				if header.Size > 0 {
					if _, err := io.CopyN(tarWriter, tarReader, header.Size); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func UnwrapOCIImage(ctx context.Context, path string, registryUrl string, imageName string, imageTag string) error {
	registryBase := polycrate.Config.Registry.Url

	if registryUrl != "" {
		registryBase = registryUrl
	}

	localImageTag := strings.Join([]string{imageName, imageTag}, ":")
	tag, err := name.NewTag(strings.Join([]string{registryBase, localImageTag}, "/"))
	if err != nil {
		return err
	}

	log := log.WithField("image", tag.String())

	log.Debugf("Pulling image")

	img, err := PullOCIImage(ctx, tag.String())
	if err != nil {
		return err
	}

	f, err := polycrate.getTempFile(ctx, strings.Join([]string{strings.Replace(imageName, "/", "-", -1), "tgz"}, "."))
	if err != nil {
		return err
	}
	//defer f.Close()
	//defer os.Remove(f.Name())

	log = log.WithField("tmp_path", f.Name())
	log.Debugf("Saving OCI image")
	fs, pw := io.Pipe()

	go func() {
		// Close the writer with any errors encountered during
		// extraction. These errors will be returned by the reader end
		// on subsequent reads. If err == nil, the reader will return
		// EOF.
		pw.CloseWithError(extract(img, pw))
	}()

	// if err := crane.Export(img, f); err != nil {
	// 	return err
	// }
	if _, err := io.Copy(f, fs); err != nil {
		return err
	}

	//log.Debugf("Unpacking image to %s", path)
	log = log.WithField("path", path)
	log.Debugf("Removing existing block")
	err = os.RemoveAll(path)
	if err != nil {
		return err
	}
	log.Debugf("Unpacking image")
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
			Mode: int64(info.Mode().Perm()),
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
	//return tarball.LayerFromOpener(&b)
}
