package imglib

import (
	"archive/tar"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"

	"github.com/CGCL-codes/AndroidX/src/cmd/androidkit/initrd"
	log "github.com/sirupsen/logrus"
)

var outFuns = map[string]func(string, io.Reader) error{
	"kernel+initrd": func(base string, image io.Reader) error {
		kernel, initrd, cmdline, err := tarToKernelInitrd(image)
		if err != nil {
			return fmt.Errorf("Error filtering kernel and initrd: %v", err)
		}
		err = outputKernelInitrd(base, kernel, initrd, cmdline)
		if err != nil {
			return fmt.Errorf("Error writing kernel+initrd output: %v", err)
		}
		return nil
	},
}

// OutputTypes returns a list of the valid output types
func OutputTypes() []string {
	ts := []string{}
	for k := range outFuns {
		ts = append(ts, k)
	}
	sort.Strings(ts)
	return ts
}

// ValidateFormats checks if the format type is known
func ValidateFormats(formats []string) error {
	log.Debugf("validating output: %v", formats)

	for _, o := range formats {
		f := outFuns[o]
		if f == nil {
			return fmt.Errorf("Unknown format type %s", o)
		}
	}
	return nil
}

// Formats generates all the specified output formats
func Formats(base string, image string, formats []string) error {
	log.Debugf("format: %v %s", formats, base)

	err := ValidateFormats(formats)
	if err != nil {
		return err
	}
	for _, o := range formats {
		ir, err := os.Open(image)
		if err != nil {
			return err
		}
		defer ir.Close()
		f := outFuns[o]
		if err := f(base, ir); err != nil {
			return err
		}
	}

	return nil
}

func tarToKernelInitrd(r io.Reader) ([]byte, []byte, string, error) {
	tr := tar.NewReader(r)
	kernel, initrd, cmdline, err := initrd.SplitTar(tr)
	if err != nil {
		return []byte{}, []byte{}, "", err
	}
	return kernel, initrd, cmdline, nil
}

func outputKernelInitrd(base string, kernel []byte, initrd []byte, cmdline string) error {
	log.Debugf("output kernel and initrd: %s %s", base, cmdline)

	log.Infof("  %s %s %s", base+"-kernel", base+"-initrd.img", base+"-cmdline")
	if err := ioutil.WriteFile(base+"-initrd.img", initrd, os.FileMode(0644)); err != nil {
		return err
	}
	if err := ioutil.WriteFile(base+"-kernel", kernel, os.FileMode(0644)); err != nil {
		return err
	}
	return ioutil.WriteFile(base+"-cmdline", []byte(cmdline), os.FileMode(0644))
}
