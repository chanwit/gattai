package client

import (
	"fmt"
)

func (cli *DockerCli) CmdInit(args ...string) error {
	// init gattai workflow
	// create .gattai/
	// ignore if already init
	fmt.Println("init command")
	return nil
}