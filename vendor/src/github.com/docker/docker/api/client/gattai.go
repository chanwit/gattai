package client

import (
	"github.com/chanwit/gattai/client"
)

//
// Repository
//

func (cli *DockerCli) CmdInit(args ...string) error {
	return client.DoInit(cli, args...)
}

//
// Provision
//

func (cli *DockerCli) CmdActive(args ...string) error {
	return client.DoActive(cli, args...)
}

func (cli *DockerCli) CmdLs(args ...string) error {
	return client.DoLs(cli, args...)
}

func (cli *DockerCli) CmdRmm(args ...string) error {
	return client.DoRmm(cli, args...)
}

//
// Composition
//

func (cli *DockerCli) CmdUp(args ...string) error {
	return client.DoUp(cli, args...)
}
