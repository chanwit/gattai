package main

import (
	"fmt"
	"strconv"

	"github.com/docker/docker/autogen/dockerversion"
	_ "github.com/docker/machine/drivers/amazonec2"
	_ "github.com/docker/machine/drivers/azure"
	_ "github.com/docker/machine/drivers/digitalocean"
	_ "github.com/docker/machine/drivers/exoscale"
	_ "github.com/docker/machine/drivers/generic"
	_ "github.com/docker/machine/drivers/google"
	_ "github.com/docker/machine/drivers/hyperv"
	_ "github.com/docker/machine/drivers/none"
	_ "github.com/docker/machine/drivers/openstack"
	_ "github.com/docker/machine/drivers/rackspace"
	_ "github.com/docker/machine/drivers/softlayer"
	_ "github.com/docker/machine/drivers/virtualbox"
	_ "github.com/docker/machine/drivers/vmwarefusion"
	_ "github.com/docker/machine/drivers/vmwarevcloudair"
	_ "github.com/docker/machine/drivers/vmwarevsphere"

	_ "github.com/docker/libcompose"
	_ "github.com/docker/machine/libmachine"
	machinelog "github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/ssh"

	"io/ioutil"
	"os"

	log "github.com/Sirupsen/logrus"
)

var (
	daemonUsage            = ""
	handleGlobalDaemonFlag = func() {}
)

var (
	backTabs       = "\b\b\b\b\b\b\b\b\b\b\b\b"
	separator      = command{"", ""}
	gattaiCommands = []command{
		separator,
		{"", backTabs + "Global:"},
		{"init", "Initialize a Gattai mission repository (.gattai)"},

		separator,
		{"", backTabs + "Provision:"},
		{"active", "Set a machine as the active Docker engine"},
		{"ls", "List machines"},
		{"provision", "Provision a set of machines (alias: p)"},
		{"rmm", "Remove machines"},
		// {"service", "Manage the Docker service on machines"},
		// {"ssh", "Run an SSH command on a set of machines"},

		separator,
		{"", backTabs + "Clustering:"},
		{"cluster", "Form the cluster with a set of machines"},
		// {"disti", "Distribute images across the cluster"},
		{"master", "Set machines to be the cluster's masters"},
		// {"refresh", "Refresh a snapshot of the cluster information"},
		// {"select", "Select a candidate engine to place a container"},
		{"token", "Manage a cluster's token on Docker Hub"},
		// {"htop", "htop"},

		// separator,
		// {"", backTabs + "Composition:"},
		// {"scale", "Scale services or pods"},
		// {"up", "Build and start services"},

		separator,
		{"", backTabs + "Engine:"},
	}
)

func readFile(file string) ([]byte, error) {
	result := []byte{}

	log.Debugf("Opening file: %s", file)

	if file == "-" {
		if bytes, err := ioutil.ReadAll(os.Stdin); err != nil {
			log.Debugf("Failed to read file from stdin: %v", err)
			return nil, err
		} else {
			result = bytes
		}
	} else if file != "" {
		if bytes, err := ioutil.ReadFile(file); os.IsNotExist(err) {
			log.Debugf("Failed to find %s", file)
			return nil, err
		} else if err != nil {
			log.Debugf("Failed to open %s", file)
			return nil, err
		} else {
			result = bytes
		}
	}

	return result, nil
}

type SimpleFormatter struct {
	log.TextFormatter
}

func (s *SimpleFormatter) Format(entry *log.Entry) ([]byte, error) {
	return []byte(entry.Message + "\n"), nil
}

func setDebugOutputLevel() {
	// TODO: I'm not really a fan of this method and really would rather
	// use -v / --verbose TBQH
	for _, f := range os.Args {
		if f == "-D" || f == "--debug" || f == "-debug" {
			machinelog.IsDebug = true
		}
	}

	debugEnv := os.Getenv("MACHINE_DEBUG")
	if debugEnv != "" {
		showDebug, err := strconv.ParseBool(debugEnv)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing boolean value from MACHINE_DEBUG: %s\n", err)
			os.Exit(1)
		}
		machinelog.IsDebug = showDebug
	}
}

func main() {

	setDebugOutputLevel()

	ssh.SetDefaultClient(ssh.Native)
	log.SetFormatter(&SimpleFormatter{})
	dockerversion.VERSION = "0.1"
	dockerversion.GITCOMMIT = "HEAD"

	if os.Getenv("MACHINE_STORAGE_PATH") == "" {
		os.Setenv("MACHINE_STORAGE_PATH", ".gattai/machine")
	}

	dockerCommands = append(gattaiCommands, dockerCommands...)

	dockerMain()
}
