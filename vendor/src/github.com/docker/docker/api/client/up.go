package client

import (
	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/project"
)

// Usage: gattai up
func (cli *DockerCli) CmdUp(args ...string) error {
	project, err := docker.NewProject(&docker.Context{
		Context: project.Context{
			ComposeFile: "docker-compose.yml",
			ProjectName: "yeah-compose",
		},
	})

	if err != nil {
		return err
	}

	project.Up()
	return nil
}
