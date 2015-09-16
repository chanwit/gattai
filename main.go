package main

import (
	_ "github.com/docker/libcompose"
	_ "github.com/docker/machine/libmachine"
)

var daemonUsage = ""
var handleGlobalDaemonFlag = false

func main() {
	dockerClientMain()
}
