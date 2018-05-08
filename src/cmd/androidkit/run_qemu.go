package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	defaultFWPath = "/usr/share/ovmf/bios.bin"
)

// QemuConfig contains the config for Qemu
type QemuConfig struct {
	Path           string
	Kernel         bool
	GUI            bool
	StatePath      string
	FWPath         string
	Arch           string
	CPUs           string
	Memory         string
	KVM            bool
	QemuBinPath    string
	PublishedPorts []string
	NetdevConfig   string
	UUID           uuid.UUID
}

const (
	qemuNetworkingNone    string = "none"
	qemuNetworkingUser           = "user"
	qemuNetworkingTap            = "tap"
	qemuNetworkingBridge         = "bridge"
	qemuNetworkingDefault        = qemuNetworkingUser
)

var (
	defaultArch string
)

func init() {
	switch runtime.GOARCH {
	case "arm64":
		defaultArch = "aarch64"
	case "amd64":
		defaultArch = "x86_64"
	}
}

func haveKVM() bool {
	_, err := os.Stat("/dev/kvm")
	return !os.IsNotExist(err)
}

func envOverrideBool(env string, b *bool) {
	val := os.Getenv(env)
	if val == "" {
		return
	}

	var err error
	*b, err = strconv.ParseBool(val)
	if err != nil {
		log.Fatal("Unable to parse %q=%q as a boolean", env, val)
	}
}

func retrieveMAC(statePath string) net.HardwareAddr {
	var mac net.HardwareAddr
	fileName := filepath.Join(statePath, "mac-addr")

	if macString, err := ioutil.ReadFile(fileName); err == nil {
		if mac, err = net.ParseMAC(string(macString)); err != nil {
			log.Fatal("failed to parse mac-addr file: %s\n", macString)
		}
	} else {
		// we did not generate a mac yet. generate one
		mac = generateMAC()
		if err = ioutil.WriteFile(fileName, []byte(mac.String()), 0640); err != nil {
			log.Fatalln("failed to write mac-addr file:", err)
		}
	}

	return mac
}

func generateMAC() net.HardwareAddr {
	mac := make([]byte, 6)
	n, err := rand.Read(mac)
	if err != nil {
		log.WithError(err).Fatal("failed to generate random mac address")
	}
	if n != 6 {
		log.WithError(err).Fatal("generated %d bytes for random mac address", n)
	}
	mac[0] &^= 0x01 // Clear multicast bit
	mac[0] |= 0x2   // Set locally administered bit
	return net.HardwareAddr(mac)
}

func runQemu(args []string) {
	invoked := filepath.Base(os.Args[0])
	flags := flag.NewFlagSet("qemu", flag.ExitOnError)
	flags.Usage = func() {
		fmt.Printf("USAGE: %s run qemu [options] path\n\n", invoked)
		fmt.Printf("'path' specifies the path to the VM image.\n")
		fmt.Printf("\n")
		fmt.Printf("Options:\n")
		flags.PrintDefaults()
		fmt.Printf("\n")
		fmt.Printf("If not running as root note that '-networking bridge,br0' requires a\n")
		fmt.Printf("setuid network helper and appropriate host configuration, see\n")
		fmt.Printf("http://wiki.qemu.org/Features/HelperNetworking.\n")
	}

	// Display flags
	enableGUI := flags.Bool("gui", false, "Set qemu to use video output instead of stdio")

	// Boot type; now we only support kernel+initrd
	kernelBoot := flags.Bool("kernel", true, "Boot image is kernel+initrd+cmdline 'path'-kernel/-initrd/-cmdline")

	// State flags
	state := flags.String("state", "", "Path to directory to keep VM state in")

	// Paths and settings for UEFI firware
	// Note, we do not use defaultFWPath here as we have a special case for containerised execution
	fw := flags.String("fw", "", "Path to OVMF firmware for UEFI boot")

	// VM configuration
	enableKVM := flags.Bool("kvm", haveKVM(), "Enable KVM acceleration")
	arch := flags.String("arch", defaultArch, "Type of architecture to use, e.g. x86_64, aarch64")
	cpus := flags.String("cpus", "1", "Number of CPUs")
	mem := flags.String("mem", "1024", "Amount of memory in MB")

	// Generate UUID, so that /sys/class/dmi/id/product_uuid is populated
	vmUUID := uuid.New()

	// Networking
	networking := flags.String("networking", qemuNetworkingDefault, "Networking mode. Valid options are 'default', 'user', 'bridge[,name]', tap[,name] and 'none'. 'user' uses QEMUs userspace networking. 'bridge' connects to a preexisting bridge. 'tap' uses a prexisting tap device. 'none' disables networking.`")

	publishFlags := multipleFlag{}
	flags.Var(&publishFlags, "publish", "Publish a vm's port(s) to the host (default [])")

	if err := flags.Parse(args); err != nil {
		log.Fatal("Unable to parse args")
	}
	remArgs := flags.Args()

	// These envvars override the corresponding command line
	// options. So this must remain after the `flags.Parse` above.
	envOverrideBool("AndroidX_QEMU_KVM", enableKVM)

	if len(remArgs) == 0 {
		fmt.Println("Please specify the path to the image to boot")
		flags.Usage()
		os.Exit(1)
	}
	path := remArgs[0]
	prefix := path

	_, err := os.Stat(path)
	stat := err == nil

	// if the path does not exist, must be trying to do a kernel boot
	if !stat {
		_, err = os.Stat(path + "-kernel")
		statKernel := err == nil
		if !statKernel {
			fmt.Println("Could not find kernel image")
			os.Exit(1)
		}
	}

	if *state == "" {
		*state = prefix + "-state"
	}

	if err := os.MkdirAll(*state, 0755); err != nil {
		log.Fatalf("Could not create state directory: %v", err)
	}

	if *networking == "" || *networking == "default" {
		dflt := qemuNetworkingDefault
		networking = &dflt
	}
	netMode := strings.SplitN(*networking, ",", 2)

	var netdevConfig string
	switch netMode[0] {
	case qemuNetworkingUser:
		netdevConfig = "user,id=t0"
	case qemuNetworkingTap:
		if len(netMode) != 2 {
			log.Fatalf("Not enough arguments for %q networking mode", qemuNetworkingTap)
		}
		if len(publishFlags) != 0 {
			log.Fatalf("Port publishing requires %q networking mode", qemuNetworkingUser)
		}
		netdevConfig = fmt.Sprintf("tap,id=t0,ifname=%s,script=no,downscript=no", netMode[1])
	case qemuNetworkingBridge:
		if len(netMode) != 2 {
			log.Fatalf("Not enough arguments for %q networking mode", qemuNetworkingBridge)
		}
		if len(publishFlags) != 0 {
			log.Fatalf("Port publishing requires %q networking mode", qemuNetworkingUser)
		}
		netdevConfig = fmt.Sprintf("bridge,id=t0,br=%s", netMode[1])
	case qemuNetworkingNone:
		if len(publishFlags) != 0 {
			log.Fatalf("Port publishing requires %q networking mode", qemuNetworkingUser)
		}
		netdevConfig = ""
	default:
		log.Fatalf("Invalid networking mode: %s", netMode[0])
	}

	config := QemuConfig{
		Path:           path,
		Kernel:         *kernelBoot,
		GUI:            *enableGUI,
		StatePath:      *state,
		FWPath:         *fw,
		Arch:           *arch,
		CPUs:           *cpus,
		Memory:         *mem,
		KVM:            *enableKVM,
		PublishedPorts: publishFlags,
		NetdevConfig:   netdevConfig,
		UUID:           vmUUID,
	}

	config = discoverBackend(config)
	if err := runQemuLocal(config); err != nil {
		log.Fatal(err.Error())

	}
}

func runQemuLocal(config QemuConfig) error {
	var args []string
	config, args = buildQemuCmdline(config)

	qemuCmd := exec.Command(config.QemuBinPath, args...)
	// If verbosity is enabled print out the full path/arguments
	log.Debugf("%v\n", qemuCmd.Args)

	// If we're not using a separate window then link the execution to stdin/out
	if config.GUI != true {
		qemuCmd.Stdin = os.Stdin
		qemuCmd.Stdout = os.Stdout
		qemuCmd.Stderr = os.Stderr
	}

	return qemuCmd.Run()
}

func buildQemuCmdline(config QemuConfig) (QemuConfig, []string) {
	// Iterate through the flags and build arguments
	var qemuArgs []string
	qemuArgs = append(qemuArgs, "-smp", config.CPUs)
	qemuArgs = append(qemuArgs, "-m", config.Memory)
	qemuArgs = append(qemuArgs, "-uuid", config.UUID.String())
	qemuArgs = append(qemuArgs, "-pidfile", filepath.Join(config.StatePath, "qemu.pid"))
	// Need to specify the vcpu type when running qemu on arm64 platform, for security reason,
	// the vcpu should be "host" instead of other names such as "cortex-a53"...
	if config.Arch == "aarch64" {
		if runtime.GOARCH == "arm64" {
			qemuArgs = append(qemuArgs, "-cpu", "host")
		} else {
			qemuArgs = append(qemuArgs, "-cpu", "cortex-a57")
		}
	}

	if config.KVM {
		qemuArgs = append(qemuArgs, "-enable-kvm")
		if config.Arch == "aarch64" {
			qemuArgs = append(qemuArgs, "-machine", "virt,gic_version=host")
		} else {
			qemuArgs = append(qemuArgs, "-machine", "q35,accel=kvm:tcg")
		}
	} else {
		if config.Arch == "aarch64" {
			qemuArgs = append(qemuArgs, "-machine", "virt")
		} else {
			qemuArgs = append(qemuArgs, "-machine", "q35")
		}
	}

	rng := "rng-random,id=rng0"
	if runtime.GOOS == "linux" {
		rng = rng + ",filename=/dev/urandom"
	}
	qemuArgs = append(qemuArgs, "-object", rng, "-device", "virtio-rng-pci,rng=rng0")

	// build kernel boot config from kernel/initrd/cmdline
	if config.Kernel {
		qemuKernelPath := config.Path + "-kernel"
		qemuInitrdPath := config.Path + "-initrd.img"
		qemuArgs = append(qemuArgs, "-kernel", qemuKernelPath)
		qemuArgs = append(qemuArgs, "-initrd", qemuInitrdPath)
		cmdlineString, err := ioutil.ReadFile(config.Path + "-cmdline")
		if err != nil {
			log.Errorf("Cannot open cmdline file: %v", err)
		} else {
			qemuArgs = append(qemuArgs, "-append", string(cmdlineString))
		}
	}

	if config.NetdevConfig == "" {
		qemuArgs = append(qemuArgs, "-net", "none")
	} else {
		mac := retrieveMAC(config.StatePath)
		qemuArgs = append(qemuArgs, "-device", "virtio-net-pci,netdev=t0,mac="+mac.String())
		forwardings, err := buildQemuForwardings(config.PublishedPorts)
		if err != nil {
			log.Error(err)
		}
		qemuArgs = append(qemuArgs, "-netdev", config.NetdevConfig+forwardings)
	}

	if config.GUI != true {
		qemuArgs = append(qemuArgs, "-nographic")
	}

	return config, qemuArgs
}

func discoverBackend(config QemuConfig) QemuConfig {
	qemuBinPath := "qemu-system-" + config.Arch

	var err error
	config.QemuBinPath, err = exec.LookPath(qemuBinPath)
	if err != nil {
		fmt.Printf("Unable to find %s within the $PATH\n", qemuBinPath)
		os.Exit(1)
	}
	return config
}

func buildQemuForwardings(publishFlags multipleFlag) (string, error) {
	if len(publishFlags) == 0 {
		return "", nil
	}
	var forwardings string
	for _, publish := range publishFlags {
		p, err := NewPublishedPort(publish)
		if err != nil {
			return "", err
		}
		hostPort := p.Host
		guestPort := p.Guest
		forwardings = fmt.Sprintf("%s,hostfwd=%s::%d-:%d", forwardings, p.Protocol, hostPort, guestPort)
	}

	return forwardings, nil
}

func buildDockerForwardings(publishedPorts []string) ([]string, error) {
	pmap := []string{}
	for _, port := range publishedPorts {
		s, err := NewPublishedPort(port)
		if err != nil {
			return nil, err
		}
		pmap = append(pmap, "-p", fmt.Sprintf("%d:%d/%s", s.Host, s.Guest, s.Protocol))
	}
	return pmap, nil
}
