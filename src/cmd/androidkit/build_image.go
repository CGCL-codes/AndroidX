package main

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/CGCL-codes/AndroidX/src/cmd/androidkit/imglib"
	log "github.com/sirupsen/logrus"
)

// imageFilter is a tar.Writer that transforms a kernel image and an initrd image into the output we want on underlying tar writer
type imageFilter struct {
	tw          *tar.Writer
	buffer      *bytes.Buffer
	discard     bool
	foundKernel bool
	foundKTar   bool
	foundInitrd bool
	//filter kernel and cmdline images
	cmdline string
	kernel  string
	//filter initrd image
	initrd string
}

// Build performs the actual build process
func Build(x imglib.AndroidX, w io.Writer, pull bool) error {
	/* create ~/.AndroidX as cache directory */
	if imglib.AndroidXDir == "" {
		imglib.AndroidXDir = imglib.DefaultAndroidXConfigDir()
	}
	// create tmp dir in case needed
	if err := os.MkdirAll(filepath.Join(imglib.AndroidXDir, "tmp"), 0755); err != nil {
		return err
	}

	iw := tar.NewWriter(w)
	var kf *imageFilter
	if x.Kernel.Ref != nil {
		// get kernel and kernel-modules images from docker container
		// and then output them to a tar stream
		log.Infof("Extract kernel image: %s", x.Kernel.Ref)
		kf = newImageFilter(iw, x.Kernel.Cmdline)
		err := imglib.ImageToTar(x.Kernel.Ref, kf, enforceContentTrust(x.Kernel.Ref.String(), &x.Trust), pull)
		if err != nil {
			return fmt.Errorf("Failed to extract kernel image and tarball: %v", err)
		}
		err = kf.Close()
		if err != nil {
			return fmt.Errorf("Close error: %v", err)
		}
	}

	// get init images from docker container and then output then to a tar stream
	if len(x.Init) != 0 {
		log.Infof("Add init containers:")
	}
	for _, ii := range x.InitRefs {
		log.Infof("Process init image: %s", ii)
		err := imglib.ImageToTar(ii, kf, enforceContentTrust(ii.String(), &x.Trust), pull)
		if err != nil {
			return fmt.Errorf("Failed to build init tarball from %s: %v", ii, err)
		}
	}
	if err := iw.Close(); err != nil {
		return fmt.Errorf("initrd close error: %v", err)
	}
	return nil
}

func enforceContentTrust(fullImageName string, config *imglib.TrustConfig) bool {
	//check image name
	for _, img := range config.Image {
		// First check for an exact name match
		if img == fullImageName {
			return true
		}
		// Also check for an image name only match
		// by removing a possible tag (with possibly added digest):
		imgAndTag := strings.Split(fullImageName, ":")
		if len(imgAndTag) >= 2 && img == imgAndTag[0] {
			return true
		}
		// and by removing a possible digest:
		imgAndDigest := strings.Split(fullImageName, "@sha256:")
		if len(imgAndDigest) >= 2 && img == imgAndDigest[0] {
			return true
		}
	}
	//check org name
	for _, org := range config.Org {
		var imgOrg string
		splitName := strings.Split(fullImageName, "/")
		switch len(splitName) {
		case 0:
			// if the image is empty, return false
			return false
		case 1:
			// for single names like nginx, use library
			imgOrg = "library"
		case 2:
			// for names that assume docker hub, like linxukit/alpine, take the first split
			imgOrg = splitName[0]
		default:
			// for names that include the registry, the second piece is the org, ex: docker.io/library/alpine
			imgOrg = splitName[1]
		}
		if imgOrg == org {
			return true
		}
	}
	return false
}

func newImageFilter(tw *tar.Writer, cmdline string) *imageFilter {
	return &imageFilter{tw: tw, cmdline: cmdline, kernel: "kernel", initrd: "initrd.img"}
}

func (i *imageFilter) finishTar() error {
	if i.buffer == nil {
		return nil
	}
	tr := tar.NewReader(i.buffer)
	err := tarAppend(i.tw, tr)
	i.buffer = nil
	return err
}

func tarAppend(iw *tar.Writer, tr *tar.Reader) error {
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		err = iw.WriteHeader(hdr)
		if err != nil {
			return err
		}
		_, err = io.Copy(iw, tr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *imageFilter) Close() error {
	if !i.foundKernel {
		return errors.New("did not find kernel in kernel image")
	}
	return i.finishTar()
}

func (i *imageFilter) Flush() error {
	err := i.finishTar()
	if err != nil {
		return err
	}
	return i.tw.Flush()
}

func (i *imageFilter) Write(b []byte) (n int, err error) {
	if i.discard {
		return len(b), nil
	}
	if i.buffer != nil {
		return i.buffer.Write(b)
	}
	return i.tw.Write(b)
}

func (i *imageFilter) WriteHeader(hdr *tar.Header) error {
	err := i.finishTar()
	if err != nil {
		return err
	}
	tw := i.tw
	switch hdr.Name {
	case i.kernel:
		if i.foundKernel {
			return errors.New("found more than one possible kernel image")
		}
		i.foundKernel = true
		i.discard = false
		// If we handled the initrd, /boot already exist.
		if !i.foundInitrd {
			whdr := &tar.Header{
				Name:     "boot",
				Mode:     0755,
				Typeflag: tar.TypeDir,
			}
			if err := tw.WriteHeader(whdr); err != nil {
				return err
			}
		}
		// add the cmdline in /boot/cmdline
		whdr := &tar.Header{
			Name: "boot/cmdline",
			Mode: 0644,
			Size: int64(len(i.cmdline)),
		}
		if err := tw.WriteHeader(whdr); err != nil {
			return err
		}
		buf := bytes.NewBufferString(i.cmdline)
		_, err = io.Copy(tw, buf)
		if err != nil {
			return err
		}
		whdr = &tar.Header{
			Name: "boot/kernel",
			Mode: hdr.Mode,
			Size: hdr.Size,
		}
		if err := tw.WriteHeader(whdr); err != nil {
			return err
		}
	//case k.tar:
	//k.foundKTar = true
	//k.discard = false
	//k.buffer = new(bytes.Buffer)
	case i.initrd:
		log.Infof("found initrd")
		i.foundInitrd = true
		i.discard = false
		// If we handled the kernel, /boot already exist.
		if !i.foundKernel {
			whdr := &tar.Header{
				Name:     "boot",
				Mode:     0755,
				Typeflag: tar.TypeDir,
			}
			if err := tw.WriteHeader(whdr); err != nil {
				return err
			}
		}
		whdr := &tar.Header{
			Name: "boot/initrd.img",
			Mode: hdr.Mode,
			Size: hdr.Size,
		}
		if err := tw.WriteHeader(whdr); err != nil {
			return err
		}
	default:
		i.discard = true
	}

	return nil
}
