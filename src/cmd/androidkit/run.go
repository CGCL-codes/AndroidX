package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	log "github.com/sirupsen/logrus"
)

func runUsage() {
	invoked := filepath.Base(os.Args[0])
	fmt.Printf("USAGE: %s run [backend] [options] [prefix]\n\n", invoked)
	fmt.Printf("'backend' specifies the run backend.\n")
	fmt.Printf("If not specified the platform specific default will be used\n")
	fmt.Printf("Supported backends are only qemu\n")
	fmt.Printf("  qemu [linux]\n")
	fmt.Printf("\n")
	fmt.Printf("'options' are the backend specific options.\n")
	fmt.Printf("See '%s run [backend] --help' for details.\n\n", invoked)
	fmt.Printf("'prefix' specifies the path to the VM image.\n")
	fmt.Printf("It defaults to './image'.\n")
}

func run(args []string) {
	if len(args) < 1 {
		runUsage()
		os.Exit(1)
	}
	switch args[0] {
	case "qemu":
		runQemu(args[1:])
	default:
		switch runtime.GOOS {
		case "linux":
			runQemu(args)
		default:
			log.Errorf("There currently is no default 'run' backend for your platform.")
		}
	}
}
