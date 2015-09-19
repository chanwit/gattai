package main

import (
	"github.com/docker/docker/autogen/dockerversion"
	_ "github.com/docker/machine/drivers"
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
)

var (
	daemonUsage            = ""
	handleGlobalDaemonFlag = func() {}
)

var (
	backTabs          = "\b\b\b\b\b\b\b\b\b\b\b\b"
	separator         = command{"", ""}
	provisionCommands = []command{
		separator,
		{"", backTabs + "Provision:"},
		{"ls", "List machines"},
		{"provision", "Provision a set of machines"},
		{"rmm", "Remove machines"},
		{"service", "Manage Docker service"},
		{"ssh", "Run an SSH command on a set of machines"},

		separator,
		{"", backTabs + "Clustering:"},
		{"disti", "Distribute images across the cluster"},
		{"refresh", "Refresh a snapshot of the cluster information"},
		{"select", "Select a candidate engine to place a container"},

		separator,
		{"", backTabs + "Composition:"},
		{"scale", "Scale services or pods"},
		{"up", "Build and start services"},

		separator,
		{"", backTabs + "Engine:"},
	}
)

func main() {
	dockerversion.VERSION = "0.1"
	dockerversion.GITCOMMIT = "HEAD"

	dockerCommands = append(provisionCommands, dockerCommands...)

	dockerMain()
}
