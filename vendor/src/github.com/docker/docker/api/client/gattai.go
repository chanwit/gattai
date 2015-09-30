package client

import (
	"github.com/chanwit/gattai/client"
)

func (cli *DockerCli) CmdActive(args ...string) error {
	return client.DoActive(cli, args...)
}

func (cli *DockerCli) CmdInit(args ...string) error {
	return client.DoInit(cli, args...)
}

func (cli *DockerCli) CmdLs(args ...string) error {
	return client.DoLs(cli, args...)
}

func (cli *DockerCli) CmdUp(args ...string) error {
	return client.DoUp(cli, args...)
}
