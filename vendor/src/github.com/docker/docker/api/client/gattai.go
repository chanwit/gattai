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

func (cli *DockerCli) CmdProvision(args ...string) error {
	return client.DoProvision(cli, args...)
}
func (cli *DockerCli) CmdP(args ...string) error {
	return client.DoProvision(cli, args...)
}

func (cli *DockerCli) CmdSsh(args ...string) error {
	return client.DoSsh(cli, args...)
}

//
// Clustering
//

func (cli *DockerCli) CmdToken(args ...string) error {
	return client.DoToken(cli, args...)
}

func (cli *DockerCli) CmdCluster(args ...string) error {
	return client.DoCluster(cli, args...)
}

func (cli *DockerCli) CmdHtop(args ...string) error {
	return client.DoHtop(cli, args...)
}

//
// Composition
//

func (cli *DockerCli) CmdUp(args ...string) error {
	return client.DoUp(cli, args...)
}
