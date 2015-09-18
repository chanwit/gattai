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

func main() {
	dockerversion.VERSION = "0.1"
	dockerversion.GITCOMMIT = "HEAD"
	dockerMain()
}
