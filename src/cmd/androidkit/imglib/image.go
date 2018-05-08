package imglib

import (
	"archive/tar"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/containerd/containerd/reference"
	log "github.com/sirupsen/logrus"
)

type tarWriter interface {
	Close() error
	Flush() error
	Write(b []byte) (n int, err error)
	WriteHeader(hdr *tar.Header) error
}

// This uses Docker to convert a Docker image into a tarball. It would be an improvement if we
// used the containerd libraries to do this instead locally direct from a local image
// cache as it would be much simpler.

// Unfortunately there are some files that Docker always makes appear in a running image and
// export shows them. In particular we have no way for a user to specify their own resolv.conf.
// Even if we were not using docker export to get the image, users of docker build cannot override
// the resolv.conf either, as it is not writeable and bind mounted in.

var exclude = map[string]bool{
	".dockerenv":      true,
	"Dockerfile":      true,
	"dev/console":     true,
	"dev/pts":         true,
	"dev/shm":         true,
	"etc/hostname":    true,
	"etc/hosts":       true,
	"etc/mtab":        true,
	"etc/resolv.conf": true,
}

// ImageToTar takes a Docker image and outputs it to a tar stream
func ImageToTar(ref *reference.Spec, tw tarWriter, trust bool, pull bool) (e error) {
	log.Debugf("image tar: %s", ref)

	if pull || trust {
		err := dockerPull(ref, pull, trust)
		if err != nil {
			return fmt.Errorf("Could not pull image %s: %v", ref, err)
		}
	}
	container, err := dockerCreate(ref.String())
	if err != nil {
		// if the image wasn't found, pull it down.  Bail on other errors.
		if strings.Contains(err.Error(), "No such image") {
			err := dockerPull(ref, true, trust)
			if err != nil {
				return fmt.Errorf("Could not pull image %s: %v", ref, err)
			}
			container, err = dockerCreate(ref.String())
			if err != nil {
				return fmt.Errorf("Failed to docker create image %s: %v", ref, err)
			}
		} else {
			return fmt.Errorf("Failed to create docker image %s: %v", ref, err)
		}
	}
	contents, err := dockerExport(container)
	if err != nil {
		return fmt.Errorf("Failed to docker export container from container %s: %v", container, err)
	}
	defer func() {
		contents.Close()

		if err := dockerRm(container); e == nil && err != nil {
			e = fmt.Errorf("Failed to docker rm container %s: %v", container, err)
		}
	}()

	// now we need to filter out some files from the resulting tar archive
	tr := tar.NewReader(contents)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if exclude[hdr.Name] {
			log.Debugf("image tar: %s exclude %s", ref, hdr.Name)
			_, err = io.Copy(ioutil.Discard, tr)
			if err != nil {
				return err
			}
		} else {
			log.Debugf("image tar: %s %s add %s", ref, hdr.Name)
			if err := tw.WriteHeader(hdr); err != nil {
				return err
			}
			_, err = io.Copy(tw, tr)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
