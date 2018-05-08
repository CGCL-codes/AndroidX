package initrd

import (
	"archive/tar"
	"io"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
)

// SplitTar splits the tar stream and gets kernel, cmdline, initrd
func SplitTar(r *tar.Reader) (kernel []byte, initrd []byte, cmdline string, err error) {
	for {
		var thdr *tar.Header
		thdr, err = r.Next()
		if err == io.EOF {
			return kernel, initrd, cmdline, nil
		}
		if err != nil {
			return
		}
		switch thdr.Name {
		case "boot/kernel":
			log.Infof("process kernel")
			kernel, err = ioutil.ReadAll(r)
			if err != nil {
				return
			}
		case "boot/cmdline":
			var buf []byte
			buf, err = ioutil.ReadAll(r)
			if err != nil {
				return
			}
			cmdline = string(buf)
		case "boot/initrd.img":
			log.Infof("process initrd.img")
			initrd, err = ioutil.ReadAll(r)
			if err != nil {
				return
			}
		case "boot":
			// skip this entry
		default:
			// skip this entry
		}
	}
}
