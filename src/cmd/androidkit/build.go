package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/CGCL-codes/AndroidX/src/cmd/androidkit/imglib"
	log "github.com/sirupsen/logrus"
)

// currently we support only one format: kernel+initrd
type formatList []string

func (f *formatList) String() string {
	return fmt.Sprint(*f)
}
func (f *formatList) Set(value string) error {
	// allow comma separated options or multiple options
	for _, cs := range strings.Split(value, ",") {
		*f = append(*f, cs)
	}
	return nil
}

// Process the build arguments and execute build
func build(args []string) {
	var buildFormats formatList
	/* only support one format: kernel+initrd */
	outputTypes := imglib.OutputTypes()

	buildCmd := flag.NewFlagSet("build", flag.ExitOnError)
	buildCmd.Usage = func() {
		fmt.Printf("USAGE: %s build [options] <file>[.yml]\n\n", os.Args[0])
		fmt.Printf("Options:\n")
		buildCmd.PrintDefaults()
	}
	buildPull := buildCmd.Bool("pull", false, "Always pull images")
	buildDisableTrust := buildCmd.Bool("disable-content-trust", false, "Skip image trust verification specified in trust section of config (default false)")
	/* kernel+initrd */
	buildCmd.Var(&buildFormats, "format", "Formats to create [ "+strings.Join(outputTypes, " ")+" ]")
	if err := buildCmd.Parse(args); err != nil {
		log.Fatal("Unable to parse args")
	}
	remArgs := buildCmd.Args()
	if len(remArgs) != 1 {
		fmt.Println("Please specify a configuration file")
		buildCmd.Usage()
		os.Exit(1)
	}

	conf := remArgs[0]
	name := strings.TrimSuffix(filepath.Base(conf), filepath.Ext(conf))

	// check build format, now we only support kernel+initrd
	if len(buildFormats) == 0 {
		buildFormats = formatList{"kernel+initrd"}
	} else {
		err := imglib.ValidateFormats(buildFormats)
		if err != nil {
			log.Errorf("Error parsing formats: %v", err)
			buildCmd.Usage()
			os.Exit(1)
		}
	}

	// read config file and parse it
	//image trust verification
	var config []byte
	var err error
	if config, err = ioutil.ReadFile(conf); err != nil {
		log.Fatalf("Cannot open config file: %v", err)
	}

	var x imglib.AndroidX
	if x, err = imglib.NewConfig(config); err != nil {
		log.Fatalf("Invalid config: %v", err)
	}

	if *buildDisableTrust {
		log.Debugf("Disabling content trust checks for this build")
		x.Trust = imglib.TrustConfig{}
	}
	// create tempfile
	var tf *os.File
	var w io.Writer
	if tf, err = ioutil.TempFile("", ""); err != nil {
		log.Fatalf("Error creating tempfile: %v", err)
	}
	defer os.Remove(tf.Name())
	w = tf

	// build images
	if err := Build(x, w, *buildPull); err != nil {
		log.Fatalf("%v", err)
	}
	// create outputs
	image := tf.Name()
	if err := tf.Close(); err != nil {
		log.Fatalf("Error closing tempfile: %v", err)
	}

	log.Infof("Create outputs:")
	if err := imglib.Formats(name, image, buildFormats); err != nil {
		log.Fatalf("Error writing outputs: %v", err)
	}
}
